import { Injectable, signal, computed, OnDestroy, NgZone, inject } from '@angular/core';
import { Subject, Observable, timer, Subscription, BehaviorSubject } from 'rxjs';
import { filter, map, share, takeUntil } from 'rxjs/operators';

// --- Event Types ---
export type MiningEventType =
  | 'miner.starting'
  | 'miner.started'
  | 'miner.stopping'
  | 'miner.stopped'
  | 'miner.stats'
  | 'miner.error'
  | 'miner.connected'
  | 'profile.created'
  | 'profile.updated'
  | 'profile.deleted'
  | 'pong';

export interface MinerStatsData {
  name: string;
  hashrate: number;
  shares: number;
  rejected: number;
  uptime: number;
  algorithm?: string;
  diffCurrent?: number;
}

export interface MinerEventData {
  name: string;
  profileId?: string;
  reason?: string;
  error?: string;
  pool?: string;
}

export interface MiningEvent<T = unknown> {
  type: MiningEventType;
  timestamp: string;
  data?: T;
}

export type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'reconnecting';

@Injectable({
  providedIn: 'root'
})
export class WebSocketService implements OnDestroy {
  private ngZone = inject(NgZone);

  // WebSocket connection
  private socket: WebSocket | null = null;
  private wsUrl = 'ws://localhost:9090/api/v1/mining/ws/events';

  // Connection state
  private connectionState = signal<ConnectionState>('disconnected');
  readonly isConnected = computed(() => this.connectionState() === 'connected');
  readonly state = this.connectionState.asReadonly();

  // Event stream
  private eventsSubject = new Subject<MiningEvent>();
  private destroy$ = new Subject<void>();

  // Reconnection
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 10;
  private baseReconnectDelay = 1000; // 1 second
  private maxReconnectDelay = 30000; // 30 seconds
  private reconnectSubscription?: Subscription;
  private pingInterval?: ReturnType<typeof setInterval>;

  // Observable streams for specific event types
  readonly events$ = this.eventsSubject.asObservable().pipe(share());

  readonly minerStats$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerStatsData> => e.type === 'miner.stats'),
    map(e => e.data!)
  );

  readonly minerStarting$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerEventData> => e.type === 'miner.starting'),
    map(e => e.data!)
  );

  readonly minerStarted$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerEventData> => e.type === 'miner.started'),
    map(e => e.data!)
  );

  readonly minerStopping$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerEventData> => e.type === 'miner.stopping'),
    map(e => e.data!)
  );

  readonly minerStopped$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerEventData> => e.type === 'miner.stopped'),
    map(e => e.data!)
  );

  readonly minerError$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerEventData> => e.type === 'miner.error'),
    map(e => e.data!)
  );

  readonly minerConnected$ = this.events$.pipe(
    filter((e): e is MiningEvent<MinerEventData> => e.type === 'miner.connected'),
    map(e => e.data!)
  );

  constructor() {
    // Auto-connect on service creation
    this.connect();
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
    this.disconnect();
  }

  /**
   * Connect to the WebSocket server
   */
  connect(): void {
    if (this.socket && (this.socket.readyState === WebSocket.CONNECTING || this.socket.readyState === WebSocket.OPEN)) {
      return; // Already connected or connecting
    }

    this.connectionState.set('connecting');
    console.log('[WebSocket] Connecting to', this.wsUrl);

    try {
      this.socket = new WebSocket(this.wsUrl);

      this.socket.onopen = () => {
        this.ngZone.run(() => {
          console.log('[WebSocket] Connected');
          this.connectionState.set('connected');
          this.reconnectAttempts = 0;

          // Subscribe to all miners by default
          this.send({ type: 'subscribe', miners: ['*'] });

          // Start ping interval to keep connection alive
          this.startPingInterval();
        });
      };

      this.socket.onmessage = (event) => {
        this.ngZone.run(() => {
          try {
            const data = JSON.parse(event.data) as MiningEvent;
            this.eventsSubject.next(data);

            // Log non-stats events for debugging
            if (data.type !== 'miner.stats' && data.type !== 'pong') {
              console.log('[WebSocket] Event:', data.type, data.data);
            }
          } catch (err) {
            console.error('[WebSocket] Failed to parse message:', err);
          }
        });
      };

      this.socket.onclose = (event) => {
        this.ngZone.run(() => {
          console.log('[WebSocket] Connection closed:', event.code, event.reason);
          this.stopPingInterval();
          this.connectionState.set('disconnected');
          this.socket = null;

          // Attempt reconnection unless intentionally closed
          if (event.code !== 1000) {
            this.scheduleReconnect();
          }
        });
      };

      this.socket.onerror = (error) => {
        this.ngZone.run(() => {
          console.error('[WebSocket] Error:', error);
          // The onclose event will handle reconnection
        });
      };
    } catch (err) {
      console.error('[WebSocket] Failed to create connection:', err);
      this.connectionState.set('disconnected');
      this.scheduleReconnect();
    }
  }

  /**
   * Disconnect from the WebSocket server
   */
  disconnect(): void {
    this.cancelReconnect();
    this.stopPingInterval();

    if (this.socket) {
      this.socket.close(1000, 'Client disconnecting');
      this.socket = null;
    }

    this.connectionState.set('disconnected');
  }

  /**
   * Send a message to the server
   */
  private send(message: object): void {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(JSON.stringify(message));
    }
  }

  /**
   * Subscribe to specific miners (or '*' for all)
   */
  subscribeToMiners(miners: string[]): void {
    this.send({ type: 'subscribe', miners });
  }

  /**
   * Schedule a reconnection with exponential backoff
   */
  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.log('[WebSocket] Max reconnection attempts reached');
      return;
    }

    this.cancelReconnect();
    this.connectionState.set('reconnecting');

    // Exponential backoff with jitter
    const delay = Math.min(
      this.baseReconnectDelay * Math.pow(2, this.reconnectAttempts) + Math.random() * 1000,
      this.maxReconnectDelay
    );

    this.reconnectAttempts++;
    console.log(`[WebSocket] Reconnecting in ${Math.round(delay / 1000)}s (attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts})`);

    this.reconnectSubscription = timer(delay)
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        this.connect();
      });
  }

  /**
   * Cancel any pending reconnection
   */
  private cancelReconnect(): void {
    if (this.reconnectSubscription) {
      this.reconnectSubscription.unsubscribe();
      this.reconnectSubscription = undefined;
    }
  }

  /**
   * Start sending periodic pings to keep the connection alive
   */
  private startPingInterval(): void {
    this.stopPingInterval();
    this.pingInterval = setInterval(() => {
      this.send({ type: 'ping' });
    }, 30000); // Every 30 seconds
  }

  /**
   * Stop the ping interval
   */
  private stopPingInterval(): void {
    if (this.pingInterval) {
      clearInterval(this.pingInterval);
      this.pingInterval = undefined;
    }
  }

  /**
   * Get events for a specific miner
   */
  getMinerEvents(minerName: string): Observable<MiningEvent> {
    return this.events$.pipe(
      filter(e => {
        const data = e.data as MinerEventData | MinerStatsData | undefined;
        return data?.name === minerName;
      })
    );
  }

  /**
   * Get stats events for a specific miner
   */
  getMinerStats(minerName: string): Observable<MinerStatsData> {
    return this.minerStats$.pipe(
      filter(data => data.name === minerName)
    );
  }
}
