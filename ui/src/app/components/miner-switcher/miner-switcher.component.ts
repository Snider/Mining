import { Component, inject, computed, signal, output } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';

@Component({
  selector: 'app-miner-switcher',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="miner-switcher" [class.open]="dropdownOpen()">
      <!-- Current Selection Button -->
      <button class="switcher-btn" (click)="toggleDropdown()">
        <div class="switcher-content">
          @if (viewMode() === 'all') {
            <svg class="switcher-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/>
            </svg>
            <span class="switcher-label">All Workers</span>
            <span class="switcher-count">({{ minerCount() }})</span>
          } @else {
            <div class="miner-status-dot" [class.online]="isSelectedMinerOnline()"></div>
            <span class="switcher-label">{{ selectedMinerName() }}</span>
          }
        </div>
        <svg class="dropdown-arrow" [class.rotated]="dropdownOpen()" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7"/>
        </svg>
      </button>

      <!-- Dropdown Menu -->
      @if (dropdownOpen()) {
        <div class="dropdown-menu">
          <!-- All Workers Option -->
          <button
            class="dropdown-item all-workers"
            [class.active]="viewMode() === 'all'"
            (click)="selectAll()">
            <svg class="item-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z"/>
            </svg>
            <span>All Workers</span>
            <span class="item-count">{{ minerCount() }}</span>
          </button>

          <div class="dropdown-divider"></div>

          <!-- Individual Miners -->
          @for (miner of runningMiners(); track miner.name) {
            <div class="dropdown-item miner-item" [class.active]="selectedMinerName() === miner.name">
              <button class="miner-select" (click)="selectMiner(miner.name)">
                <div class="miner-status-dot online"></div>
                <span class="miner-name">{{ miner.name }}</span>
                <span class="miner-hashrate">{{ formatHashrate(getHashrate(miner)) }}</span>
              </button>
              <div class="miner-actions">
                <button
                  class="action-btn stop"
                  title="Stop miner"
                  (click)="stopMiner($event, miner.name)">
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
                  </svg>
                </button>
                <button
                  class="action-btn edit"
                  title="Edit configuration"
                  (click)="editMiner($event, miner.name)">
                  <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                  </svg>
                </button>
              </div>
            </div>
          }

          @if (runningMiners().length === 0) {
            <div class="dropdown-empty">
              <p>No active workers</p>
            </div>
          }

          <div class="dropdown-divider"></div>

          <!-- Start New Miner -->
          @if (profiles().length > 0) {
            <div class="start-section">
              <span class="section-label">Start Worker</span>
              @for (profile of profiles(); track profile.id) {
                <button class="dropdown-item start-item" (click)="startProfile(profile.id, profile.name)">
                  <svg class="item-icon play" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                  </svg>
                  <span>{{ profile.name }}</span>
                  <span class="profile-type">{{ profile.minerType }}</span>
                </button>
              }
            </div>
          }
        </div>
      }
    </div>

    <!-- Backdrop to close dropdown -->
    @if (dropdownOpen()) {
      <div class="backdrop" (click)="closeDropdown()"></div>
    }
  `,
  styles: [`
    .miner-switcher {
      position: relative;
    }

    .switcher-btn {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.375rem 0.625rem;
      background: var(--color-surface-200);
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      color: white;
      cursor: pointer;
      transition: all 0.15s ease;
      min-width: 140px;
    }

    .switcher-btn:hover {
      background: rgb(37 37 66 / 0.5);
      border-color: var(--color-accent-500);
    }

    .miner-switcher.open .switcher-btn {
      border-color: var(--color-accent-500);
    }

    .switcher-content {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      flex: 1;
    }

    .switcher-icon {
      width: 16px;
      height: 16px;
      color: var(--color-accent-500);
    }

    .switcher-label {
      font-size: 0.8125rem;
      font-weight: 500;
    }

    .switcher-count {
      font-size: 0.75rem;
      color: #64748b;
    }

    .dropdown-arrow {
      width: 14px;
      height: 14px;
      color: #64748b;
      transition: transform 0.2s ease;
    }

    .dropdown-arrow.rotated {
      transform: rotate(180deg);
    }

    .miner-status-dot {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      background: #64748b;
    }

    .miner-status-dot.online {
      background: var(--color-success-500);
      box-shadow: 0 0 6px var(--color-success-500);
    }

    .dropdown-menu {
      position: absolute;
      top: calc(100% + 4px);
      left: 0;
      right: 0;
      min-width: 260px;
      background: var(--color-surface-100);
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.5rem;
      box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.4);
      z-index: 100;
      overflow: hidden;
    }

    .dropdown-item {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      width: 100%;
      padding: 0.5rem 0.75rem;
      background: transparent;
      border: none;
      color: #94a3b8;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: all 0.15s ease;
      text-align: left;
    }

    .dropdown-item:hover {
      background: rgb(37 37 66 / 0.3);
      color: white;
    }

    .dropdown-item.active {
      background: rgb(0 212 255 / 0.1);
      color: var(--color-accent-400);
    }

    .dropdown-item.all-workers {
      padding: 0.625rem 0.75rem;
    }

    .item-icon {
      width: 16px;
      height: 16px;
      flex-shrink: 0;
    }

    .item-icon.play {
      color: var(--color-success-500);
    }

    .item-count {
      margin-left: auto;
      font-size: 0.75rem;
      color: #64748b;
    }

    .dropdown-divider {
      height: 1px;
      background: rgb(37 37 66 / 0.3);
      margin: 0.25rem 0;
    }

    .miner-item {
      padding: 0.375rem 0.5rem 0.375rem 0.75rem;
    }

    .miner-select {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      flex: 1;
      background: none;
      border: none;
      color: inherit;
      cursor: pointer;
      padding: 0.25rem 0;
    }

    .miner-name {
      flex: 1;
      text-align: left;
      font-weight: 500;
    }

    .miner-hashrate {
      font-size: 0.75rem;
      color: #64748b;
      font-family: var(--font-family-mono);
    }

    .miner-actions {
      display: flex;
      gap: 0.25rem;
      opacity: 0;
      transition: opacity 0.15s ease;
    }

    .miner-item:hover .miner-actions {
      opacity: 1;
    }

    .action-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 24px;
      height: 24px;
      background: transparent;
      border: none;
      border-radius: 0.25rem;
      color: #64748b;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .action-btn:hover {
      background: rgb(37 37 66 / 0.5);
    }

    .action-btn.stop:hover {
      color: var(--color-danger-500);
    }

    .action-btn.edit:hover {
      color: var(--color-accent-500);
    }

    .action-btn svg {
      width: 14px;
      height: 14px;
    }

    .dropdown-empty {
      padding: 1rem;
      text-align: center;
      color: #64748b;
      font-size: 0.8125rem;
    }

    .start-section {
      padding: 0.25rem 0;
    }

    .section-label {
      display: block;
      padding: 0.375rem 0.75rem;
      font-size: 0.6875rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #64748b;
    }

    .start-item {
      padding-left: 0.75rem;
    }

    .profile-type {
      margin-left: auto;
      font-size: 0.6875rem;
      padding: 0.125rem 0.375rem;
      background: rgb(37 37 66 / 0.5);
      border-radius: 0.25rem;
      color: #64748b;
    }

    .backdrop {
      position: fixed;
      inset: 0;
      z-index: 99;
    }
  `]
})
export class MinerSwitcherComponent {
  private minerService = inject(MinerService);

  // Output for edit action (navigate to profiles page)
  editProfile = output<string>();

  dropdownOpen = signal(false);

  viewMode = this.minerService.viewMode;
  selectedMinerName = this.minerService.selectedMinerName;
  runningMiners = this.minerService.runningMiners;
  profiles = this.minerService.profiles;

  minerCount = computed(() => this.runningMiners().length);

  isSelectedMinerOnline = computed(() => {
    const name = this.selectedMinerName();
    if (!name) return false;
    return this.runningMiners().some(m => m.name === name);
  });

  toggleDropdown() {
    this.dropdownOpen.update(v => !v);
  }

  closeDropdown() {
    this.dropdownOpen.set(false);
  }

  selectAll() {
    this.minerService.selectAllMiners();
    this.closeDropdown();
  }

  selectMiner(name: string) {
    this.minerService.selectMiner(name);
    this.closeDropdown();
  }

  stopMiner(event: Event, name: string) {
    event.stopPropagation();
    this.minerService.stopMiner(name).subscribe({
      next: () => {
        // If this was the selected miner, switch to all view
        if (this.selectedMinerName() === name) {
          this.minerService.selectAllMiners();
        }
      }
    });
  }

  editMiner(event: Event, name: string) {
    event.stopPropagation();
    // Find the profile for this miner and emit it
    const profile = this.minerService.getProfileForMiner(name);
    if (profile) {
      this.editProfile.emit(profile.id);
    }
    this.closeDropdown();
  }

  startProfile(profileId: string, profileName: string) {
    this.minerService.startMiner(profileId).subscribe();
    this.closeDropdown();
  }

  getHashrate(miner: any): number {
    return miner.full_stats?.hashrate?.total?.[0] || 0;
  }

  formatHashrate(hashrate: number): string {
    if (hashrate >= 1000000) return (hashrate / 1000000).toFixed(1) + ' MH/s';
    if (hashrate >= 1000) return (hashrate / 1000).toFixed(1) + ' kH/s';
    return hashrate.toFixed(0) + ' H/s';
  }
}
