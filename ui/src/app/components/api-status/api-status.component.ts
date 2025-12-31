import { Component, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';

@Component({
  selector: 'app-api-status',
  standalone: true,
  imports: [CommonModule],
  template: `
    @if (!minerService.apiAvailable()) {
      <div class="api-error-banner">
        <div class="banner-content">
          <svg class="banner-icon" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path fill-rule="evenodd" d="M9.401 3.003c1.155-2 4.043-2 5.197 0l7.355 12.748c1.154 2-.29 4.5-2.599 4.5H4.645c-2.309 0-3.752-2.5-2.598-4.5L9.4 3.003zM12 8.25a.75.75 0 01.75.75v3.75a.75.75 0 01-1.5 0V9a.75.75 0 01.75-.75zm0 8.25a.75.75 0 100-1.5.75.75 0 000 1.5z" clip-rule="evenodd" />
          </svg>
          <div class="banner-text">
            <span class="banner-title">Connection Error</span>
            <span class="banner-message">Unable to connect to the mining API. Make sure the backend is running.</span>
          </div>
          <button class="retry-btn" (click)="retry()">
            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
            </svg>
            Retry
          </button>
        </div>
      </div>
    }
  `,
  styles: [`
    .api-error-banner {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      z-index: 9998;
      background: linear-gradient(135deg, #991b1b 0%, #7f1d1d 100%);
      border-bottom: 1px solid rgb(239 68 68 / 0.3);
      padding: 0.75rem 1rem;
    }

    .banner-content {
      display: flex;
      align-items: center;
      gap: 1rem;
      max-width: 1200px;
      margin: 0 auto;
    }

    .banner-icon {
      width: 1.5rem;
      height: 1.5rem;
      color: #fca5a5;
      flex-shrink: 0;
    }

    .banner-text {
      flex: 1;
      display: flex;
      flex-direction: column;
      gap: 0.125rem;
    }

    .banner-title {
      font-weight: 600;
      font-size: 0.875rem;
      color: #fef2f2;
    }

    .banner-message {
      font-size: 0.8125rem;
      color: #fecaca;
    }

    .retry-btn {
      display: inline-flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.5rem 1rem;
      background: rgba(255, 255, 255, 0.1);
      border: 1px solid rgba(255, 255, 255, 0.2);
      border-radius: 0.375rem;
      color: white;
      font-size: 0.8125rem;
      font-weight: 500;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .retry-btn:hover {
      background: rgba(255, 255, 255, 0.2);
    }

    .retry-btn svg {
      width: 1rem;
      height: 1rem;
    }
  `]
})
export class ApiStatusComponent {
  minerService = inject(MinerService);

  retry() {
    this.minerService.forceRefreshState();
  }
}
