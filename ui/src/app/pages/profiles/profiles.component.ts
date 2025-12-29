import { Component, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';
import { ProfileCreateComponent } from '../../profile-create.component';

@Component({
  selector: 'app-profiles',
  standalone: true,
  imports: [CommonModule, ProfileCreateComponent],
  template: `
    <div class="profiles-page">
      <div class="page-header">
        <div>
          <h2>Mining Profiles</h2>
          <p>Manage your mining configurations</p>
        </div>
        <button class="btn btn-primary" (click)="showCreateForm.set(true)">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
          </svg>
          New Profile
        </button>
      </div>

      @if (showCreateForm()) {
        <div class="create-form-container">
          <snider-mining-profile-create></snider-mining-profile-create>
        </div>
      }

      @if (profiles().length > 0) {
        <div class="profiles-grid">
          @for (profile of profiles(); track profile.id) {
            <div class="profile-card" [class.active]="isRunning(profile.id)">
              <div class="profile-header">
                <div class="profile-info">
                  <h3>{{ profile.name }}</h3>
                  <span class="profile-miner">{{ profile.minerType }}</span>
                </div>
                @if (isRunning(profile.id)) {
                  <span class="running-badge">
                    <div class="pulse-dot"></div>
                    Running
                  </span>
                }
              </div>

              <div class="profile-details">
                <div class="detail-row">
                  <span class="detail-label">Pool</span>
                  <span class="detail-value">{{ profile.config?.pool || 'Not set' }}</span>
                </div>
                <div class="detail-row">
                  <span class="detail-label">Wallet</span>
                  <span class="detail-value">{{ profile.config?.wallet || 'Not set' }}</span>
                </div>
              </div>

              <div class="profile-actions">
                @if (!isRunning(profile.id)) {
                  <button class="action-btn start" (click)="startProfile(profile.id)">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                    </svg>
                    Start
                  </button>
                } @else {
                  <button class="action-btn stop" (click)="stopProfile(profile.id)">
                    <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
                    </svg>
                    Stop
                  </button>
                }
                <button class="action-btn delete" (click)="deleteProfile(profile.id)" [disabled]="isRunning(profile.id)">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
                  </svg>
                </button>
              </div>
            </div>
          }
        </div>
      } @else if (!showCreateForm()) {
        <div class="empty-state">
          <svg class="w-16 h-16 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10"/>
          </svg>
          <h3>No Profiles Yet</h3>
          <p>Create a profile to save your mining configuration.</p>
          <button class="btn btn-primary mt-4" (click)="showCreateForm.set(true)">
            Create Your First Profile
          </button>
        </div>
      }
    </div>
  `,
  styles: [`
    .profiles-page {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }

    .page-header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
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

    .btn-primary {
      background: var(--color-accent-500);
      color: #0f0f1a;
    }

    .btn-primary:hover {
      background: rgb(0 212 255 / 0.8);
    }

    .create-form-container {
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      padding: 1.5rem;
    }

    .profiles-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 1rem;
    }

    .profile-card {
      padding: 1.25rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      transition: border-color 0.15s ease;
    }

    .profile-card.active {
      border-color: rgb(16 185 129 / 0.3);
    }

    .profile-header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      margin-bottom: 1rem;
    }

    .profile-info h3 {
      font-size: 1rem;
      font-weight: 600;
      color: white;
    }

    .profile-miner {
      display: inline-block;
      margin-top: 0.25rem;
      padding: 0.125rem 0.375rem;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: var(--color-accent-500);
      text-transform: uppercase;
    }

    .running-badge {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.25rem 0.5rem;
      background: rgb(16 185 129 / 0.1);
      border-radius: 0.25rem;
      font-size: 0.75rem;
      color: var(--color-success-500);
    }

    .pulse-dot {
      width: 6px;
      height: 6px;
      background: var(--color-success-500);
      border-radius: 50%;
      animation: pulse 2s infinite;
    }

    @keyframes pulse {
      0%, 100% { opacity: 1; }
      50% { opacity: 0.4; }
    }

    .profile-details {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
      padding: 0.75rem 0;
      border-top: 1px solid rgb(37 37 66 / 0.2);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
      margin-bottom: 1rem;
    }

    .detail-row {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 1rem;
    }

    .detail-label {
      font-size: 0.75rem;
      color: #64748b;
    }

    .detail-value {
      font-size: 0.8125rem;
      color: #e2e8f0;
      font-family: var(--font-family-mono);
      text-align: right;
      max-width: 200px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }

    .profile-actions {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .action-btn {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.5rem 0.75rem;
      background: transparent;
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .action-btn.start {
      color: var(--color-success-500);
      border-color: rgb(16 185 129 / 0.3);
    }

    .action-btn.start:hover {
      background: rgb(16 185 129 / 0.1);
    }

    .action-btn.stop {
      color: var(--color-warning-500);
      border-color: rgb(245 158 11 / 0.3);
    }

    .action-btn.stop:hover {
      background: rgb(245 158 11 / 0.1);
    }

    .action-btn.delete {
      color: var(--color-danger-500);
      border-color: rgb(239 68 68 / 0.3);
      margin-left: auto;
    }

    .action-btn.delete:hover:not(:disabled) {
      background: rgb(239 68 68 / 0.1);
    }

    .action-btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
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

    .mt-4 {
      margin-top: 1rem;
    }
  `]
})
export class ProfilesComponent {
  private minerService = inject(MinerService);
  state = this.minerService.state;

  showCreateForm = signal(false);

  profiles = () => this.state().profiles;

  isRunning(profileId: string): boolean {
    return this.state().runningMiners.some(m => m.profile_id === profileId);
  }

  startProfile(profileId: string) {
    this.minerService.startMiner(profileId).subscribe({
      error: (err) => console.error('Failed to start profile:', err)
    });
  }

  stopProfile(profileId: string) {
    const miner = this.state().runningMiners.find(m => m.profile_id === profileId);
    if (miner) {
      this.minerService.stopMiner(miner.name).subscribe({
        error: (err) => console.error('Failed to stop miner:', err)
      });
    }
  }

  deleteProfile(profileId: string) {
    if (confirm('Are you sure you want to delete this profile?')) {
      this.minerService.deleteProfile(profileId).subscribe({
        error: (err) => console.error('Failed to delete profile:', err)
      });
    }
  }

  onProfileCreated() {
    this.showCreateForm.set(false);
  }
}
