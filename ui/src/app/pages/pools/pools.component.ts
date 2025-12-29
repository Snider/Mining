import { Component, inject, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';

interface PoolInfo {
  name: string;
  host: string;
  miners: string[];
  ping: number;
  difficulty: number;
  connected: boolean;
}

@Component({
  selector: 'app-pools',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="pools-page">
      <div class="page-header">
        <h2>Mining Pools</h2>
        <p>Active pool connections from running miners</p>
      </div>

      @if (pools().length > 0) {
        <div class="pools-grid">
          @for (pool of pools(); track pool.host) {
            <div class="pool-card" [class.connected]="pool.connected">
              <div class="pool-header">
                <div class="pool-status">
                  <div class="status-dot" [class.online]="pool.connected"></div>
                  <span class="pool-name">{{ pool.name }}</span>
                </div>
                <span class="pool-ping" [class.good]="pool.ping < 100" [class.warning]="pool.ping >= 100 && pool.ping < 200">
                  {{ pool.ping }}ms
                </span>
              </div>

              <div class="pool-host">{{ pool.host }}</div>

              <div class="pool-stats">
                <div class="pool-stat">
                  <span class="stat-label">Difficulty</span>
                  <span class="stat-value">{{ formatDifficulty(pool.difficulty) }}</span>
                </div>
                <div class="pool-stat">
                  <span class="stat-label">Workers</span>
                  <span class="stat-value">{{ pool.miners.length }}</span>
                </div>
              </div>

              <div class="pool-miners">
                @for (miner of pool.miners; track miner) {
                  <span class="miner-badge">{{ miner }}</span>
                }
              </div>
            </div>
          }
        </div>
      } @else {
        <div class="empty-state">
          <svg class="w-16 h-16 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M21 12a9 9 0 01-9 9m9-9a9 9 0 00-9-9m9 9H3m9 9a9 9 0 01-9-9m9 9c1.657 0 3-4.03 3-9s-1.343-9-3-9m0 18c-1.657 0-3-4.03-3-9s1.343-9 3-9m-9 9a9 9 0 019-9"/>
          </svg>
          <h3>No Pool Connections</h3>
          <p>Start mining to see active pool connections here.</p>
        </div>
      }
    </div>
  `,
  styles: [`
    .pools-page {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }

    .page-header h2 {
      font-size: 1.25rem;
      font-weight: 600;
      color: white;
    }

    .page-header p {
      margin-top: 0.25rem;
      font-size: 0.875rem;
      color: #64748b;
    }

    .pools-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 1rem;
    }

    .pool-card {
      padding: 1.25rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      transition: border-color 0.15s ease;
    }

    .pool-card.connected {
      border-color: rgb(16 185 129 / 0.3);
    }

    .pool-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 0.75rem;
    }

    .pool-status {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #64748b;
    }

    .status-dot.online {
      background: var(--color-success-500);
      box-shadow: 0 0 6px var(--color-success-500);
    }

    .pool-name {
      font-size: 1rem;
      font-weight: 600;
      color: white;
    }

    .pool-ping {
      font-family: var(--font-family-mono);
      font-size: 0.8125rem;
      color: var(--color-danger-500);
    }

    .pool-ping.good {
      color: var(--color-success-500);
    }

    .pool-ping.warning {
      color: var(--color-warning-500);
    }

    .pool-host {
      font-size: 0.8125rem;
      color: #64748b;
      font-family: var(--font-family-mono);
      margin-bottom: 1rem;
      word-break: break-all;
    }

    .pool-stats {
      display: flex;
      gap: 2rem;
      padding: 0.75rem 0;
      border-top: 1px solid rgb(37 37 66 / 0.2);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
      margin-bottom: 0.75rem;
    }

    .pool-stat {
      display: flex;
      flex-direction: column;
      gap: 0.125rem;
    }

    .stat-label {
      font-size: 0.6875rem;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .stat-value {
      font-size: 0.9375rem;
      font-weight: 600;
      color: white;
      font-family: var(--font-family-mono);
    }

    .pool-miners {
      display: flex;
      flex-wrap: wrap;
      gap: 0.375rem;
    }

    .miner-badge {
      padding: 0.25rem 0.5rem;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.25rem;
      font-size: 0.75rem;
      color: var(--color-accent-500);
    }

    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 4rem 2rem;
      text-align: center;
    }

    .empty-state h3 {
      margin-top: 1rem;
      font-size: 1.125rem;
      font-weight: 600;
      color: white;
    }

    .empty-state p {
      margin-top: 0.5rem;
      color: #64748b;
      font-size: 0.875rem;
    }
  `]
})
export class PoolsComponent {
  private minerService = inject(MinerService);
  private state = this.minerService.state;

  pools = computed<PoolInfo[]>(() => {
    const poolMap = new Map<string, PoolInfo>();

    for (const miner of this.state().runningMiners) {
      const conn = miner.full_stats?.connection;
      if (!conn?.pool) continue;

      const host = conn.pool;
      const name = host.split(':')[0];

      if (!poolMap.has(host)) {
        poolMap.set(host, {
          name,
          host,
          miners: [],
          ping: conn.ping || 0,
          difficulty: conn.diff || 0,
          connected: true
        });
      }

      const pool = poolMap.get(host)!;
      pool.miners.push(miner.name);
      // Average ping if multiple miners
      pool.ping = Math.round((pool.ping + (conn.ping || 0)) / 2);
      pool.difficulty = Math.max(pool.difficulty, conn.diff || 0);
    }

    return Array.from(poolMap.values());
  });

  formatDifficulty(diff: number): string {
    if (diff >= 1000000) return (diff / 1000000).toFixed(1) + 'M';
    if (diff >= 1000) return (diff / 1000).toFixed(1) + 'K';
    return diff.toString();
  }
}
