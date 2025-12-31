import { Component, inject, computed, signal, effect, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MinerService } from '../../miner.service';
import { NotificationService } from '../../notification.service';
import { TerminalModalComponent } from '../../terminal-modal.component';

export interface WorkerStats {
  name: string;
  hashrate: number;
  shares: number;
  rejected: number;
  uptime: number;
  pool: string;
  algorithm: string;
  cpu?: string;
  threads?: number;
}

@Component({
  selector: 'app-workers',
  standalone: true,
  imports: [CommonModule, FormsModule, TerminalModalComponent],
  template: `
    <div class="workers-page">
      <!-- Firewall Warning -->
      @if (firewallWarning()) {
        <div class="warning-banner">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
          <span>Miner running but no hashrate detected. This may indicate a firewall blocking the connection.</span>
          <div class="warning-actions">
            <button class="view-logs-btn" (click)="viewAffectedMinerLogs()">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
              </svg>
              View Logs
            </button>
            <button class="dismiss-btn" (click)="dismissFirewallWarning()">Dismiss</button>
          </div>
        </div>
      }

      <!-- Quick Actions Bar -->
      <div class="actions-bar">
        <div class="profile-selector">
          <select
            class="profile-select"
            [value]="selectedProfileId() || ''"
            (change)="onProfileSelect($event)">
            <option value="" disabled>Select profile...</option>
            @for (profile of state().profiles; track profile.id) {
              <option [value]="profile.id">{{ profile.name }}</option>
            }
          </select>
          <button
            class="btn btn-primary"
            [disabled]="!selectedProfileId() || state().runningMiners.length > 0 || starting()"
            (click)="startMining()">
            @if (starting()) {
              <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Starting...
            } @else {
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
              </svg>
              Start
            }
          </button>
        </div>

        @if (workers().length > 0) {
          <button class="btn btn-danger" [disabled]="stoppingAll()" (click)="stopAllWorkers()">
            @if (stoppingAll()) {
              <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
              </svg>
              Stopping...
            } @else {
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
              </svg>
              Stop All
            }
          </button>
        }
      </div>

      <!-- Workers Table -->
      @if (workers().length > 0) {
        <div class="workers-table-container">
          <table class="workers-table">
            <thead>
              <tr>
                <th>Worker</th>
                <th class="text-right">Hashrate</th>
                <th class="text-right">Shares</th>
                <th class="text-right">Efficiency</th>
                <th class="text-right">Uptime</th>
                <th>Pool</th>
                <th class="text-center">Actions</th>
              </tr>
            </thead>
            <tbody>
              @for (worker of workers(); track worker.name) {
                <tr>
                  <td>
                    <div class="worker-name">
                      <div class="status-dot online"></div>
                      <span>{{ worker.name }}</span>
                      @if (worker.algorithm) {
                        <span class="algo-badge">{{ worker.algorithm }}</span>
                      }
                    </div>
                  </td>
                  <td class="text-right tabular-nums">
                    <span class="hashrate-value">{{ formatHashrate(worker.hashrate) }}</span>
                    <span class="hashrate-unit">{{ getHashrateUnit(worker.hashrate) }}</span>
                    <div class="hashrate-bar">
                      <div class="hashrate-fill" [style.width.%]="getHashratePercent(worker)"></div>
                    </div>
                  </td>
                  <td class="text-right tabular-nums">
                    <span class="shares-good">{{ worker.shares }}</span>
                    @if (worker.rejected > 0) {
                      <span class="shares-rejected">/ {{ worker.rejected }}</span>
                    }
                  </td>
                  <td class="text-right tabular-nums">
                    <span [class.text-success-500]="getEfficiency(worker) >= 99"
                          [class.text-warning-500]="getEfficiency(worker) >= 95 && getEfficiency(worker) < 99"
                          [class.text-danger-500]="getEfficiency(worker) < 95">
                      {{ getEfficiency(worker).toFixed(1) }}%
                    </span>
                  </td>
                  <td class="text-right tabular-nums">{{ formatUptime(worker.uptime) }}</td>
                  <td class="pool-cell">{{ worker.pool }}</td>
                  <td class="actions-cell">
                    <button class="icon-btn" title="View logs" (click)="openTerminal(worker.name)">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
                      </svg>
                    </button>
                    <button
                      class="icon-btn icon-btn-danger"
                      title="Stop worker"
                      [disabled]="stoppingWorker() === worker.name"
                      (click)="stopWorker(worker.name)">
                      @if (stoppingWorker() === worker.name) {
                        <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                          <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                          <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                        </svg>
                      } @else {
                        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                        </svg>
                      }
                    </button>
                  </td>
                </tr>
              }
            </tbody>
          </table>
        </div>
      } @else {
        <div class="empty-state">
          <svg class="empty-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24" width="64" height="64">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2"/>
          </svg>
          <h3>No Active Workers</h3>
          <p>Select a profile and start mining to see workers here.</p>
        </div>
      }
    </div>

    <!-- Terminal Modal -->
    @if (terminalMinerName) {
      <app-terminal-modal
        [minerName]="terminalMinerName"
        (close)="closeTerminal()">
      </app-terminal-modal>
    }
  `,
  styles: [`
    .workers-page {
      display: flex;
      flex-direction: column;
      gap: 1rem;
    }

    .warning-banner {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      padding: 0.75rem 1rem;
      background: rgb(245 158 11 / 0.1);
      border: 1px solid rgb(245 158 11 / 0.3);
      border-radius: 0.5rem;
      color: var(--color-warning-500);
      font-size: 0.875rem;
    }

    .warning-banner span {
      flex: 1;
    }

    .warning-actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-left: auto;
    }

    .view-logs-btn {
      display: inline-flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.25rem 0.75rem;
      background: rgb(0 212 255 / 0.15);
      border: 1px solid var(--color-accent-500);
      border-radius: 0.25rem;
      color: var(--color-accent-500);
      cursor: pointer;
      font-size: 0.75rem;
      transition: all 0.15s ease;
    }

    .view-logs-btn:hover {
      background: rgb(0 212 255 / 0.25);
    }

    .warning-banner .dismiss-btn {
      padding: 0.25rem 0.75rem;
      background: transparent;
      border: 1px solid currentColor;
      border-radius: 0.25rem;
      color: inherit;
      cursor: pointer;
      font-size: 0.75rem;
    }

    .actions-bar {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
    }

    .profile-selector {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .profile-select {
      min-width: 200px;
      padding: 0.5rem 0.75rem;
      background: var(--color-surface-200);
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      color: white;
      font-size: 0.875rem;
    }

    .profile-select:focus {
      outline: none;
      border-color: var(--color-accent-500);
    }

    .btn {
      display: inline-flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.5rem 1rem;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.15s ease;
      border: none;
    }

    .btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    .btn-primary {
      background: var(--color-accent-500);
      color: #0f0f1a;
    }

    .btn-primary:hover:not(:disabled) {
      background: rgb(0 212 255 / 0.8);
    }

    .btn-danger {
      background: rgb(239 68 68 / 0.2);
      color: var(--color-danger-500);
    }

    .btn-danger:hover:not(:disabled) {
      background: rgb(239 68 68 / 0.3);
    }

    .workers-table-container {
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      overflow: hidden;
    }

    .workers-table {
      width: 100%;
      border-collapse: collapse;
    }

    .workers-table th {
      padding: 0.75rem 1rem;
      text-align: left;
      font-size: 0.75rem;
      font-weight: 600;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      background: var(--color-surface-200);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .workers-table td {
      padding: 0.75rem 1rem;
      font-size: 0.875rem;
      color: #e2e8f0;
      border-bottom: 1px solid rgb(37 37 66 / 0.1);
    }

    .workers-table tbody tr:hover {
      background: rgb(37 37 66 / 0.2);
    }

    .text-right {
      text-align: right;
    }

    .text-center {
      text-align: center;
    }

    .tabular-nums {
      font-family: var(--font-family-mono);
      font-variant-numeric: tabular-nums;
    }

    .worker-name {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
    }

    .status-dot.online {
      background: var(--color-success-500);
      box-shadow: 0 0 6px var(--color-success-500);
    }

    .algo-badge {
      padding: 0.125rem 0.375rem;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: var(--color-accent-500);
      text-transform: uppercase;
    }

    .hashrate-value {
      font-weight: 600;
      color: white;
    }

    .hashrate-unit {
      margin-left: 0.25rem;
      font-size: 0.75rem;
      color: #94a3b8;
    }

    .hashrate-bar {
      margin-top: 0.25rem;
      height: 3px;
      background: rgb(37 37 66 / 0.5);
      border-radius: 2px;
      overflow: hidden;
    }

    .hashrate-fill {
      height: 100%;
      background: var(--color-accent-500);
      transition: width 0.3s ease;
    }

    .shares-good {
      color: var(--color-success-500);
    }

    .shares-rejected {
      color: var(--color-danger-500);
      font-size: 0.75rem;
    }

    .pool-cell {
      max-width: 150px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .actions-cell {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
    }

    .icon-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: transparent;
      border: none;
      border-radius: 0.25rem;
      color: #94a3b8;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .icon-btn:hover {
      background: rgb(37 37 66 / 0.5);
      color: white;
    }

    .icon-btn-danger:hover {
      background: rgb(239 68 68 / 0.2);
      color: var(--color-danger-500);
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

    .empty-icon {
      width: 64px;
      height: 64px;
      color: #475569;
    }

    .animate-spin {
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    .icon-btn:disabled {
      opacity: 0.6;
      cursor: not-allowed;
    }

    /* Mobile responsive styles */
    @media (max-width: 768px) {
      .actions-bar {
        flex-direction: column;
        align-items: stretch;
      }

      .profile-selector {
        flex-direction: column;
      }

      .profile-select {
        min-width: 100%;
      }

      .btn {
        width: 100%;
        justify-content: center;
      }

      .workers-table-container {
        overflow-x: auto;
      }

      .workers-table {
        min-width: 600px;
      }

      .empty-state {
        padding: 2rem 1rem;
      }
    }
  `]
})
export class WorkersComponent implements OnDestroy {
  private minerService = inject(MinerService);
  private notifications = inject(NotificationService);
  state = this.minerService.state;

  selectedProfileId = signal<string | null>(null);
  terminalMinerName: string | null = null;
  firewallWarningDismissed = signal<boolean>(false);

  // Loading states
  starting = signal(false);
  stoppingWorker = signal<string | null>(null);
  stoppingAll = signal(false);

  firewallWarning = computed(() => {
    if (this.firewallWarningDismissed()) return false;
    const miners = this.state().runningMiners;
    if (miners.length === 0) return false;

    for (const miner of miners) {
      const uptime = miner.full_stats?.uptime || 0;
      const hashrate = miner.full_stats?.hashrate?.total?.[0] || 0;
      if (uptime >= 15 && hashrate === 0) return true;
    }
    return false;
  });

  constructor() {
    effect(() => {
      if (this.state().runningMiners.length === 0) {
        this.firewallWarningDismissed.set(false);
      }
    });
  }

  ngOnDestroy() {}

  workers = computed(() => {
    return this.state().runningMiners.map(miner => {
      const stats = miner.full_stats;
      return {
        name: miner.name,
        hashrate: stats?.hashrate?.total?.[0] || 0,
        shares: stats?.results?.shares_good || 0,
        rejected: (stats?.results?.shares_total || 0) - (stats?.results?.shares_good || 0),
        uptime: stats?.uptime || 0,
        pool: stats?.connection?.pool?.split(':')[0] || 'N/A',
        algorithm: stats?.algo || '',
        cpu: stats?.cpu?.brand,
        threads: stats?.cpu?.threads
      };
    });
  });

  totalHashrate = computed(() => this.workers().reduce((sum, w) => sum + w.hashrate, 0));

  dismissFirewallWarning() {
    this.firewallWarningDismissed.set(true);
  }

  viewAffectedMinerLogs() {
    // Find the first miner with 0 hashrate but uptime >= 15s (the firewall-affected one)
    const miners = this.state().runningMiners;
    for (const miner of miners) {
      const uptime = miner.full_stats?.uptime || 0;
      const hashrate = miner.full_stats?.hashrate?.total?.[0] || 0;
      if (uptime >= 15 && hashrate === 0) {
        this.openTerminal(miner.name);
        return;
      }
    }
    // Fallback: open first miner if any
    if (miners.length > 0) {
      this.openTerminal(miners[0].name);
    }
  }

  onProfileSelect(event: Event) {
    const select = event.target as HTMLSelectElement;
    this.selectedProfileId.set(select.value);
  }

  startMining() {
    const profileId = this.selectedProfileId();
    if (profileId) {
      const profile = this.state().profiles.find(p => p.id === profileId);
      const name = profile?.name || 'Miner';
      this.starting.set(true);
      this.minerService.startMiner(profileId).subscribe({
        next: () => {
          this.starting.set(false);
          this.notifications.success(`${name} started successfully`, 'Miner Started');
        },
        error: (err) => {
          this.starting.set(false);
          console.error('Failed to start miner:', err);
          this.notifications.error(`Failed to start ${name}: ${err.message || 'Unknown error'}`, 'Start Failed');
        }
      });
    }
  }

  stopWorker(name: string) {
    this.stoppingWorker.set(name);
    this.minerService.stopMiner(name).subscribe({
      next: () => {
        this.stoppingWorker.set(null);
        this.notifications.success(`${name} stopped`, 'Worker Stopped');
      },
      error: (err) => {
        this.stoppingWorker.set(null);
        console.error(`Failed to stop ${name}:`, err);
        this.notifications.error(`Failed to stop ${name}: ${err.message || 'Unknown error'}`, 'Stop Failed');
      }
    });
  }

  stopAllWorkers() {
    const workerCount = this.workers().length;
    let stoppedCount = 0;
    let errorCount = 0;

    this.stoppingAll.set(true);
    this.workers().forEach(w => {
      this.minerService.stopMiner(w.name).subscribe({
        next: () => {
          stoppedCount++;
          if (stoppedCount + errorCount === workerCount) {
            this.stoppingAll.set(false);
            if (errorCount === 0) {
              this.notifications.success(`All ${stoppedCount} workers stopped`, 'All Workers Stopped');
            } else {
              this.notifications.warning(`Stopped ${stoppedCount} workers, ${errorCount} failed`, 'Partial Stop');
            }
          }
        },
        error: (err) => {
          console.error(`Failed to stop ${w.name}:`, err);
          errorCount++;
          if (stoppedCount + errorCount === workerCount) {
            this.stoppingAll.set(false);
            this.notifications.warning(`Stopped ${stoppedCount} workers, ${errorCount} failed`, 'Partial Stop');
          }
        }
      });
    });
  }

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

  getHashratePercent(worker: WorkerStats): number {
    const total = this.totalHashrate();
    if (total === 0) return 0;
    return (worker.hashrate / total) * 100;
  }

  getEfficiency(worker: WorkerStats): number {
    const total = worker.shares + worker.rejected;
    if (total === 0) return 100;
    return (worker.shares / total) * 100;
  }

  openTerminal(minerName: string) {
    this.terminalMinerName = minerName;
  }

  closeTerminal() {
    this.terminalMinerName = null;
  }
}
