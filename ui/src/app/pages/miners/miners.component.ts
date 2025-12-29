import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';

@Component({
  selector: 'app-miners',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="miners-page">
      <div class="page-header">
        <div>
          <h2>Miner Software</h2>
          <p>Install and manage mining software</p>
        </div>
      </div>

      <div class="miners-grid">
        @for (miner of availableMiners(); track miner.type) {
          <div class="miner-card" [class.installed]="miner.installed">
            <div class="miner-icon">
              <svg class="w-10 h-10" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
              </svg>
            </div>

            <div class="miner-info">
              <h3>{{ miner.name }}</h3>
              <p class="miner-description">{{ miner.description }}</p>

              <div class="miner-meta">
                @if (miner.version) {
                  <span class="meta-badge">v{{ miner.version }}</span>
                }
                @if (miner.algorithms.length > 0) {
                  <span class="meta-badge algo">{{ miner.algorithms.join(', ') }}</span>
                }
              </div>
            </div>

            <div class="miner-actions">
              @if (!miner.installed) {
                <button
                  class="btn btn-primary"
                  [disabled]="installing() === miner.type"
                  (click)="installMiner(miner.type)">
                  @if (installing() === miner.type) {
                    <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                      <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                      <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                    </svg>
                    Installing...
                  } @else {
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4"/>
                    </svg>
                    Install
                  }
                </button>
              } @else {
                <div class="installed-badge">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                  Installed
                </div>
                <button class="btn btn-outline" (click)="uninstallMiner(miner.type)">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                  </svg>
                  Uninstall
                </button>
              }
            </div>
          </div>
        }
      </div>

      <!-- System Info -->
      @if (systemInfo()) {
        <div class="system-info-section">
          <h3>System Information</h3>
          <div class="system-info-grid">
            <div class="info-item">
              <span class="info-label">Platform</span>
              <span class="info-value">{{ systemInfo()?.os }} ({{ systemInfo()?.arch }})</span>
            </div>
            <div class="info-item">
              <span class="info-label">CPU</span>
              <span class="info-value">{{ systemInfo()?.cpu_model || 'Unknown' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Cores</span>
              <span class="info-value">{{ systemInfo()?.cpu_cores || 0 }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Memory</span>
              <span class="info-value">{{ formatMemory(systemInfo()?.memory_total || 0) }}</span>
            </div>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .miners-page {
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

    .miners-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(360px, 1fr));
      gap: 1rem;
    }

    .miner-card {
      display: flex;
      gap: 1rem;
      padding: 1.25rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      transition: border-color 0.15s ease;
    }

    .miner-card.installed {
      border-color: rgb(16 185 129 / 0.2);
    }

    .miner-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 56px;
      height: 56px;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.5rem;
      color: var(--color-accent-500);
      flex-shrink: 0;
    }

    .miner-info {
      flex: 1;
      min-width: 0;
    }

    .miner-info h3 {
      font-size: 1rem;
      font-weight: 600;
      color: white;
    }

    .miner-description {
      margin-top: 0.25rem;
      font-size: 0.8125rem;
      color: #64748b;
      line-height: 1.4;
    }

    .miner-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 0.375rem;
      margin-top: 0.75rem;
    }

    .meta-badge {
      padding: 0.125rem 0.5rem;
      background: rgb(37 37 66 / 0.5);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: #94a3b8;
    }

    .meta-badge.algo {
      background: rgb(0 212 255 / 0.1);
      color: var(--color-accent-500);
    }

    .miner-actions {
      display: flex;
      flex-direction: column;
      align-items: flex-end;
      gap: 0.5rem;
      flex-shrink: 0;
    }

    .btn {
      display: inline-flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.5rem 1rem;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.15s ease;
      border: none;
      white-space: nowrap;
    }

    .btn:disabled {
      opacity: 0.7;
      cursor: not-allowed;
    }

    .btn-primary {
      background: var(--color-accent-500);
      color: #0f0f1a;
    }

    .btn-primary:hover:not(:disabled) {
      background: rgb(0 212 255 / 0.8);
    }

    .btn-outline {
      background: transparent;
      border: 1px solid rgb(37 37 66 / 0.3);
      color: #94a3b8;
    }

    .btn-outline:hover {
      background: rgb(37 37 66 / 0.3);
      color: white;
    }

    .installed-badge {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.5rem 0.75rem;
      background: rgb(16 185 129 / 0.1);
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      color: var(--color-success-500);
    }

    .animate-spin {
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    .system-info-section {
      margin-top: 1rem;
      padding: 1.25rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
    }

    .system-info-section h3 {
      font-size: 0.875rem;
      font-weight: 600;
      color: white;
      margin-bottom: 1rem;
    }

    .system-info-grid {
      display: grid;
      grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
      gap: 1rem;
    }

    .info-item {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .info-label {
      font-size: 0.6875rem;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .info-value {
      font-size: 0.875rem;
      color: #e2e8f0;
      font-family: var(--font-family-mono);
    }
  `]
})
export class MinersComponent {
  private minerService = inject(MinerService);
  state = this.minerService.state;

  installing = signal<string | null>(null);

  availableMiners = () => this.state().manageableMiners.map((m: any) => ({
    type: m.name,
    name: m.name,
    description: m.description || this.getMinerDescription(m.name),
    version: this.getInstalledVersion(m.name),
    installed: m.is_installed,
    algorithms: this.getMinerAlgorithms(m.name)
  }));

  getInstalledVersion(type: string): string {
    const installed = this.state().installedMiners.find(m => m.type === type);
    return installed?.version || '';
  }

  systemInfo = () => this.state().systemInfo;

  getMinerDescription(type: string): string {
    const descriptions: Record<string, string> = {
      'xmrig': 'High-performance RandomX and CryptoNight miner',
      'ttminer': 'NVIDIA GPU miner with broad algorithm support',
      'lolminer': 'Multi-algorithm AMD & NVIDIA miner',
      'trex': 'NVIDIA-focused miner for modern GPUs'
    };
    return descriptions[type] || 'Mining software';
  }

  getMinerAlgorithms(type: string): string[] {
    const algorithms: Record<string, string[]> = {
      'xmrig': ['RandomX', 'CryptoNight'],
      'ttminer': ['Ethash', 'KawPow', 'Autolykos2'],
      'lolminer': ['Ethash', 'Beam', 'Equihash'],
      'trex': ['Ethash', 'KawPow', 'Octopus']
    };
    return algorithms[type] || [];
  }

  installMiner(type: string) {
    this.installing.set(type);
    this.minerService.installMiner(type).subscribe({
      next: () => this.installing.set(null),
      error: (err) => {
        console.error('Failed to install miner:', err);
        this.installing.set(null);
      }
    });
  }

  uninstallMiner(type: string) {
    if (confirm(`Are you sure you want to uninstall ${type}?`)) {
      this.minerService.uninstallMiner(type).subscribe({
        error: (err) => console.error('Failed to uninstall miner:', err)
      });
    }
  }

  formatMemory(bytes: number): string {
    const gb = bytes / (1024 * 1024 * 1024);
    return gb.toFixed(1) + ' GB';
  }
}
