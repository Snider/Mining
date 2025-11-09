import {
  Component,
  OnInit,
  Input,
  OnDestroy,
  ElementRef,
  ViewEncapsulation,
  CUSTOM_ELEMENTS_SCHEMA
} from '@angular/core';
import { HttpClient, HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { interval, Subscription } from 'rxjs';
import { switchMap, startWith } from 'rxjs/operators';
import "@awesome.me/webawesome/dist/webawesome.js";
// Import Shoelace components
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';
// Add other Shoelace components as needed for your UI

interface HashratePoint {
  timestamp: string; // ISO string
  hashrate: number;
}

@Component({
  selector: 'mde-mining-dashboard', // This will be your custom element tag
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, HttpClientModule], // HttpClientModule is needed for HttpClient
  template: `
    <wa-card class="mining-dashboard-card">
      <div slot="header" class="card-header">
        <wa-icon name="cpu"></wa-icon>
        <h3>Mining Dashboard: {{ minerName }}</h3>
      </div>

      <p *ngIf="loading">Loading hashrate data...</p>
      <p *ngIf="error" class="error-message">Error: {{ error }}</p>

      <div *ngIf="!loading && !error">
        <p>Current Hashrate: <strong>{{ currentHashrate }} H/s</strong></p>
        <p>Last Updated: {{ lastUpdated | date:'mediumTime' }}</p>

        <h4>Hashrate History ({{ hashrateHistory.length }} points)</h4>
        <ul class="history-list">
          <li *ngFor="let point of hashrateHistory | slice:0:5">
            {{ point.timestamp | date:'mediumTime' }}: {{ point.hashrate }} H/s
          </li>
          <li *ngIf="hashrateHistory.length > 5">...</li>
        </ul>
        <!-- Here you would integrate a charting library -->
        <!-- <canvas #hashrateChart></canvas> -->
      </div>

      <div slot="footer" class="card-footer">
        <wa-button variant="brand" (click)="fetchHashrate()">Refresh</wa-button>
        <wa-button variant="neutral" (click)="toggleDetails()">{{ showDetails ? 'Hide Details' : 'Show Details' }}</wa-button>
      </div>

      <div *ngIf="showDetails" class="details-section">
        <h5>Raw History Data:</h5>
        <pre>{{ hashrateHistory | json }}</pre>
      </div>
    </wa-card>
  `,
  styles: [`
    .mining-dashboard-card {
      width: 100%;
      max-width: 500px;
      margin: 20px auto;
      border: 1px solid var(--wa-color-neutral-300);
      border-radius: var(--wa-border-radius-medium);
      box-shadow: var(--wa-shadow-medium);
    }
    .card-header {
      display: flex;
      align-items: center;
      gap: 10px;
      padding: var(--wa-spacing-medium);
      border-bottom: 1px solid var(--wa-color-neutral-200);
    }
    .card-header h3 {
      margin: 0;
      font-size: var(--wa-font-size-large);
    }
    .card-header sl-icon {
      font-size: var(--wa-font-size-x-large);
      color: var(--wa-color-primary-500);
    }
    .error-message {
      color: var(--wa-color-danger-500);
      font-weight: bold;
    }
    .history-list {
      list-style-type: none;
      padding: 0;
      max-height: 150px;
      overflow-y: auto;
      border: 1px solid var(--wa-color-neutral-100);
      border-radius: var(--wa-border-radius-small);
      padding: var(--wa-spacing-x-small);
      background-color: var(--wa-color-neutral-50);
    }
    .history-list li {
      padding: var(--wa-spacing-2x-small) 0;
      border-bottom: 1px dotted var(--wa-color-neutral-100);
    }
    .history-list li:last-child {
      border-bottom: none;
    }
    .card-footer {
      display: flex;
      justify-content: flex-end;
      gap: var(--wa-spacing-small);
      padding-top: var(--wa-spacing-medium);
      border-top: 1px solid var(--wa-color-neutral-200);
    }
    .details-section {
      margin-top: var(--wa-spacing-medium);
      padding: var(--wa-spacing-small);
      background-color: var(--wa-color-neutral-50);
      border: 1px solid var(--wa-color-neutral-200);
      border-radius: var(--wa-border-radius-small);
    }
    .details-section pre {
      white-space: pre-wrap;
      word-break: break-all;
      font-size: var(--wa-font-size-x-small);
      max-height: 200px;
      overflow-y: auto;
    }
  `],
  encapsulation: ViewEncapsulation.ShadowDom // Crucial for isolation
})
export class MiningDashboardElementComponent implements OnInit, OnDestroy {
  @Input() minerName: string = 'xmrig'; // Default miner name
  @Input() apiBaseUrl: string = 'http://localhost:9090/api/v1/mining'; // Default API base URL

  hashrateHistory: HashratePoint[] = [];
  currentHashrate: number = 0;
  lastUpdated: Date | null = null;
  loading: boolean = true;
  error: string | null = null;
  showDetails: boolean = false;

  private refreshSubscription: Subscription | undefined;

  constructor(private http: HttpClient, private elementRef: ElementRef) {}

  ngOnInit(): void {
    this.startAutoRefresh();
  }

  ngOnDestroy(): void {
    this.stopAutoRefresh();
  }

  startAutoRefresh(): void {
    this.stopAutoRefresh(); // Stop any existing refresh
    this.refreshSubscription = interval(10000) // Refresh every 10 seconds
      .pipe(startWith(0), switchMap(() => this.fetchHashrateObservable()))
      .subscribe({
        next: (history) => {
          this.hashrateHistory = history;
          if (history.length > 0) {
            this.currentHashrate = history[history.length - 1].hashrate;
            this.lastUpdated = new Date(history[history.length - 1].timestamp);
          } else {
            this.currentHashrate = 0;
            this.lastUpdated = null;
          }
          this.loading = false;
          this.error = null;
        },
        error: (err) => {
          console.error('Failed to fetch hashrate history:', err);
          this.error = 'Failed to fetch hashrate history.';
          this.loading = false;
        }
      });
  }

  stopAutoRefresh(): void {
    if (this.refreshSubscription) {
      this.refreshSubscription.unsubscribe();
      this.refreshSubscription = undefined;
    }
  }

  fetchHashrate(): void {
    this.loading = true;
    this.error = null;
    this.fetchHashrateObservable().subscribe({
      next: (history) => {
        this.hashrateHistory = history;
        if (history.length > 0) {
          this.currentHashrate = history[history.length - 1].hashrate;
          this.lastUpdated = new Date(history[history.length - 1].timestamp);
        } else {
          this.currentHashrate = 0;
          this.lastUpdated = null;
        }
        this.loading = false;
      },
      error: (err) => {
        console.error('Failed to fetch hashrate history:', err);
        this.error = 'Failed to fetch hashrate history.';
        this.loading = false;
      }
    });
  }

  private fetchHashrateObservable() {
    const url = `${this.apiBaseUrl}/miners/${this.minerName}/hashrate-history`;
    return this.http.get<HashratePoint[]>(url);
  }

  toggleDetails(): void {
    this.showDetails = !this.showDetails;
  }
}
