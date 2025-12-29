import { Injectable, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { of, interval, Subscription } from 'rxjs';
import { switchMap, catchError, tap } from 'rxjs/operators';

// --- Node Interfaces ---
export interface NodeIdentity {
  id: string;
  name: string;
  publicKey: string;
  createdAt: string;
  role: 'controller' | 'worker' | 'dual';
}

export interface Peer {
  id: string;
  name: string;
  publicKey: string;
  address: string;
  role: 'controller' | 'worker' | 'dual';
  addedAt: string;
  lastSeen: string;
  pingMs: number;
  hops: number;
  geoKm: number;
  score: number;
  status?: 'online' | 'offline' | 'connecting';
}

export interface PeerStats {
  nodeId: string;
  nodeName: string;
  miners: MinerStats[];
}

export interface MinerStats {
  name: string;
  type: string;
  hashrate: number;
  shares: number;
  rejected: number;
  uptime: number;
  pool: string;
  algorithm: string;
}

export interface NodeInfoResponse {
  hasIdentity: boolean;
  identity?: NodeIdentity;
  registeredPeers: number;
  connectedPeers: number;
}

export interface NodeState {
  initialized: boolean;
  identity: NodeIdentity | null;
  peers: Peer[];
  remoteStats: Map<string, PeerStats>;
  error: string | null;
}

@Injectable({
  providedIn: 'root'
})
export class NodeService {
  private apiBaseUrl = 'http://localhost:9090/api/v1/mining';
  private pollingSubscription?: Subscription;

  // --- State Signal ---
  public state = signal<NodeState>({
    initialized: false,
    identity: null,
    peers: [],
    remoteStats: new Map(),
    error: null
  });

  // --- Computed Signals ---
  public identity = computed(() => this.state().identity);
  public peers = computed(() => this.state().peers);
  public initialized = computed(() => this.state().initialized);
  public onlinePeers = computed(() => this.state().peers.filter(p => p.status === 'online'));
  public totalRemoteHashrate = computed(() => {
    let total = 0;
    this.state().remoteStats.forEach(stats => {
      stats.miners.forEach(m => total += m.hashrate);
    });
    return total;
  });

  constructor(private http: HttpClient) {
    this.loadNodeInfo();
  }

  // --- Data Loading ---

  /**
   * Load node identity and peers
   */
  public loadNodeInfo() {
    this.getNodeInfo().pipe(
      catchError(() => of(null))
    ).subscribe(response => {
      if (response && response.hasIdentity && response.identity) {
        this.state.update(s => ({ ...s, initialized: true, identity: response.identity! }));
        this.loadPeers();
      } else {
        this.state.update(s => ({ ...s, initialized: false, identity: null }));
      }
    });
  }

  /**
   * Load peer list
   */
  public loadPeers() {
    this.listPeers().pipe(
      catchError(() => of([]))
    ).subscribe(peers => {
      this.state.update(s => ({ ...s, peers }));
    });
  }

  /**
   * Start polling for peer status and stats
   */
  public startPolling() {
    if (this.pollingSubscription) return;

    this.pollingSubscription = interval(10000).pipe(
      switchMap(() => this.listPeers().pipe(catchError(() => of(this.state().peers))))
    ).subscribe(peers => {
      this.state.update(s => ({ ...s, peers }));
    });
  }

  /**
   * Stop polling
   */
  public stopPolling() {
    this.pollingSubscription?.unsubscribe();
    this.pollingSubscription = undefined;
  }

  // --- Public API Methods ---

  /**
   * Initialize node with identity
   */
  initNode(name: string, role: 'controller' | 'worker' | 'dual' = 'dual') {
    return this.http.post<NodeIdentity>(`${this.apiBaseUrl}/node/init`, { name, role }).pipe(
      tap(identity => {
        this.state.update(s => ({ ...s, initialized: true, identity }));
      })
    );
  }

  /**
   * Get node identity
   */
  getNodeInfo() {
    return this.http.get<NodeInfoResponse>(`${this.apiBaseUrl}/node/info`);
  }

  /**
   * List all registered peers
   */
  listPeers() {
    return this.http.get<Peer[]>(`${this.apiBaseUrl}/peers`);
  }

  /**
   * Add a new peer
   */
  addPeer(address: string, name?: string) {
    return this.http.post<Peer>(`${this.apiBaseUrl}/peers`, { address, name }).pipe(
      tap(() => this.loadPeers())
    );
  }

  /**
   * Remove a peer
   */
  removePeer(peerId: string) {
    return this.http.delete(`${this.apiBaseUrl}/peers/${peerId}`).pipe(
      tap(() => this.loadPeers())
    );
  }

  /**
   * Ping a peer to update metrics
   */
  pingPeer(peerId: string) {
    return this.http.post<{ pingMs: number }>(`${this.apiBaseUrl}/peers/${peerId}/ping`, {}).pipe(
      tap(() => this.loadPeers())
    );
  }

  /**
   * Get stats from all remote peers
   */
  getRemoteStats() {
    return this.http.get<Record<string, PeerStats>>(`${this.apiBaseUrl}/remote/stats`).pipe(
      tap(statsObj => {
        const statsMap = new Map<string, PeerStats>();
        Object.entries(statsObj).forEach(([peerId, stats]) => {
          statsMap.set(peerId, stats);
        });
        this.state.update(s => ({ ...s, remoteStats: statsMap }));
      })
    );
  }

  /**
   * Get stats from specific peer
   */
  getPeerStats(peerId: string) {
    return this.http.get<PeerStats>(`${this.apiBaseUrl}/remote/${peerId}/stats`);
  }

  /**
   * Start miner on remote peer
   */
  startRemoteMiner(peerId: string, profileId: string) {
    return this.http.post(`${this.apiBaseUrl}/remote/${peerId}/start`, { profileId });
  }

  /**
   * Stop miner on remote peer
   */
  stopRemoteMiner(peerId: string, minerName?: string) {
    const body = minerName ? { minerName } : {};
    return this.http.post(`${this.apiBaseUrl}/remote/${peerId}/stop`, body);
  }

  /**
   * Get logs from remote miner
   */
  getRemoteLogs(peerId: string, minerName: string, lines: number = 100) {
    return this.http.get<string[]>(`${this.apiBaseUrl}/remote/${peerId}/logs/${minerName}?lines=${lines}`);
  }

  /**
   * Deploy profile to remote peer
   */
  deployProfile(peerId: string, profileId: string) {
    return this.http.post(`${this.apiBaseUrl}/remote/${peerId}/deploy`, { type: 'profile', profileId });
  }
}
