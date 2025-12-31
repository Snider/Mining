import { Injectable, OnDestroy, signal, computed, inject } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { of, forkJoin, Subscription, interval, merge } from 'rxjs';
import { switchMap, catchError, map, tap, filter, debounceTime } from 'rxjs/operators';
import { WebSocketService, MinerEventData, MinerStatsData } from './websocket.service';

// --- Interfaces ---
export interface InstallationDetails {
  is_installed: boolean;
  version: string;
  path: string;
  miner_binary: string;
  config_path?: string;
  type?: string;
}

export interface AvailableMiner {
  name: string;
  description: string;
}

export interface HashratePoint {
  timestamp: string;
  hashrate: number;
}

export interface MiningProfile {
  id: string;
  name: string;
  minerType: string;
  config: any;
}

export interface SystemState {
  needsSetup: boolean;
  apiAvailable: boolean;
  error: string | null;
  systemInfo: any;
  manageableMiners: any[];
  installedMiners: InstallationDetails[];
  runningMiners: any[];
  profiles: MiningProfile[];
}

@Injectable({
  providedIn: 'root'
})
export class MinerService implements OnDestroy {
  private apiBaseUrl = 'http://localhost:9090/api/v1/mining';
  private pollingSubscription?: Subscription;
  private wsSubscriptions: Subscription[] = [];
  private ws = inject(WebSocketService);

  // --- State Signals ---
  public state = signal<SystemState>({
    needsSetup: false,
    apiAvailable: true,
    error: null,
    systemInfo: {},
    manageableMiners: [],
    installedMiners: [],
    runningMiners: [],
    profiles: []
  });

  // Separate signal for hashrate history as it updates frequently
  public hashrateHistory = signal<Map<string, HashratePoint[]>>(new Map());

  // Historical hashrate data from database with configurable time range
  public historicalHashrate = signal<Map<string, HashratePoint[]>>(new Map());
  public selectedTimeRange = signal<number>(60); // Default 60 minutes
  private historyPollingSubscription?: Subscription;

  // Available time ranges in minutes
  public readonly timeRanges = [
    { label: '5m', minutes: 5 },
    { label: '15m', minutes: 15 },
    { label: '30m', minutes: 30 },
    { label: '45m', minutes: 45 },
    { label: '1h', minutes: 60 },
    { label: '3h', minutes: 180 },
    { label: '6h', minutes: 360 },
    { label: '12h', minutes: 720 },
    { label: '24h', minutes: 1440 },
  ];

  // --- View Mode Signals (single/multi miner view) ---
  public viewMode = signal<'all' | 'single'>('all');
  public selectedMinerName = signal<string | null>(null);

  // --- Computed Signals for easy access in components ---
  public runningMiners = computed(() => this.state().runningMiners);
  public installedMiners = computed(() => this.state().installedMiners);
  public apiAvailable = computed(() => this.state().apiAvailable);
  public profiles = computed(() => this.state().profiles);

  // Selected miner object (when in single mode)
  public selectedMiner = computed(() => {
    const name = this.selectedMinerName();
    if (!name) return null;
    return this.runningMiners().find(m => m.name === name) || null;
  });

  // Displayed miners based on view mode
  public displayedMiners = computed(() => {
    if (this.viewMode() === 'all') {
      return this.runningMiners();
    }
    const selected = this.selectedMiner();
    return selected ? [selected] : [];
  });

  constructor(private http: HttpClient) {
    this.forceRefreshState();
    this.startPollingLive_Data();
    this.startPollingHistoricalData();
    this.subscribeToWebSocketEvents();
  }

  ngOnDestroy(): void {
    this.stopPolling();
    this.historyPollingSubscription?.unsubscribe();
    this.wsSubscriptions.forEach(sub => sub.unsubscribe());
  }

  // --- WebSocket Event Subscriptions ---

  /**
   * Subscribe to WebSocket events for real-time updates.
   * This supplements polling with instant event-driven updates.
   */
  private subscribeToWebSocketEvents(): void {
    // Listen for miner started/stopped events to refresh the miner list immediately
    const minerLifecycleEvents = merge(
      this.ws.minerStarted$,
      this.ws.minerStopped$
    ).pipe(
      debounceTime(500) // Debounce to avoid rapid-fire updates
    ).subscribe(() => {
      // Refresh running miners when a miner starts or stops
      this.getRunningMiners().pipe(
        catchError(() => of([]))
      ).subscribe(runningMiners => {
        this.state.update(s => ({ ...s, runningMiners }));
        this.updateHashrateHistory(runningMiners);
      });
    });
    this.wsSubscriptions.push(minerLifecycleEvents);

    // Listen for stats events to update hashrates in real-time
    // This provides more immediate updates than the 5-second polling interval
    const statsSubscription = this.ws.minerStats$.subscribe((stats: MinerStatsData) => {
      // Update the running miners with fresh hashrate data
      this.state.update(s => {
        const runningMiners = s.runningMiners.map(miner => {
          if (miner.name === stats.name) {
            return {
              ...miner,
              stats: {
                ...miner.stats,
                hashrate: stats.hashrate,
                shares: stats.shares,
                rejected: stats.rejected,
                uptime: stats.uptime,
                algorithm: stats.algorithm || miner.stats?.algorithm,
              }
            };
          }
          return miner;
        });
        return { ...s, runningMiners };
      });
    });
    this.wsSubscriptions.push(statsSubscription);

    // Listen for error events to show notifications
    const errorSubscription = this.ws.minerError$.subscribe((data: MinerEventData) => {
      console.error(`[MinerService] Miner error for ${data.name}:`, data.error);
      // Notification can be handled by components listening to this event
    });
    this.wsSubscriptions.push(errorSubscription);
  }

  // --- Data Loading and Polling Logic ---

  /**
   * Loads all system data. Can be called to force a full refresh.
   */
  public forceRefreshState() {
    forkJoin({
      available: this.getAvailableMiners().pipe(catchError(() => of([]))),
      info: this.getSystemInfo().pipe(catchError(() => of({ installed_miners_info: [] }))),
      running: this.getRunningMiners().pipe(catchError(() => of([]))),
      profiles: this.getProfiles().pipe(catchError(() => of([])))
    }).pipe(
      map(({ available, info, running, profiles }) => this.processSystemState(available, info, running, profiles)),
      catchError(err => this.handleApiError(err))
    ).subscribe(initialState => {
      if (initialState) {
        this.state.set(initialState);
        this.updateHashrateHistory(initialState.runningMiners);
        this.fetchHistoricalHashrate();
      }
    });
  }

  /**
   * Starts a polling interval to fetch only live data (running miners and hashrates).
   */
  private startPollingLive_Data() {
    this.pollingSubscription = interval(5000).pipe(
      switchMap(() => this.getRunningMiners().pipe(catchError(() => of([]))))
    ).subscribe(runningMiners => {
      this.state.update(s => ({ ...s, runningMiners }));
      this.updateHashrateHistory(runningMiners);
    });
  }

  /**
   * Starts a polling interval to fetch historical data from database.
   * Polls every 30 seconds. Initial fetch happens in forceRefreshState after miners are loaded.
   */
  private startPollingHistoricalData() {
    // Poll every 30 seconds (initial fetch happens in forceRefreshState)
    this.historyPollingSubscription = interval(30000).subscribe(() => {
      this.fetchHistoricalHashrate();
    });
  }

  /**
   * Fetches 24-hour historical hashrate data for all running miners from the database.
   */
  private fetchHistoricalHashrate() {
    const runningMiners = this.state().runningMiners;
    if (runningMiners.length === 0) {
      this.historicalHashrate.set(new Map());
      return;
    }

    // Fetch historical data for each running miner
    const requests = runningMiners.map(miner =>
      this.getHistoricalHashrateForMiner(miner.name).pipe(
        map(data => ({ name: miner.name, data })),
        catchError(() => of({ name: miner.name, data: [] as HashratePoint[] }))
      )
    );

    forkJoin(requests).subscribe(results => {
      const newHistory = new Map<string, HashratePoint[]>();
      results.forEach(result => {
        if (result.data && result.data.length > 0) {
          newHistory.set(result.name, result.data);
        }
      });
      this.historicalHashrate.set(newHistory);
    });
  }

  /**
   * Fetches historical hashrate for a specific miner based on selected time range.
   */
  private getHistoricalHashrateForMiner(minerName: string) {
    const minutes = this.selectedTimeRange();
    const since = new Date(Date.now() - minutes * 60 * 1000).toISOString();
    const until = new Date().toISOString();
    return this.http.get<HashratePoint[]>(
      `${this.apiBaseUrl}/history/miners/${minerName}/hashrate?since=${since}&until=${until}`
    );
  }

  /**
   * Sets the time range for historical data and refreshes immediately.
   */
  public setTimeRange(minutes: number) {
    this.selectedTimeRange.set(minutes);
    this.fetchHistoricalHashrate();
  }

  private stopPolling() {
    this.pollingSubscription?.unsubscribe();
  }

  /**
   * Refreshes only the list of profiles. Called after create, update, or delete.
   */
  private refreshProfiles() {
    this.getProfiles().pipe(catchError(() => of(this.state().profiles))).subscribe(profiles => {
      this.state.update(s => ({ ...s, profiles }));
    });
  }

  /**
   * Refreshes system information, typically after installing or uninstalling a miner.
   */
  private refreshSystemInfo() {
    forkJoin({
      available: this.getAvailableMiners().pipe(catchError(() => of([]))),
      info: this.getSystemInfo().pipe(catchError(() => of({ installed_miners_info: [] })))
    }).subscribe(({ available, info }) => {
      const { manageableMiners, installedMiners: allInstalledMiners } = this.processStaticMinerInfo(available, info);
      this.state.update(s => ({ ...s, manageableMiners, installedMiners: allInstalledMiners, systemInfo: info }));
    });
  }

  // --- Public API Methods for Components ---

  installMiner(minerType: string) {
    return this.http.post(`${this.apiBaseUrl}/miners/${minerType}/install`, {}).pipe(
      tap(() => setTimeout(() => this.refreshSystemInfo(), 1000))
    );
  }

  uninstallMiner(minerType: string) {
    return this.http.delete(`${this.apiBaseUrl}/miners/${minerType}/uninstall`).pipe(
      tap(() => setTimeout(() => this.refreshSystemInfo(), 1000))
    );
  }

  startMiner(profileId: string) {
    return this.http.post(`${this.apiBaseUrl}/profiles/${profileId}/start`, {}).pipe(
      // An immediate poll for running miners will be triggered by the interval soon enough
    );
  }

  stopMiner(minerName: string) {
    return this.http.delete(`${this.apiBaseUrl}/miners/${minerName}`).pipe(
      // An immediate poll for running miners will be triggered by the interval soon enough
    );
  }

  getMinerLogs(minerName: string) {
    return this.http.get<string[]>(`${this.apiBaseUrl}/miners/${minerName}/logs`).pipe(
      map(logs => logs.map(line => {
        try {
          // Decode base64 encoded log lines
          return atob(line);
        } catch {
          // If decoding fails, return the original line
          return line;
        }
      }))
    );
  }

  sendStdin(minerName: string, input: string) {
    return this.http.post<{status: string, input: string}>(
      `${this.apiBaseUrl}/miners/${minerName}/stdin`,
      { input }
    );
  }

  createProfile(profile: MiningProfile) {
    return this.http.post(`${this.apiBaseUrl}/profiles`, profile).pipe(
      tap(() => this.refreshProfiles())
    );
  }

  updateProfile(profile: MiningProfile) {
    return this.http.put(`${this.apiBaseUrl}/profiles/${profile.id}`, profile).pipe(
      tap(() => this.refreshProfiles())
    );
  }

  deleteProfile(profileId: string) {
    return this.http.delete(`${this.apiBaseUrl}/profiles/${profileId}`).pipe(
      tap(() => this.refreshProfiles())
    );
  }

  // --- View Mode Methods ---

  /**
   * Select a specific miner for single-miner view
   */
  selectMiner(minerName: string) {
    this.selectedMinerName.set(minerName);
    this.viewMode.set('single');
  }

  /**
   * Switch to all-miners view
   */
  selectAllMiners() {
    this.selectedMinerName.set(null);
    this.viewMode.set('all');
  }

  /**
   * Find the profile associated with a running miner
   */
  getProfileForMiner(minerName: string): MiningProfile | null {
    // Extract miner type from the name (e.g., "xmrig-123" -> "xmrig")
    const minerType = minerName.split('-')[0];
    // Find matching profile by miner type
    return this.profiles().find(p => p.minerType === minerType) || null;
  }

  // --- Private Endpoints and Helpers ---

  private getAvailableMiners = () => this.http.get<AvailableMiner[]>(`${this.apiBaseUrl}/miners/available`);
  private getSystemInfo = () => this.http.get<any>(`${this.apiBaseUrl}/info`);
  private getRunningMiners = () => this.http.get<any[]>(`${this.apiBaseUrl}/miners`);
  private getProfiles = () => this.http.get<MiningProfile[]>(`${this.apiBaseUrl}/profiles`);

  private updateHashrateHistory(runningMiners: any[]) {
    const newHistory = new Map<string, HashratePoint[]>();
    runningMiners.forEach(miner => {
      if (miner.hashrateHistory) {
        newHistory.set(miner.name, miner.hashrateHistory);
      }
    });
    this.hashrateHistory.set(newHistory);
  }

  private processStaticMinerInfo(available: AvailableMiner[], info: any) {
    const installedMap = new Map<string, InstallationDetails>();
    (info.installed_miners_info || []).forEach((m: InstallationDetails) => {
      if (m.is_installed) {
        const type = this.getMinerType(m);
        installedMap.set(type, { ...m, type });
      }
    });

    const allInstalledMiners = Array.from(installedMap.values());
    const manageableMiners = available.map(availMiner => ({
      ...availMiner,
      is_installed: installedMap.has(availMiner.name),
    }));

    return { manageableMiners, installedMiners: allInstalledMiners };
  }

  private processSystemState(available: AvailableMiner[], info: any, running: any[], profiles: MiningProfile[]): SystemState {
    const { manageableMiners, installedMiners } = this.processStaticMinerInfo(available, info);

    return {
      needsSetup: installedMiners.length === 0,
      apiAvailable: true,
      error: null,
      systemInfo: info,
      manageableMiners,
      installedMiners,
      runningMiners: running,
      profiles
    };
  }

  private handleApiError(err: any) {
    console.error('API not available or needs setup:', err);
    this.hashrateHistory.set(new Map()); // Clear history on error
    this.state.set({
      needsSetup: false,
      apiAvailable: false,
      error: 'Failed to connect to the mining API.',
      systemInfo: {},
      manageableMiners: [],
      installedMiners: [],
      runningMiners: [],
      profiles: []
    });
    return of(null);
  }

  private getMinerType(miner: any): string {
    if (!miner.path) return 'unknown';
    const parts = miner.path.split('/').filter((p: string) => p);
    return parts.length > 1 ? parts[parts.length - 2] : parts[parts.length - 1] || 'unknown';
  }
}
