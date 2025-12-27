import { Component, ViewEncapsulation, CUSTOM_ELEMENTS_SCHEMA, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MinerService } from './miner.service';
import { ChartComponent } from './chart.component';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/tooltip/tooltip.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';
import '@awesome.me/webawesome/dist/components/spinner/spinner.js';
import '@awesome.me/webawesome/dist/components/input/input.js';
import '@awesome.me/webawesome/dist/components/select/select.js';
import '@awesome.me/webawesome/dist/components/badge/badge.js';
import '@awesome.me/webawesome/dist/components/details/details.js';
import '@awesome.me/webawesome/dist/components/tab-group/tab-group.js';
import '@awesome.me/webawesome/dist/components/tab/tab.js';
import '@awesome.me/webawesome/dist/components/tab-panel/tab-panel.js';

// Worker stats interface
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
  selector: 'snider-mining-dashboard',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, FormsModule, ChartComponent],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class MiningDashboardComponent {
  minerService = inject(MinerService);
  state = this.minerService.state;
  error = signal<string | null>(null);
  selectedProfileId = signal<string | null>(null);
  selectedMinerName = signal<string | null>(null); // For individual miner view

  // All running miners
  runningMiners = computed(() => this.state().runningMiners);

  // Worker stats for table display
  workers = computed<WorkerStats[]>(() => {
    return this.runningMiners().map(miner => {
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

  // Aggregate stats across all miners
  totalHashrate = computed(() => {
    return this.workers().reduce((sum, w) => sum + w.hashrate, 0);
  });

  totalShares = computed(() => {
    return this.workers().reduce((sum, w) => sum + w.shares, 0);
  });

  totalRejected = computed(() => {
    return this.workers().reduce((sum, w) => sum + w.rejected, 0);
  });

  minerCount = computed(() => this.runningMiners().length);

  // For single miner view (when selected)
  selectedMiner = computed(() => {
    const name = this.selectedMinerName();
    if (!name) return null;
    return this.runningMiners().find(m => m.name === name) || null;
  });

  // Stats for selected miner or first miner (for backward compatibility)
  stats = computed(() => {
    const selected = this.selectedMiner();
    if (selected) return selected.full_stats;
    const miners = this.runningMiners();
    return miners.length > 0 ? miners[0].full_stats : null;
  });

  currentHashrate = computed(() => {
    // Show total hashrate in overview mode
    return this.totalHashrate();
  });

  peakHashrate = computed(() => {
    // Sum of all peak hashrates
    return this.runningMiners().reduce((sum, m) => sum + (m.full_stats?.hashrate?.highest || 0), 0);
  });

  acceptedShares = computed(() => {
    return this.totalShares();
  });

  rejectedShares = computed(() => {
    return this.totalRejected();
  });

  uptime = computed(() => {
    // Show max uptime across all miners
    return Math.max(...this.workers().map(w => w.uptime), 0);
  });

  poolName = computed(() => {
    const pools = [...new Set(this.workers().map(w => w.pool).filter(p => p && p !== 'N/A'))];
    if (pools.length === 0) return 'Not connected';
    if (pools.length === 1) return pools[0];
    return `${pools.length} pools`;
  });

  poolPing = computed(() => {
    const pings = this.runningMiners()
      .map(m => m.full_stats?.connection?.ping || 0)
      .filter(p => p > 0);
    if (pings.length === 0) return 0;
    return Math.round(pings.reduce((a, b) => a + b, 0) / pings.length);
  });

  minerName = computed(() => {
    const count = this.minerCount();
    if (count === 0) return '';
    if (count === 1) return this.runningMiners()[0].name;
    return `${count} workers`;
  });

  algorithm = computed(() => {
    const algos = [...new Set(this.workers().map(w => w.algorithm).filter(Boolean))];
    if (algos.length === 0) return '';
    if (algos.length === 1) return algos[0];
    return algos.join(', ');
  });

  difficulty = computed(() => {
    // Sum of difficulties (for aggregate view)
    return this.runningMiners().reduce((sum, m) => sum + (m.full_stats?.connection?.diff || 0), 0);
  });

  // Format hashrate for display (e.g., 12345 -> "12.35")
  formatHashrate(hashrate: number): string {
    if (hashrate >= 1000000000) {
      return (hashrate / 1000000000).toFixed(2);
    } else if (hashrate >= 1000000) {
      return (hashrate / 1000000).toFixed(2);
    } else if (hashrate >= 1000) {
      return (hashrate / 1000).toFixed(2);
    }
    return hashrate.toFixed(0);
  }

  // Get hashrate unit
  getHashrateUnit(hashrate: number): string {
    if (hashrate >= 1000000000) {
      return 'GH/s';
    } else if (hashrate >= 1000000) {
      return 'MH/s';
    } else if (hashrate >= 1000) {
      return 'kH/s';
    }
    return 'H/s';
  }

  // Format uptime to human readable
  formatUptime(seconds: number): string {
    if (seconds < 60) {
      return `${seconds}s`;
    } else if (seconds < 3600) {
      const mins = Math.floor(seconds / 60);
      const secs = seconds % 60;
      return `${mins}m ${secs}s`;
    } else {
      const hours = Math.floor(seconds / 3600);
      const mins = Math.floor((seconds % 3600) / 60);
      return `${hours}h ${mins}m`;
    }
  }

  // Profile selection
  onProfileSelect(event: Event) {
    const select = event.target as HTMLSelectElement;
    this.selectedProfileId.set(select.value);
  }

  // Start mining with selected profile
  startMining() {
    const profileId = this.selectedProfileId();
    if (profileId) {
      this.minerService.startMiner(profileId).subscribe({
        error: (err) => {
          this.error.set(err.error?.error || 'Failed to start miner');
        }
      });
    }
  }

  // Stop the running miner
  stopMiner() {
    const minerName = this.minerName();
    if (minerName) {
      this.minerService.stopMiner(minerName).subscribe({
        error: (err) => {
          this.error.set(err.error?.error || 'Failed to stop miner');
        }
      });
    }
  }

  // Stop a specific worker
  stopWorker(name: string) {
    this.minerService.stopMiner(name).subscribe({
      error: (err) => {
        this.error.set(err.error?.error || `Failed to stop ${name}`);
      }
    });
  }

  // Stop all workers
  stopAllWorkers() {
    const workers = this.workers();
    workers.forEach(w => {
      this.minerService.stopMiner(w.name).subscribe({
        error: (err) => {
          console.error(`Failed to stop ${w.name}:`, err);
        }
      });
    });
  }

  // Select a specific miner for detailed view
  selectWorker(name: string) {
    this.selectedMinerName.set(name);
  }

  // Clear miner selection (go back to overview)
  clearSelection() {
    this.selectedMinerName.set(null);
  }

  // Get hashrate percentage for a worker (for bar visualization)
  getHashratePercent(worker: WorkerStats): number {
    const total = this.totalHashrate();
    if (total === 0) return 0;
    return (worker.hashrate / total) * 100;
  }

  // Get efficiency (accepted / total shares)
  getEfficiency(worker: WorkerStats): number {
    const total = worker.shares + worker.rejected;
    if (total === 0) return 100;
    return (worker.shares / total) * 100;
  }
}
