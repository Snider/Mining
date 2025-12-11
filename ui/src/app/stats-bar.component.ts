import { Component, Input, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'snider-mining-stats-bar',
  standalone: true,
  imports: [CommonModule],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  template: `
    @if(stats) {
      @if (mode === 'list') {
        <div class="stats-container list-mode">
          <dl class="stats-dl">
            <!-- General -->
            <dt>Algorithm</dt><dd>{{ stats.algo }}</dd>
            <dt>Uptime</dt><dd>{{ stats.uptime }}s</dd>
            <dt>Version</dt><dd>{{ stats.version }}</dd>

            <!-- Hashrate -->
            <dt>Hashrate (10s)</dt><dd>{{ stats.hashrate?.total[0] | number:'1.0-2' }} H/s</dd>
            <dt>Hashrate (60s)</dt><dd>{{ stats.hashrate?.total[1] | number:'1.0-2' }} H/s</dd>
            <dt>Hashrate (15m)</dt><dd>{{ stats.hashrate?.total[2] | number:'1.0-2' }} H/s</dd>
            <dt>Highest Hashrate</dt><dd>{{ stats.hashrate?.highest | number:'1.0-2' }} H/s</dd>

            <!-- Results -->
            <dt>Good Shares</dt><dd>{{ stats.results?.shares_good }}</dd>
            <dt>Total Shares</dt><dd>{{ stats.results?.shares_total }}</dd>
            <dt>Avg. Time</dt><dd>{{ stats.results?.avg_time }}s</dd>
            <dt>Total Hashes</dt><dd>{{ stats.results?.hashes_total | number }}</dd>

            <!-- Connection -->
            <dt>Pool</dt><dd>{{ stats.connection?.pool }}</dd>
            <dt>Pool Uptime</dt><dd>{{ stats.connection?.uptime }}s</dd>
            <dt>Pool Ping</dt><dd>{{ stats.connection?.ping }}ms</dd>
            <dt>Current Difficulty</dt><dd>{{ stats.connection?.diff | number }}</dd>
            <dt>Accepted Shares</dt><dd>{{ stats.connection?.accepted }}</dd>
            <dt>Rejected Shares</dt><dd>{{ stats.connection?.rejected }}</dd>

            <!-- CPU -->
            <dt>CPU Brand</dt><dd>{{ stats.cpu?.brand }}</dd>
            <dt>CPU Cores/Threads</dt><dd>{{ stats.cpu?.cores }} / {{ stats.cpu?.threads }}</dd>
          </dl>
        </div>
      } @else {
        <div class="stats-bar bar-mode">
          <div class="stat-item">
            <span class="label">Hashrate:</span>
            <span class="value">{{ stats.hashrate?.total[0] | number:'1.0-2' }} H/s</span>
          </div>
          <div class="stat-item">
            <span class="label">Algorithm:</span>
            <span class="value">{{ stats.algo }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Difficulty:</span>
            <span class="value">{{ stats.connection?.diff | number }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Accepted:</span>
            <span class="value">{{ stats.results?.shares_good }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Rejected:</span>
            <span class="value">{{ stats.connection?.rejected }}</span>
          </div>
          <div class="stat-item">
            <span class="label">Avg Time:</span>
            <span class="value">{{ stats.results?.avg_time | number }}s</span>
          </div>
          <div class="stat-item">
            <span class="label">Uptime:</span>
            <span class="value">{{ stats.uptime | number }}s</span>
          </div>
          <div class="stat-item">
            <span class="label">Pool Uptime:</span>
            <span class="value">{{ stats.connection?.uptime | number }}s</span>
          </div>
        </div>
      }
    }
  `,
  styles: [`
    /* List Mode Styles */
    .stats-container.list-mode {
      height: 100%;
      max-height: 400px;
      overflow-y: auto;
      padding-right: 1rem;
      background-color: #f9f9f9;
      border: 1px solid #eee;
      border-radius: 8px;
    }
    .list-mode .stats-dl {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 0.5rem 1rem;
      padding: 1rem;
    }
    .list-mode dt {
      font-weight: bold;
      color: #555;
      grid-column: 1;
      white-space: nowrap;
    }
    .list-mode dd {
      margin: 0;
      grid-column: 2;
      text-align: right;
      font-family: monospace;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
    }

    /* Bar Mode Styles */
    .stats-bar.bar-mode {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 0.5rem;
      padding: 0.5rem;
      background-color: #f9f9f9;
      border: 1px solid #eee;
      border-radius: 8px;
      margin-bottom: 1rem;
    }
    .bar-mode .stat-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      background-color: #fff;
      padding: 0.5rem;
      border-radius: 6px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.05);
    }
    .bar-mode .label {
      font-size: 0.8rem;
      color: #666;
      margin-bottom: 0.25rem;
    }
    .bar-mode .value {
      font-weight: bold;
      font-size: 1.1rem;
    }
  `]
})
export class StatsBarComponent {
  @Input() stats: any;
  @Input() mode: 'bar' | 'list' = 'bar';
}
