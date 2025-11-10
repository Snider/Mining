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
import * as Highcharts from 'highcharts';
import { HighchartsChartComponent } from 'highcharts-angular';

// Import Shoelace components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/tooltip/tooltip.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';

interface HashratePoint {
  timestamp: string; // ISO string
  hashrate: number;
}

@Component({
  selector: 'mde-mining-dashboard',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, HttpClientModule, HighchartsChartComponent],
  templateUrl: './app.html',
  styleUrls: ["app.css"],
  encapsulation: ViewEncapsulation.ShadowDom
})
export class MiningDashboardElementComponent implements OnInit, OnDestroy {
  @Input() minerName: string = 'xmrig';
  @Input() apiBaseUrl: string = 'http://localhost:9090/api/v1/mining';

  hashrateHistory: HashratePoint[] = [];
  currentHashrate: number = 0;
  lastUpdated: Date | null = null;
  loading: boolean = true;
  error: string | null = null;
  showDetails: boolean = false;

  private refreshSubscription: Subscription | undefined;

  chartOptions: Highcharts.Options = {
    chart: {
      type: 'spline',
    },
    title: {
      text: 'Live Hashrate'
    },
    xAxis: {
      type: 'datetime',
      title: {
        text: 'Time'
      }
    },
    yAxis: {
      title: {
        text: 'Hashrate (H/s)'
      },
      min: 0
    },
    series: [{
      name: 'Hashrate',
      type: 'line',
      data: []
    }],
    credits: {
      enabled: false
    }
  };
  updateFlag = false;

  constructor(private http: HttpClient, private elementRef: ElementRef) {}

  ngOnInit(): void {
    this.startAutoRefresh();
  }

  ngOnDestroy(): void {
    this.stopAutoRefresh();
  }

  startAutoRefresh(): void {
    this.stopAutoRefresh();
    this.refreshSubscription = interval(10000)
      .pipe(startWith(0), switchMap(() => this.fetchHashrateObservable()))
      .subscribe({
        next: (history) => {
          this.hashrateHistory = history;
          if (history && history.length > 0) {
            this.currentHashrate = history[history.length - 1].hashrate;
            this.lastUpdated = new Date(history[history.length - 1].timestamp);

            const chartData = history.map(point => [
              new Date(point.timestamp).getTime(),
              point.hashrate
            ]);

            // Safely update the chart data with type assertion
            if (this.chartOptions.series && this.chartOptions.series[0]) {
              (this.chartOptions.series[0] as Highcharts.SeriesLineOptions).data = chartData;
              this.updateFlag = true; // Trigger chart update
            }
          } else {
            this.currentHashrate = 0;
            this.lastUpdated = null;
            // Safely clear the chart data
            if (this.chartOptions.series && this.chartOptions.series[0]) {
               (this.chartOptions.series[0] as Highcharts.SeriesLineOptions).data = [];
              this.updateFlag = true;
            }
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

  private fetchHashrateObservable() {
    const url = `${this.apiBaseUrl}/miners/${this.minerName}/hashrate-history`;
    return this.http.get<HashratePoint[]>(url);
  }

  toggleDetails(): void {
    this.showDetails = !this.showDetails;
  }
}
