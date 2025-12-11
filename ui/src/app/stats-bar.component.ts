import { Component, Input, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'snider-mining-stats-bar',
  standalone: true,
  imports: [CommonModule],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  template: `
    @if(stats) {
      <div class="stats-bar">
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
  `,
  styles: [`
    .stats-bar {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 0.5rem;
      padding: 0.5rem;
      background-color: #f9f9f9;
      border: 1px solid #eee;
      border-radius: 8px;
      margin-bottom: 1rem;
    }
    .stat-item {
      display: flex;
      flex-direction: column;
      align-items: center;
      background-color: #fff;
      padding: 0.5rem;
      border-radius: 6px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.05);
    }
    .label {
      font-size: 0.8rem;
      color: #666;
      margin-bottom: 0.25rem;
    }
    .value {
      font-weight: bold;
      font-size: 1.1rem;
    }
  `]
})
export class StatsBarComponent {
  @Input() stats: any;
}
