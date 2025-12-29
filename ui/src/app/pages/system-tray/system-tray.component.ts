import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';

@Component({
  selector: 'app-system-tray',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="tray-container">
      <!-- Quick Controls -->
      <div class="tray-header">
        <div class="tray-logo">
          <svg class="w-5 h-5 text-accent-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
          </svg>
          <span>Mining</span>
        </div>
        <div class="tray-controls">
          @if (minerCount() === 0) {
            <button class="control-btn start" (click)="startFirstProfile()" [disabled]="profiles().length === 0">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
              </svg>
            </button>
          } @else {
            <button class="control-btn stop" (click)="stopAll()">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
              </svg>
            </button>
          }
        </div>
      </div>

      <!-- Aggregate Stats -->
      <div class="tray-stats">
        <div class="stat-row">
          <span class="stat-label">Total Hashrate</span>
          <span class="stat-value">
            {{ formatHashrate(totalHashrate()) }}
            <span class="stat-unit">{{ getHashrateUnit(totalHashrate()) }}</span>
          </span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Shares</span>
          <span class="stat-value">
            {{ totalShares() }}
            @if (totalRejected() > 0) {
              <span class="rejected">/ {{ totalRejected() }}</span>
            }
          </span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Efficiency</span>
          <span class="stat-value" [class.good]="efficiency() >= 99" [class.warning]="efficiency() < 99 && efficiency() >= 95">
            {{ efficiency().toFixed(1) }}%
          </span>
        </div>
        <div class="stat-row">
          <span class="stat-label">Pool</span>
          <span class="stat-value pool">{{ poolName() }}</span>
        </div>
      </div>

      <!-- Workers List -->
      <div class="workers-section">
        <div class="section-header">
          <span>Workers ({{ minerCount() }})</span>
        </div>
        <div class="workers-list">
          @if (workers().length > 0) {
            @for (worker of workers(); track worker.name) {
              <div class="worker-row">
                <div class="worker-info">
                  <div class="worker-status online"></div>
                  <span class="worker-name">{{ worker.name }}</span>
                </div>
                <div class="worker-stats">
                  <span class="worker-hashrate">
                    {{ formatHashrate(worker.hashrate) }}
                    <span class="unit">{{ getHashrateUnit(worker.hashrate) }}</span>
                  </span>
                  <div class="sparkline">
                    <svg viewBox="0 0 60 20" preserveAspectRatio="none">
                      <polyline
                        fill="none"
                        stroke="currentColor"
                        stroke-width="1.5"
                        [attr.points]="getSparklinePoints(worker.name)"
                      />
                    </svg>
                  </div>
                </div>
              </div>
            }
          } @else {
            <div class="no-workers">
              <span>No active workers</span>
            </div>
          }
        </div>
      </div>

      <!-- Connection Status -->
      <div class="tray-footer">
        <div class="connection-status" [class.connected]="minerCount() > 0">
          <div class="status-dot"></div>
          <span>{{ minerCount() > 0 ? 'Mining Active' : 'Idle' }}</span>
        </div>
        @if (poolPing() > 0) {
          <span class="ping">{{ poolPing() }}ms</span>
        }
      </div>
    </div>
  `,
  styles: [`
    :host {
      display: block;
      width: 400px;
      height: 560px;
      overflow: hidden;
    }

    .tray-container {
      display: flex;
      flex-direction: column;
      height: 100%;
      background: var(--color-surface-200);
      font-family: var(--font-family-sans);
    }

    .tray-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.75rem 1rem;
      background: var(--color-surface-100);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .tray-logo {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.9375rem;
      font-weight: 600;
      color: white;
    }

    .tray-controls {
      display: flex;
      gap: 0.5rem;
    }

    .control-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      background: transparent;
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      color: #94a3b8;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .control-btn:hover:not(:disabled) {
      background: rgb(37 37 66 / 0.3);
      color: white;
    }

    .control-btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    .control-btn.start {
      color: var(--color-success-500);
      border-color: rgb(16 185 129 / 0.3);
    }

    .control-btn.start:hover:not(:disabled) {
      background: rgb(16 185 129 / 0.1);
    }

    .control-btn.stop {
      color: var(--color-danger-500);
      border-color: rgb(239 68 68 / 0.3);
    }

    .control-btn.stop:hover {
      background: rgb(239 68 68 / 0.1);
    }

    .tray-stats {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      padding: 0.875rem 1rem;
      background: var(--color-surface-100);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .stat-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }

    .stat-label {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .stat-value {
      font-size: 0.875rem;
      font-weight: 600;
      color: white;
      font-family: var(--font-family-mono);
    }

    .stat-value .stat-unit {
      font-size: 0.6875rem;
      font-weight: 500;
      color: #64748b;
      margin-left: 0.125rem;
    }

    .stat-value .rejected {
      color: var(--color-danger-500);
      font-size: 0.75rem;
    }

    .stat-value.good {
      color: var(--color-success-500);
    }

    .stat-value.warning {
      color: var(--color-warning-500);
    }

    .stat-value.pool {
      font-family: var(--font-family-sans);
      font-weight: 500;
      max-width: 160px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .workers-section {
      flex: 1;
      display: flex;
      flex-direction: column;
      min-height: 0;
    }

    .section-header {
      padding: 0.625rem 1rem;
      font-size: 0.6875rem;
      font-weight: 600;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      background: var(--color-surface-200);
      border-bottom: 1px solid rgb(37 37 66 / 0.1);
    }

    .workers-list {
      flex: 1;
      overflow-y: auto;
      padding: 0.5rem;
    }

    .worker-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.5rem 0.625rem;
      border-radius: 0.375rem;
      transition: background 0.15s ease;
    }

    .worker-row:hover {
      background: rgb(37 37 66 / 0.3);
    }

    .worker-info {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .worker-status {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: #64748b;
    }

    .worker-status.online {
      background: var(--color-success-500);
      box-shadow: 0 0 4px var(--color-success-500);
    }

    .worker-name {
      font-size: 0.8125rem;
      color: #e2e8f0;
    }

    .worker-stats {
      display: flex;
      align-items: center;
      gap: 0.75rem;
    }

    .worker-hashrate {
      font-size: 0.8125rem;
      font-weight: 600;
      color: white;
      font-family: var(--font-family-mono);
    }

    .worker-hashrate .unit {
      font-size: 0.625rem;
      color: #64748b;
      margin-left: 0.125rem;
    }

    .sparkline {
      width: 60px;
      height: 20px;
      color: var(--color-accent-500);
    }

    .sparkline svg {
      width: 100%;
      height: 100%;
    }

    .no-workers {
      display: flex;
      align-items: center;
      justify-content: center;
      height: 100px;
      color: #64748b;
      font-size: 0.8125rem;
    }

    .tray-footer {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.625rem 1rem;
      background: var(--color-surface-100);
      border-top: 1px solid rgb(37 37 66 / 0.2);
    }

    .connection-status {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.75rem;
      color: #64748b;
    }

    .connection-status .status-dot {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: #64748b;
    }

    .connection-status.connected .status-dot {
      background: var(--color-success-500);
      box-shadow: 0 0 4px var(--color-success-500);
    }

    .connection-status.connected {
      color: var(--color-success-500);
    }

    .ping {
      font-size: 0.75rem;
      color: #94a3b8;
      font-family: var(--font-family-mono);
    }
  `]
})
export class SystemTrayComponent {
  private minerService = inject(MinerService);
  private state = this.minerService.state;

  // Simulated sparkline data (would come from real history in production)
  private sparklineData = new Map<string, number[]>();

  profiles = () => this.state().profiles;
  minerCount = computed(() => this.state().runningMiners.length);

  workers = computed(() => {
    return this.state().runningMiners.map(miner => {
      const stats = miner.full_stats;
      const hashrate = stats?.hashrate?.total?.[0] || 0;

      // Update sparkline data
      if (!this.sparklineData.has(miner.name)) {
        this.sparklineData.set(miner.name, []);
      }
      const data = this.sparklineData.get(miner.name)!;
      data.push(hashrate);
      if (data.length > 20) data.shift();

      return {
        name: miner.name,
        hashrate,
        shares: stats?.results?.shares_good || 0,
        rejected: (stats?.results?.shares_total || 0) - (stats?.results?.shares_good || 0)
      };
    });
  });

  totalHashrate = computed(() => {
    return this.workers().reduce((sum, w) => sum + w.hashrate, 0);
  });

  totalShares = computed(() => {
    return this.workers().reduce((sum, w) => sum + w.shares, 0);
  });

  totalRejected = computed(() => {
    return this.workers().reduce((sum, w) => sum + w.rejected, 0);
  });

  efficiency = computed(() => {
    const total = this.totalShares() + this.totalRejected();
    if (total === 0) return 100;
    return (this.totalShares() / total) * 100;
  });

  poolName = computed(() => {
    const pools = [...new Set(this.state().runningMiners
      .map(m => m.full_stats?.connection?.pool?.split(':')[0])
      .filter(Boolean))];
    if (pools.length === 0) return 'Not connected';
    if (pools.length === 1) return pools[0];
    return `${pools.length} pools`;
  });

  poolPing = computed(() => {
    const pings = this.state().runningMiners
      .map(m => m.full_stats?.connection?.ping || 0)
      .filter(p => p > 0);
    if (pings.length === 0) return 0;
    return Math.round(pings.reduce((a, b) => a + b, 0) / pings.length);
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

  getSparklinePoints(minerName: string): string {
    const data = this.sparklineData.get(minerName) || [];
    if (data.length < 2) return '0,10 60,10';

    const max = Math.max(...data, 1);
    const points = data.map((value, i) => {
      const x = (i / (data.length - 1)) * 60;
      const y = 18 - (value / max) * 16;
      return `${x},${y}`;
    });
    return points.join(' ');
  }

  startFirstProfile() {
    const profiles = this.profiles();
    if (profiles.length > 0) {
      this.minerService.startMiner(profiles[0].id).subscribe({
        error: (err) => console.error('Failed to start miner:', err)
      });
    }
  }

  stopAll() {
    this.state().runningMiners.forEach(miner => {
      this.minerService.stopMiner(miner.name).subscribe({
        error: (err) => console.error(`Failed to stop ${miner.name}:`, err)
      });
    });
  }
}
