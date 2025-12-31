import { Component, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';
import { NotificationService } from '../../notification.service';

interface MinerInfo {
  type: string;
  name: string;
  description: string;
  version: string;
  installed: boolean;
  algorithms: string[];
  recommended: boolean;
  homepage: string;
  license: string;
  placeholder?: boolean;  // True for miners not yet supported by backend
}

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

      <!-- Installed/Recommended Miners -->
      @if (featuredMiners().length > 0) {
        <div class="section-header">
          <h3>Installed & Recommended</h3>
        </div>
        <div class="featured-miners-grid">
          @for (miner of featuredMiners(); track miner.type) {
            <div class="miner-card featured" [class.installed]="miner.installed">
              <div class="featured-ribbon" [class.recommended]="miner.recommended && !miner.installed">
                @if (miner.installed) {
                  <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/>
                  </svg>
                  Installed
                } @else {
                  <svg class="w-3 h-3" fill="currentColor" viewBox="0 0 24 24">
                    <path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/>
                  </svg>
                  Recommended
                }
              </div>

              <div class="featured-header">
                <div class="miner-icon large">
                  @if (miner.type === 'xmrig') {
                    <svg class="w-8 h-8" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
                    </svg>
                  } @else {
                    <svg class="w-8 h-8" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                    </svg>
                  }
                </div>
                <div class="featured-title">
                  <h3>{{ miner.name }}</h3>
                  @if (miner.version) {
                    <span class="version-badge">v{{ miner.version }}</span>
                  }
                </div>
              </div>

              <p class="miner-description">{{ miner.description }}</p>

              <div class="miner-meta">
                @for (algo of miner.algorithms; track algo) {
                  <span class="meta-badge algo">{{ algo }}</span>
                }
              </div>

              <div class="featured-footer">
                <div class="miner-links">
                  @if (miner.homepage) {
                    <a [href]="miner.homepage" target="_blank" class="link-badge">
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                      </svg>
                      Website
                    </a>
                  }
                  @if (miner.license) {
                    <span class="license-badge">{{ miner.license }}</span>
                  }
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
                    <button class="btn btn-outline-danger" (click)="uninstallMiner(miner.type)">
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                      </svg>
                      Uninstall
                    </button>
                  }
                </div>
              </div>
            </div>
          }
        </div>
      }

      <!-- Other Available Miners -->
      @if (otherMiners().length > 0) {
        <div class="section-header">
          <h3>Other Available Miners</h3>
        </div>
        <div class="miners-grid">
          @for (miner of otherMiners(); track miner.type) {
            <div class="miner-card compact" [class.placeholder]="miner.placeholder">
              <div class="miner-icon">
                <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
                </svg>
              </div>

              <div class="miner-info">
                <div class="miner-name-row">
                  <h3>{{ miner.name }}</h3>
                  @if (miner.placeholder) {
                    <span class="coming-soon-badge">Coming Soon</span>
                  }
                </div>
                <p class="miner-description">{{ miner.description }}</p>
                <div class="miner-meta">
                  @for (algo of miner.algorithms.slice(0, 3); track algo) {
                    <span class="meta-badge algo">{{ algo }}</span>
                  }
                  @if (miner.algorithms.length > 3) {
                    <span class="meta-badge">+{{ miner.algorithms.length - 3 }}</span>
                  }
                  @if (miner.homepage) {
                    <a [href]="miner.homepage" target="_blank" class="meta-badge link">
                      <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14"/>
                      </svg>
                      GitHub
                    </a>
                  }
                </div>
              </div>

              <div class="miner-actions">
                @if (miner.placeholder) {
                  <span class="placeholder-badge">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    </svg>
                    Planned
                  </span>
                } @else {
                  <button
                    class="btn btn-secondary"
                    [disabled]="installing() === miner.type"
                    (click)="installMiner(miner.type)">
                    @if (installing() === miner.type) {
                      <svg class="w-4 h-4 animate-spin" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                      </svg>
                    } @else {
                      <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
                      </svg>
                    }
                    Install
                  </button>
                }
              </div>
            </div>
          }
        </div>
      }

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

    .section-header {
      margin-bottom: 0.5rem;
    }

    .section-header h3 {
      font-size: 0.875rem;
      font-weight: 600;
      color: #94a3b8;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    /* Featured miners grid - larger cards */
    .featured-miners-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 1.25rem;
    }

    /* Regular miners grid - compact cards */
    .miners-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
      gap: 1rem;
    }

    .miner-card {
      display: flex;
      gap: 1rem;
      padding: 1.25rem;
      background: var(--color-surface-100);
      border-radius: 0.75rem;
      border: 1px solid rgb(37 37 66 / 0.3);
      transition: all 0.2s ease;
    }

    .miner-card:hover {
      border-color: rgb(37 37 66 / 0.5);
    }

    /* Featured card styling */
    .miner-card.featured {
      flex-direction: column;
      position: relative;
      padding: 1.5rem;
      background: linear-gradient(135deg, var(--color-surface-100) 0%, rgb(25 25 45) 100%);
      border: 1px solid rgb(0 212 255 / 0.15);
      overflow: hidden;
    }

    .miner-card.featured::before {
      content: '';
      position: absolute;
      top: 0;
      left: 0;
      right: 0;
      height: 2px;
      background: linear-gradient(90deg, var(--color-accent-500), transparent);
    }

    .miner-card.featured.installed {
      border-color: rgb(16 185 129 / 0.3);
    }

    .miner-card.featured.installed::before {
      background: linear-gradient(90deg, var(--color-success-500), transparent);
    }

    .featured-ribbon {
      position: absolute;
      top: 0.75rem;
      right: 0.75rem;
      display: flex;
      align-items: center;
      gap: 0.25rem;
      padding: 0.25rem 0.625rem;
      background: rgb(16 185 129 / 0.15);
      border-radius: 1rem;
      font-size: 0.6875rem;
      font-weight: 600;
      color: var(--color-success-500);
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }

    .featured-ribbon.recommended {
      background: rgb(251 191 36 / 0.15);
      color: #fbbf24;
    }

    .featured-header {
      display: flex;
      align-items: center;
      gap: 1rem;
      margin-bottom: 0.75rem;
    }

    .miner-icon {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 48px;
      height: 48px;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.5rem;
      color: var(--color-accent-500);
      flex-shrink: 0;
    }

    .miner-icon.large {
      width: 56px;
      height: 56px;
      border-radius: 0.75rem;
      background: linear-gradient(135deg, rgb(0 212 255 / 0.15) 0%, rgb(0 212 255 / 0.05) 100%);
    }

    .featured-title {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .featured-title h3 {
      font-size: 1.125rem;
      font-weight: 600;
      color: white;
    }

    .version-badge {
      display: inline-flex;
      padding: 0.125rem 0.5rem;
      background: rgb(37 37 66 / 0.5);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: #94a3b8;
      font-family: var(--font-family-mono);
      width: fit-content;
    }

    .miner-info {
      flex: 1;
      min-width: 0;
    }

    .miner-info h3 {
      font-size: 0.9375rem;
      font-weight: 600;
      color: white;
    }

    .miner-name-row {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      flex-wrap: wrap;
    }

    .coming-soon-badge {
      padding: 0.125rem 0.5rem;
      background: rgb(251 191 36 / 0.15);
      border-radius: 0.25rem;
      font-size: 0.625rem;
      font-weight: 600;
      color: #fbbf24;
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }

    .miner-description {
      margin-top: 0.25rem;
      font-size: 0.8125rem;
      color: #64748b;
      line-height: 1.5;
    }

    .miner-card.featured .miner-description {
      margin-bottom: 1rem;
    }

    .miner-meta {
      display: flex;
      flex-wrap: wrap;
      gap: 0.375rem;
      margin-top: 0.5rem;
    }

    .miner-card.featured .miner-meta {
      margin-top: 0;
      margin-bottom: 1rem;
    }

    .meta-badge {
      padding: 0.1875rem 0.5rem;
      background: rgb(37 37 66 / 0.5);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: #94a3b8;
    }

    .meta-badge.algo {
      background: rgb(0 212 255 / 0.1);
      color: var(--color-accent-500);
    }

    .meta-badge.link {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      text-decoration: none;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .meta-badge.link:hover {
      background: rgb(37 37 66 / 0.8);
      color: white;
    }

    /* Placeholder card styles */
    .miner-card.placeholder {
      opacity: 0.75;
      border-style: dashed;
    }

    .miner-card.placeholder:hover {
      opacity: 0.9;
    }

    .placeholder-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.5rem 0.75rem;
      background: rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      font-size: 0.75rem;
      color: #64748b;
      border: 1px dashed rgb(37 37 66 / 0.5);
    }

    .featured-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-top: auto;
      padding-top: 1rem;
      border-top: 1px solid rgb(37 37 66 / 0.3);
    }

    .miner-links {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .link-badge {
      display: inline-flex;
      align-items: center;
      gap: 0.25rem;
      padding: 0.25rem 0.5rem;
      background: rgb(37 37 66 / 0.3);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: #94a3b8;
      text-decoration: none;
      transition: all 0.15s ease;
    }

    .link-badge:hover {
      background: rgb(37 37 66 / 0.5);
      color: white;
    }

    .license-badge {
      padding: 0.25rem 0.5rem;
      background: rgb(139 92 246 / 0.1);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: #a78bfa;
    }

    .miner-actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      flex-shrink: 0;
    }

    .miner-card.compact .miner-actions {
      flex-direction: column;
      align-items: flex-end;
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

    .btn-secondary {
      background: rgb(37 37 66 / 0.5);
      color: #e2e8f0;
    }

    .btn-secondary:hover:not(:disabled) {
      background: rgb(37 37 66 / 0.8);
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

    .btn-outline-danger {
      background: transparent;
      border: 1px solid rgb(239 68 68 / 0.3);
      color: #f87171;
    }

    .btn-outline-danger:hover {
      background: rgb(239 68 68 / 0.1);
      border-color: rgb(239 68 68 / 0.5);
    }

    .animate-spin {
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    /* Mobile responsive styles */
    @media (max-width: 768px) {
      .featured-miners-grid,
      .miners-grid {
        grid-template-columns: 1fr;
      }

      .miner-card.compact {
        flex-direction: column;
        align-items: flex-start;
      }

      .miner-card.compact .miner-actions {
        flex-direction: row;
        width: 100%;
        margin-top: 0.75rem;
      }

      .miner-card.compact .miner-actions .btn {
        flex: 1;
        justify-content: center;
      }

      .featured-footer {
        flex-direction: column;
        gap: 1rem;
        align-items: stretch;
      }

      .miner-links {
        justify-content: center;
      }

      .featured-footer .miner-actions {
        justify-content: center;
      }
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
  private notifications = inject(NotificationService);
  state = this.minerService.state;

  installing = signal<string | null>(null);

  // Miner metadata with descriptions, algorithms, homepage, and license
  // Priority: XMRig (CPU, open source), TT-Miner (NVIDIA GPU), then others by popularity
  private minerMetadata: Record<string, Partial<MinerInfo>> = {
    // Tier 1: Recommended miners with great API support
    'xmrig': {
      description: 'High-performance CPU/GPU miner for RandomX, KawPow, CryptoNight, and more. The most popular open-source miner with excellent HTTP API and stability.',
      algorithms: ['RandomX', 'KawPow', 'CryptoNight', 'GhostRider'],
      homepage: 'https://github.com/xmrig/xmrig',
      license: 'GPL-3.0',
      recommended: true
    },
    'tt-miner': {
      description: 'High-performance NVIDIA GPU miner with excellent efficiency. Supports Ethash, KawPow, Autolykos2, ProgPow and more algorithms.',
      algorithms: ['Ethash', 'KawPow', 'Autolykos2', 'ProgPow'],
      homepage: 'https://github.com/TrailingStop/TT-Miner-release',
      license: 'Proprietary',
      recommended: true
    },
    // Tier 2: Popular GPU miners with HTTP APIs
    'trex': {
      description: 'NVIDIA-focused miner with excellent HTTP REST API and Web UI. Great performance on Ethash, KawPow, Autolykos2, and Octopus.',
      algorithms: ['Ethash', 'KawPow', 'Octopus', 'Autolykos2', 'Blake3', 'FishHash'],
      homepage: 'https://github.com/trexminer/T-Rex',
      license: 'Proprietary',
      recommended: false
    },
    'lolminer': {
      description: 'Multi-GPU miner supporting AMD, NVIDIA, and Intel Arc. HTTP JSON API with Web GUI. Excellent Equihash and Beam performance.',
      algorithms: ['Ethash', 'Etchash', 'BeamHash', 'Equihash', 'Autolykos2', 'FishHash'],
      homepage: 'https://github.com/Lolliedieb/lolMiner-releases',
      license: 'Proprietary',
      recommended: false
    },
    'rigel': {
      description: 'Modern NVIDIA miner with HTTP REST API. Supports Ethash, KawPow, Autolykos2, FishHash, KarlsenHash, and more.',
      algorithms: ['Ethash', 'Etchash', 'KawPow', 'Autolykos2', 'FishHash', 'KarlsenHash'],
      homepage: 'https://github.com/rigelminer/rigel',
      license: 'Proprietary',
      recommended: false
    },
    'bzminer': {
      description: 'Multi-GPU miner for AMD and NVIDIA with HTTP API and Web GUI. Discord webhook integration for monitoring.',
      algorithms: ['KawPow', 'Ethash', 'Etchash', 'Autolykos2', 'Karlsen', 'Alephium'],
      homepage: 'https://github.com/bzminer/bzminer',
      license: 'Proprietary',
      recommended: false
    },
    'srbminer': {
      description: 'CPU+GPU miner supporting 100+ algorithms. HTTP API with Web GUI. Works with AMD, NVIDIA, and Intel GPUs.',
      algorithms: ['RandomX', 'Ethash', 'Autolykos2', 'KarlsenHash', 'CryptoNight', 'GhostRider'],
      homepage: 'https://github.com/doktor83/SRBMiner-Multi',
      license: 'Proprietary',
      recommended: false
    },
    'teamredminer': {
      description: 'AMD-focused GPU miner with Claymore-compatible API. Excellent performance on Ethash, KawPow, and Autolykos2.',
      algorithms: ['Ethash', 'Etchash', 'KawPow', 'Autolykos2', 'Karlsen', 'CryptoNight'],
      homepage: 'https://github.com/todxx/teamredminer',
      license: 'Proprietary',
      recommended: false
    },
    'gminer': {
      description: 'GPU miner with built-in Web UI. Supports Ethash, Equihash variants, KawPow, and Autolykos2.',
      algorithms: ['Ethash', 'ProgPoW', 'KawPow', 'Equihash', 'Autolykos2', 'Beam'],
      homepage: 'https://github.com/develsoftware/GMinerRelease',
      license: 'Proprietary',
      recommended: false
    },
    'nbminer': {
      description: 'GPU miner with HTTP REST API and Web Monitor. Supports Ethash, KawPow, BeamV3, Octopus, and more.',
      algorithms: ['Ethash', 'Etchash', 'KawPow', 'BeamV3', 'Octopus', 'Autolykos2'],
      homepage: 'https://github.com/NebuTech/NBMiner',
      license: 'Proprietary',
      recommended: false
    }
  };

  // All miners with full metadata (from backend + placeholders)
  private allMiners = computed<MinerInfo[]>(() => {
    // Get miners from backend
    const backendMiners = this.state().manageableMiners.map((m: any) => {
      const meta = this.minerMetadata[m.name] || {};
      return {
        type: m.name,
        name: m.name,
        description: m.description || meta.description || 'Mining software',
        version: this.getInstalledVersion(m.name),
        installed: m.is_installed,
        algorithms: meta.algorithms || [],
        recommended: meta.recommended || false,
        homepage: meta.homepage || '',
        license: meta.license || '',
        placeholder: false
      };
    });

    // Get backend miner names for filtering
    const backendNames = new Set(backendMiners.map(m => m.type));

    // Add placeholder miners that aren't from backend
    const placeholderMiners = Object.entries(this.minerMetadata)
      .filter(([name]) => !backendNames.has(name))
      .map(([name, meta]) => ({
        type: name,
        name: name,
        description: meta.description || 'Mining software',
        version: '',
        installed: false,
        algorithms: meta.algorithms || [],
        recommended: meta.recommended || false,
        homepage: meta.homepage || '',
        license: meta.license || '',
        placeholder: true
      }));

    return [...backendMiners, ...placeholderMiners];
  });

  // Featured miners: installed OR recommended (sorted: installed first)
  featuredMiners = computed<MinerInfo[]>(() => {
    return this.allMiners()
      .filter(m => m.installed || m.recommended)
      .sort((a, b) => {
        // Installed first
        if (a.installed && !b.installed) return -1;
        if (!a.installed && b.installed) return 1;
        // Then by recommended
        if (a.recommended && !b.recommended) return -1;
        if (!a.recommended && b.recommended) return 1;
        return 0;
      });
  });

  // Other miners: not installed AND not recommended
  otherMiners = computed<MinerInfo[]>(() => {
    return this.allMiners().filter(m => !m.installed && !m.recommended);
  });

  getInstalledVersion(type: string): string {
    const installed = this.state().installedMiners.find(m => m.type === type);
    return installed?.version || '';
  }

  systemInfo = () => this.state().systemInfo;

  installMiner(type: string) {
    this.installing.set(type);
    this.minerService.installMiner(type).subscribe({
      next: () => {
        this.installing.set(null);
        this.notifications.success(`${type} installed successfully`, 'Installation Complete');
      },
      error: (err) => {
        console.error('Failed to install miner:', err);
        this.installing.set(null);
        this.notifications.error(`Failed to install ${type}: ${err.message || 'Unknown error'}`, 'Installation Failed');
      }
    });
  }

  uninstallMiner(type: string) {
    if (confirm(`Are you sure you want to uninstall ${type}?`)) {
      this.minerService.uninstallMiner(type).subscribe({
        next: () => {
          this.notifications.success(`${type} uninstalled successfully`, 'Uninstall Complete');
        },
        error: (err) => {
          console.error('Failed to uninstall miner:', err);
          this.notifications.error(`Failed to uninstall ${type}: ${err.message || 'Unknown error'}`, 'Uninstall Failed');
        }
      });
    }
  }

  formatMemory(bytes: number): string {
    const gb = bytes / (1024 * 1024 * 1024);
    return gb.toFixed(1) + ' GB';
  }
}
