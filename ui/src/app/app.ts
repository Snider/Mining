import {
  Component,
  OnInit,
  ElementRef,
  ViewEncapsulation,
  CUSTOM_ELEMENTS_SCHEMA
} from '@angular/core';
import { HttpClient, HttpClientModule, HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { of, forkJoin } from 'rxjs';
import { switchMap, catchError, map } from 'rxjs/operators';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/tooltip/tooltip.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';
import '@awesome.me/webawesome/dist/components/spinner/spinner.js';
import '@awesome.me/webawesome/dist/components/input/input.js';

// Define interfaces for our data structures
interface InstallationDetails {
  is_installed: boolean;
  version: string;
  path: string;
  miner_binary: string;
  config_path?: string;
  type?: string;
}

interface AvailableMiner {
  name: string;
  description: string;
}

@Component({
  selector: 'mde-mining-dashboard',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, HttpClientModule, FormsModule],
  templateUrl: './app.html',
  styleUrls: ["app.css"],
  encapsulation: ViewEncapsulation.ShadowDom
})
export class MiningDashboardElementComponent implements OnInit {
  apiBaseUrl: string = 'http://localhost:9090/api/v1/mining';

  // State management
  needsSetup: boolean = false;
  apiAvailable: boolean = true;
  error: string | null = null;
  showAdminPanel: boolean = false;
  actionInProgress: string | null = null; // To track which miner action is running

  systemInfo: any = null;
  manageableMiners: any[] = [];
  runningMiners: any[] = [];
  installedMiners: InstallationDetails[] = [];
  whitelistPaths: string[] = [];

  // Form inputs
  poolAddress: string = 'pool.hashvault.pro:80';
  walletAddress: string = '888tNkZrPN6JsEgekjMnABU4TBzc2Dt29EPAvkRxbANsAnjyPbb3iQ1YBRk1UXcdRsiKc9dhwMVgN5S9cQUiyoogDavup3H';
  showStartOptionsFor: string | null = null;

  constructor(private http: HttpClient, private elementRef: ElementRef) {}

  ngOnInit(): void {
    this.checkSystemState();
  }

  private handleError(err: HttpErrorResponse, defaultMessage: string) {
    console.error(err);
    if (err.error && err.error.error) {
      this.error = `${defaultMessage}: ${err.error.error}`;
    } else if (typeof err.error === 'string' && err.error.length < 200) {
      this.error = `${defaultMessage}: ${err.error}`;
    } else {
      this.error = `${defaultMessage}. Please check the console for details.`;
    }
  }

  checkSystemState() {
    this.error = null;
    forkJoin({
      available: this.http.get<AvailableMiner[]>(`${this.apiBaseUrl}/miners/available`),
      info: this.http.get<any>(`${this.apiBaseUrl}/info`)
    }).pipe(
      switchMap(({ available, info }) => {
        this.apiAvailable = true;
        this.systemInfo = info;

        const trulyInstalledMiners = (info.installed_miners_info || []).filter((m: InstallationDetails) => m.is_installed);

        if (trulyInstalledMiners.length === 0) {
          this.needsSetup = true;
          this.manageableMiners = available.map(availMiner => ({ ...availMiner, is_installed: false }));
          this.installedMiners = [];
          this.runningMiners = [];
          return of(null);
        }

        this.needsSetup = false;
        const installedMap = new Map<string, InstallationDetails>(
          (info.installed_miners_info || []).map((m: InstallationDetails) => [this.getMinerType(m), m])
        );

        this.manageableMiners = available.map(availMiner => ({
          ...availMiner,
          is_installed: installedMap.get(availMiner.name)?.is_installed ?? false,
        }));

        this.installedMiners = trulyInstalledMiners.map((m: InstallationDetails) => ({ ...m, type: this.getMinerType(m) }));

        this.updateWhitelistPaths();
        return this.fetchRunningMiners();
      }),
      catchError(err => {
        if (err.status === 500) {
          this.needsSetup = true;
          this.fetchAvailableMinersForWizard();
        } else {
          this.apiAvailable = false;
          this.error = 'Failed to connect to the mining API.';
        }
        this.systemInfo = {};
        this.installedMiners = [];
        this.runningMiners = [];
        console.error('API not available or needs setup:', err);
        return of(null);
      })
    ).subscribe();
  }

  fetchAvailableMinersForWizard(): void {
    this.http.get<AvailableMiner[]>(`${this.apiBaseUrl}/miners/available`).subscribe({
      next: miners => { this.manageableMiners = miners.map(m => ({...m, is_installed: false})); },
      error: err => { this.handleError(err, 'Could not fetch available miners for setup'); }
    });
  }

  fetchRunningMiners() {
    return this.http.get<any[]>(`${this.apiBaseUrl}/miners`).pipe(
      map(miners => { this.runningMiners = miners; this.updateWhitelistPaths(); }),
      catchError(err => {
        this.handleError(err, 'Could not fetch running miners');
        this.runningMiners = [];
        return of([]);
      })
    );
  }

  private updateWhitelistPaths() {
    const paths = new Set<string>();
    this.installedMiners.forEach(miner => {
      if (miner.miner_binary) paths.add(miner.miner_binary);
      if (miner.config_path) paths.add(miner.config_path);
    });
    this.runningMiners.forEach(miner => {
      if (miner.configPath) paths.add(miner.configPath);
    });
    this.whitelistPaths = Array.from(paths);
  }

  installMiner(minerType: string): void {
    this.actionInProgress = `install-${minerType}`;
    this.error = null;
    this.http.post(`${this.apiBaseUrl}/miners/${minerType}/install`, {}).subscribe({
      next: () => {
        setTimeout(() => {
          this.checkSystemState();
          this.actionInProgress = null;
        }, 1000);
      },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to install ${minerType}`);
        this.actionInProgress = null;
      }
    });
  }

  uninstallMiner(minerType: string): void {
    this.actionInProgress = `uninstall-${minerType}`;
    this.error = null;
    this.http.delete(`${this.apiBaseUrl}/miners/${minerType}/uninstall`).subscribe({
      next: () => {
        setTimeout(() => {
          this.checkSystemState();
          this.actionInProgress = null;
        }, 1000);
      },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to uninstall ${minerType}`);
        this.actionInProgress = null;
      }
    });
  }

  startMiner(miner: any, useLastConfig: boolean = false): void {
    this.actionInProgress = `start-${miner.type}`;
    this.error = null;
    let config = {};
    if (!useLastConfig) {
      config = {
        pool: this.poolAddress,
        wallet: this.walletAddress,
        tls: true,
        hugePages: true,
      };
    }
    this.http.post(`${this.apiBaseUrl}/miners/${miner.type}`, config).subscribe({
      next: () => {
        setTimeout(() => {
          this.checkSystemState();
          this.actionInProgress = null;
        }, 1000);
      },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to start ${miner.type}`);
        this.actionInProgress = null;
      }
    });
    this.showStartOptionsFor = null;
  }

  stopMiner(miner: any): void {
    const runningInstance = this.getRunningMinerInstance(miner);
    if (!runningInstance) {
      this.error = "Cannot stop a miner that is not running.";
      return;
    }
    this.actionInProgress = `stop-${miner.type}`;
    this.error = null;
    this.http.delete(`${this.apiBaseUrl}/miners/${runningInstance.name}`).subscribe({
      next: () => {
        setTimeout(() => {
          this.checkSystemState();
          this.actionInProgress = null;
        }, 1000);
      },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to stop ${runningInstance.name}`);
        this.actionInProgress = null;
      }
    });
  }

  toggleAdminPanel(): void {
    this.showAdminPanel = !this.showAdminPanel;
  }

  toggleStartOptions(minerType: string | undefined): void {
    if (!minerType) return;
    this.showStartOptionsFor = this.showStartOptionsFor === minerType ? null : minerType;
  }

  getMinerType(miner: any): string {
    if (!miner.path) return 'unknown';
    const parts = miner.path.split('/').filter((p: string) => p);
    return parts.length > 1 ? parts[parts.length - 2] : parts[parts.length - 1] || 'unknown';
  }

  getRunningMinerInstance(miner: any): any {
    return this.runningMiners.find(m => m.name.startsWith(miner.type));
  }

  isMinerRunning(miner: any): boolean {
    return !!this.getRunningMinerInstance(miner);
  }
}
