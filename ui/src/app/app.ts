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
import { of } from 'rxjs';
import { switchMap, catchError, map } from 'rxjs/operators';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/tooltip/tooltip.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';
import '@awesome.me/webawesome/dist/components/spinner/spinner.js';
import '@awesome.me/webawesome/dist/components/input/input.js';

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
  apiAvailable: boolean = true;
  error: string | null = null;

  systemInfo: any = null;
  availableMiners: any[] = [];
  runningMiners: any[] = [];
  installedMiners: any[] = [];

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
      // Handles { "error": "..." } from the backend
      this.error = `${defaultMessage}: ${err.error.error}`;
    } else if (typeof err.error === 'string' && err.error.length < 200) {
      // Handles plain text errors
      this.error = `${defaultMessage}: ${err.error}`;
    } else {
      this.error = `${defaultMessage}. Please check the console for details.`;
    }
  }

  checkSystemState() {
    this.error = null;
    this.http.get<any>(`${this.apiBaseUrl}/info`).pipe(
      switchMap(info => {
        this.apiAvailable = true;
        this.systemInfo = info;

        this.installedMiners = (info.installed_miners_info || [])
          .filter((m: any) => m.is_installed)
          .map((m: any) => ({ ...m, type: this.getMinerType(m) }));

        if (this.installedMiners.length === 0) {
          this.fetchAvailableMiners();
        }

        return this.fetchRunningMiners();
      }),
      catchError(err => {
        this.apiAvailable = false;
        this.error = 'Failed to connect to the mining API.';
        this.systemInfo = {};
        this.installedMiners = [];
        this.runningMiners = [];
        console.error('API not available:', err);
        return of(null);
      })
    ).subscribe();
  }

  fetchAvailableMiners(): void {
    this.http.get<any[]>(`${this.apiBaseUrl}/miners/available`).subscribe({
      next: miners => { this.availableMiners = miners; },
      error: err => { this.handleError(err, 'Could not fetch available miners'); }
    });
  }

  fetchRunningMiners() {
    return this.http.get<any[]>(`${this.apiBaseUrl}/miners`).pipe(
      map(miners => { this.runningMiners = miners; }),
      catchError(err => {
        this.handleError(err, 'Could not fetch running miners');
        this.runningMiners = [];
        return of([]);
      })
    );
  }

  private performAction(action: any) {
    action.subscribe({
      next: () => {
        setTimeout(() => this.checkSystemState(), 1000);
      },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, 'Action failed');
      }
    });
  }

  installMiner(minerType: string): void {
    this.performAction(this.http.post(`${this.apiBaseUrl}/miners/${minerType}/install`, {}));
  }

  startMiner(miner: any, useLastConfig: boolean = false): void {
    let config = {};
    if (!useLastConfig) {
      config = {
        pool: this.poolAddress,
        wallet: this.walletAddress,
        tls: true,
        hugePages: true,
      };
    }
    this.performAction(this.http.post(`${this.apiBaseUrl}/miners/${miner.type}`, config));
    this.showStartOptionsFor = null;
  }

  stopMiner(miner: any): void {
    const runningInstance = this.getRunningMinerInstance(miner);
    if (!runningInstance) {
      this.error = "Cannot stop a miner that is not running.";
      return;
    }
    this.performAction(this.http.delete(`${this.apiBaseUrl}/miners/${runningInstance.name}`));
  }

  toggleStartOptions(minerType: string): void {
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
