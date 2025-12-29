import { Injectable, OnDestroy, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { of, forkJoin, Subscription, interval } from 'rxjs';
import { switchMap, catchError, map, tap } from 'rxjs/operators';

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
  }

  ngOnDestroy(): void {
    this.stopPolling();
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
    return this.http.get<string[]>(`${this.apiBaseUrl}/miners/${minerName}/logs`);
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
