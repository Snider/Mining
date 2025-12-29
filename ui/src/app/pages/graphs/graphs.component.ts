import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';
import { ChartComponent } from '../../chart.component';

@Component({
  selector: 'app-graphs',
  standalone: true,
  imports: [CommonModule, ChartComponent],
  template: `
    <div class="graphs-page">
      <div class="charts-grid">
        <!-- Main Hashrate Chart -->
        <div class="chart-card large">
          <div class="chart-header">
            <h3>Hashrate Over Time</h3>
            <div class="chart-legend">
              @for (miner of runningMiners(); track miner.name) {
                <div class="legend-item">
                  <div class="legend-color"></div>
                  <span>{{ miner.name }}</span>
                </div>
              }
            </div>
          </div>
          <div class="chart-body">
            @if (runningMiners().length > 0) {
              <snider-mining-chart></snider-mining-chart>
            } @else {
              <div class="chart-empty">
                <svg class="w-12 h-12 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                </svg>
                <p>Start mining to see hashrate graphs</p>
              </div>
            }
          </div>
        </div>

        <!-- Stats Cards -->
        <div class="stats-grid">
          <div class="stat-card">
            <div class="stat-icon text-accent-500">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6"/>
              </svg>
            </div>
            <div class="stat-info">
              <span class="stat-label">Peak Hashrate</span>
              <span class="stat-value">{{ formatHashrate(peakHashrate()) }} {{ getHashrateUnit(peakHashrate()) }}</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon text-success-500">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <div class="stat-info">
              <span class="stat-label">Efficiency</span>
              <span class="stat-value">{{ efficiency().toFixed(1) }}%</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon text-warning-500">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
              </svg>
            </div>
            <div class="stat-info">
              <span class="stat-label">Avg. Share Time</span>
              <span class="stat-value">{{ avgShareTime() }}s</span>
            </div>
          </div>

          <div class="stat-card">
            <div class="stat-icon text-accent-500">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </div>
            <div class="stat-info">
              <span class="stat-label">Difficulty</span>
              <span class="stat-value">{{ formatDifficulty(totalDifficulty()) }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .graphs-page {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .charts-grid {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .chart-card {
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      overflow: hidden;
    }

    .chart-card.large {
      min-height: 350px;
    }

    .chart-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1rem 1.25rem;
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .chart-header h3 {
      font-size: 0.9375rem;
      font-weight: 600;
      color: white;
    }

    .chart-legend {
      display: flex;
      align-items: center;
      gap: 1rem;
    }

    .legend-item {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      font-size: 0.75rem;
      color: #94a3b8;
    }

    .legend-color {
      width: 10px;
      height: 10px;
      background: var(--color-accent-500);
      border-radius: 2px;
    }

    .chart-body {
      padding: 1rem;
      min-height: 280px;
    }

    .chart-empty {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 250px;
      color: #64748b;
    }

    .chart-empty p {
      margin-top: 0.75rem;
      font-size: 0.875rem;
    }

    .stats-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
      gap: 1rem;
    }

    .stat-card {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 1rem 1.25rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
    }

    .stat-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 40px;
      height: 40px;
      background: currentColor;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.5rem;
    }

    .stat-icon svg {
      width: 20px;
      height: 20px;
    }

    .stat-info {
      display: flex;
      flex-direction: column;
      gap: 0.125rem;
    }

    .stat-label {
      font-size: 0.75rem;
      color: #64748b;
    }

    .stat-value {
      font-size: 1.125rem;
      font-weight: 600;
      color: white;
      font-family: var(--font-family-mono);
    }
  `]
})
export class GraphsComponent {
  private minerService = inject(MinerService);
  private state = this.minerService.state;

  runningMiners = computed(() => this.state().runningMiners);

  peakHashrate = computed(() => {
    return this.runningMiners().reduce((sum, m) => sum + (m.full_stats?.hashrate?.highest || 0), 0);
  });

  totalShares = computed(() => {
    return this.runningMiners().reduce((sum, m) => sum + (m.full_stats?.results?.shares_good || 0), 0);
  });

  totalRejected = computed(() => {
    return this.runningMiners().reduce((sum, m) => {
      const total = m.full_stats?.results?.shares_total || 0;
      const good = m.full_stats?.results?.shares_good || 0;
      return sum + (total - good);
    }, 0);
  });

  efficiency = computed(() => {
    const total = this.totalShares() + this.totalRejected();
    if (total === 0) return 100;
    return (this.totalShares() / total) * 100;
  });

  avgShareTime = computed(() => {
    const miners = this.runningMiners();
    if (miners.length === 0) return 0;
    const avgTimes = miners
      .map(m => m.full_stats?.results?.avg_time || 0)
      .filter(t => t > 0);
    if (avgTimes.length === 0) return 0;
    return Math.round(avgTimes.reduce((a, b) => a + b, 0) / avgTimes.length);
  });

  totalDifficulty = computed(() => {
    return this.runningMiners().reduce((sum, m) => sum + (m.full_stats?.connection?.diff || 0), 0);
  });

  formatHashrate(hashrate: number): string {
    if (hashrate >= 1000000000) return (hashrate / 1000000000).toFixed(2);
    if (hashrate >= 1000000) return (hashrate / 1000000).toFixed(2);
    if (hashrate >= 1000) return (hashrate / 1000).toFixed(2);
    return hashrate.toFixed(0);
  }

  getHashrateUnit(hashrate: number): string {
    if (hashrate >= 1000000000) return 'GH/s';
    if (hashrate >= 1000000) return 'MH/s';
    if (hashrate >= 1000) return 'kH/s';
    return 'H/s';
  }

  formatDifficulty(diff: number): string {
    if (diff >= 1000000) return (diff / 1000000).toFixed(1) + 'M';
    if (diff >= 1000) return (diff / 1000).toFixed(1) + 'K';
    return diff.toString();
  }
}
