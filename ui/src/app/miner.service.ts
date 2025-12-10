import { Injectable, OnDestroy, signal, computed } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { of, forkJoin, Subscription, interval } from 'rxjs';
import { switchMap, catchError, map, startWith, tap } from 'rxjs/operators';

// Define interfaces
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

  // State Signals
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

  public hashrateHistory = signal<Map<string, HashratePoint[]>>(new Map());

  // Computed signals for convenience (optional, but helpful for components)
  public runningMiners = computed(() => this.state().runningMiners);
  public installedMiners = computed(() => this.state().installedMiners);
  public apiAvailable = computed(() => this.state().apiAvailable);
  public profiles = computed(() => this.state().profiles);

  private pollingSubscription: Subscription | undefined;

  constructor(private http: HttpClient) {
    // Initial check
    this.checkSystemState();
    // Start polling for system state every 5 seconds for chart updates
    this.pollingSubscription = interval(5000).subscribe(() => this.checkSystemState());
  }

  ngOnDestroy(): void {
    if (this.pollingSubscription) {
      this.pollingSubscription.unsubscribe();
    }
  }

  checkSystemState() {
    forkJoin({
      available: this.getAvailableMiners().pipe(catchError(() => of([]))),
      info: this.getSystemInfo().pipe(catchError(() => of({ installed_miners_info: [] }))),
      running: this.getRunningMiners().pipe(catchError(() => of([]))), // This endpoint contains the history
      profiles: this.getProfiles().pipe(catchError(() => of([])))
    }).pipe(
      map(({ available, info, running, profiles }) => {
        const installedMap = new Map<string, InstallationDetails>();

        (info.installed_miners_info || []).forEach((m: InstallationDetails) => {
          if (m.is_installed) {
            const type = this.getMinerType(m);
            installedMap.set(type, { ...m, type });
          }
        });

        running.forEach((miner: any) => {
          const type = miner.name.split('-')[0];
          if (!installedMap.has(type)) {
            installedMap.set(type, {
              is_installed: true,
              version: 'unknown (running)',
              path: 'unknown (running)',
              miner_binary: 'unknown (running)',
              type: type,
            } as InstallationDetails);
          }
        });

        const allInstalledMiners = Array.from(installedMap.values());

        // Populate hashrate history directly from the running miners data
        const newHistory = new Map<string, HashratePoint[]>();
        running.forEach((miner: any) => {
          if (miner.hashrateHistory) {
            newHistory.set(miner.name, miner.hashrateHistory);
          }
        });
        this.hashrateHistory.set(newHistory);

        if (allInstalledMiners.length === 0) {
          this.state.set({
            needsSetup: true,
            apiAvailable: true,
            error: null,
            systemInfo: info,
            manageableMiners: available.map(availMiner => ({ ...availMiner, is_installed: false })),
            installedMiners: [],
            runningMiners: [],
            profiles: profiles
          });
          return;
        }

        const manageableMiners = available.map(availMiner => ({
          ...availMiner,
          is_installed: installedMap.has(availMiner.name),
        }));

        this.state.set({
          needsSetup: false,
          apiAvailable: true,
          error: null,
          systemInfo: info,
          manageableMiners,
          installedMiners: allInstalledMiners,
          runningMiners: running,
          profiles: profiles
        });
      }),
      catchError(err => {
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
      })
    ).subscribe();
  }

  getAvailableMiners() {
    return this.http.get<AvailableMiner[]>(`${this.apiBaseUrl}/miners/available`);
  }

  getSystemInfo() {
    return this.http.get<any>(`${this.apiBaseUrl}/info`);
  }

  getRunningMiners() {
    return this.http.get<any[]>(`${this.apiBaseUrl}/miners`);
  }

  getMinerHashrateHistory(minerName: string) {
    return this.http.get<HashratePoint[]>(`${this.apiBaseUrl}/miners/${minerName}/hashrate-history`);
  }

  installMiner(minerType: string) {
    return this.http.post(`${this.apiBaseUrl}/miners/${minerType}/install`, {}).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  uninstallMiner(minerType: string) {
    return this.http.delete(`${this.apiBaseUrl}/miners/${minerType}/uninstall`).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  startMiner(profileId: string) {
    return this.http.post(`${this.apiBaseUrl}/profiles/${profileId}/start`, {}).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  stopMiner(minerName: string) {
    return this.http.delete(`${this.apiBaseUrl}/miners/${minerName}`).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  getProfiles() {
    return this.http.get<MiningProfile[]>(`${this.apiBaseUrl}/profiles`);
  }

  createProfile(profile: MiningProfile) {
    return this.http.post(`${this.apiBaseUrl}/profiles`, profile).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  updateProfile(profile: MiningProfile) {
    return this.http.put(`${this.apiBaseUrl}/profiles/${profile.id}`, profile).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  deleteProfile(profileId: string) {
    return this.http.delete(`${this.apiBaseUrl}/profiles/${profileId}`).pipe(
      tap(() => setTimeout(() => this.checkSystemState(), 1000))
    );
  }

  private getMinerType(miner: any): string {
    if (!miner.path) return 'unknown';
    const parts = miner.path.split('/').filter((p: string) => p);
    return parts.length > 1 ? parts[parts.length - 2] : parts[parts.length - 1] || 'unknown';
  }
}
