import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';

@Component({
  selector: 'app-stats-panel',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="stats-panel">
      <div class="stat-item">
        <svg class="stat-icon text-accent-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"/>
        </svg>
        <div class="stat-content">
          <span class="stat-value tabular-nums">{{ formatHashrate(totalHashrate()) }}</span>
          <span class="stat-unit">{{ getHashrateUnit(totalHashrate()) }}</span>
        </div>
        <span class="stat-label">Hashrate</span>
      </div>

      <div class="stat-divider"></div>

      <div class="stat-item">
        <svg class="stat-icon text-success-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <div class="stat-content">
          <span class="stat-value tabular-nums">{{ totalShares() }}</span>
          @if (totalRejected() > 0) {
            <span class="stat-rejected">/ {{ totalRejected() }}</span>
          }
        </div>
        <span class="stat-label">Shares</span>
      </div>

      <div class="stat-divider"></div>

      <div class="stat-item">
        <svg class="stat-icon text-accent-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <div class="stat-content">
          <span class="stat-value tabular-nums">{{ formatUptime(maxUptime()) }}</span>
        </div>
        <span class="stat-label">Uptime</span>
      </div>

      <div class="stat-divider"></div>

      <div class="stat-item">
        <svg class="stat-icon text-warning-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
        </svg>
        <div class="stat-content">
          <span class="stat-value pool-name">{{ poolName() }}</span>
        </div>
        <span class="stat-label">Pool</span>
      </div>

      <div class="stat-divider"></div>

      <div class="stat-item workers">
        <svg class="stat-icon" [class.text-success-500]="minerCount() > 0" [class.text-slate-500]="minerCount() === 0" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
        </svg>
        <div class="stat-content">
          @if (viewMode() === 'single') {
            <span class="stat-value single-label">{{ selectedMinerName() }}</span>
          } @else {
            <span class="stat-value tabular-nums">{{ minerCount() }}</span>
          }
        </div>
        <span class="stat-label">{{ viewMode() === 'single' ? 'Worker' : 'Workers' }}</span>
      </div>
    </div>
  `,
  styles: [`
    .stats-panel {
      display: flex;
      align-items: center;
      gap: 1.5rem;
      padding: 0.75rem 1.5rem;
      background: var(--color-surface-100);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .stat-item {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .stat-icon {
      width: 18px;
      height: 18px;
      flex-shrink: 0;
    }

    .stat-content {
      display: flex;
      align-items: baseline;
      gap: 0.25rem;
    }

    .stat-value {
      font-size: 0.9375rem;
      font-weight: 600;
      color: white;
      font-family: var(--font-family-mono);
    }

    .stat-value.pool-name,
    .stat-value.single-label {
      max-width: 120px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
      font-family: var(--font-family-sans);
      font-size: 0.875rem;
    }

    .stat-value.single-label {
      color: var(--color-accent-400);
    }

    .stat-unit {
      font-size: 0.75rem;
      color: #94a3b8;
      font-weight: 500;
    }

    .stat-rejected {
      font-size: 0.75rem;
      color: var(--color-danger-500);
      font-family: var(--font-family-mono);
    }

    .stat-label {
      font-size: 0.6875rem;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .stat-divider {
      width: 1px;
      height: 24px;
      background: rgb(37 37 66 / 0.3);
    }

    @media (max-width: 768px) {
      .stats-panel {
        gap: 0.75rem;
        padding: 0.5rem 1rem;
        overflow-x: auto;
      }

      .stat-label {
        display: none;
      }
    }
  `]
})
export class StatsPanelComponent {
  private minerService = inject(MinerService);
  private state = this.minerService.state;

  // Use displayedMiners which respects single/multi view mode
  miners = this.minerService.displayedMiners;
  viewMode = this.minerService.viewMode;
  selectedMinerName = this.minerService.selectedMinerName;

  totalHashrate = computed(() => {
    return this.miners().reduce((sum, m) => sum + (m.full_stats?.hashrate?.total?.[0] || 0), 0);
  });

  totalShares = computed(() => {
    return this.miners().reduce((sum, m) => sum + (m.full_stats?.results?.shares_good || 0), 0);
  });

  totalRejected = computed(() => {
    return this.miners().reduce((sum, m) => {
      const total = m.full_stats?.results?.shares_total || 0;
      const good = m.full_stats?.results?.shares_good || 0;
      return sum + (total - good);
    }, 0);
  });

  maxUptime = computed(() => {
    return Math.max(...this.miners().map(m => m.full_stats?.uptime || 0), 0);
  });

  poolName = computed(() => {
    const pools = [...new Set(this.miners()
      .map(m => m.full_stats?.connection?.pool?.split(':')[0])
      .filter(Boolean))];
    if (pools.length === 0) return 'Not connected';
    if (pools.length === 1) return pools[0];
    return `${pools.length} pools`;
  });

  minerCount = computed(() => this.miners().length);

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

  formatUptime(seconds: number): string {
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) {
      const mins = Math.floor(seconds / 60);
      const secs = seconds % 60;
      return `${mins}m ${secs}s`;
    }
    const hours = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${mins}m`;
  }
}
