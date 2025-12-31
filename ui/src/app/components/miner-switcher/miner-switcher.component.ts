import { Component, inject, computed, signal, output, HostListener, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Subscription } from 'rxjs';
import { MinerService } from '../../miner.service';
import { WebSocketService } from '../../websocket.service';

interface ContextMenuState {
  visible: boolean;
  x: number;
  y: number;
  minerName: string;
}

// Spinner SVG for loading states
const SPINNER_SVG = `<svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
  <circle cx="12" cy="12" r="10" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
</svg>`;

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
            <div class="dropdown-item miner-item"
                 [class.active]="selectedMinerName() === miner.name"
                 [class.stopping]="isLoading('stop-' + miner.name)"
                 (contextmenu)="openContextMenu($event, miner.name)">
              <button class="miner-select" (click)="selectMiner(miner.name)">
                <div class="miner-status-dot online"></div>
                <span class="miner-name">{{ miner.name }}</span>
                <span class="miner-hashrate">{{ formatHashrate(getHashrate(miner)) }}</span>
              </button>
              <div class="miner-actions" [class.show]="isLoading('stop-' + miner.name)">
                @if (isLoading('stop-' + miner.name)) {
                  <div class="stopping-indicator">
                    <svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                      <circle cx="12" cy="12" r="10" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
                    </svg>
                    <span class="stopping-text">Stopping</span>
                  </div>
                  <button
                    class="cancel-btn"
                    title="Cancel"
                    (click)="cancelAction($event, 'stop-' + miner.name)">
                    <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/>
                    </svg>
                  </button>
                } @else {
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
                }
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
                <div class="start-item-wrapper" [class.loading]="isLoading('start-' + profile.id)">
                  <button
                    class="dropdown-item start-item"
                    [class.loading]="isLoading('start-' + profile.id)"
                    [disabled]="isLoading('start-' + profile.id)"
                    (click)="startProfile(profile.id, profile.name)">
                    @if (isLoading('start-' + profile.id)) {
                      <svg class="item-icon spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
                        <circle cx="12" cy="12" r="10" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
                      </svg>
                    } @else {
                      <svg class="item-icon play" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"/>
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
                      </svg>
                    }
                    <span>{{ profile.name }}</span>
                    <span class="profile-type">{{ profile.minerType }}</span>
                  </button>
                  @if (isLoading('start-' + profile.id)) {
                    <button
                      class="cancel-btn"
                      title="Cancel"
                      (click)="cancelAction($event, 'start-' + profile.id)">
                      <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2.5" d="M6 18L18 6M6 6l12 12"/>
                      </svg>
                    </button>
                  }
                </div>
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

    <!-- Context Menu -->
    @if (contextMenu().visible) {
      <div class="context-menu-backdrop" (click)="closeContextMenu()"></div>
      <div class="context-menu"
           [style.left.px]="contextMenu().x"
           [style.top.px]="contextMenu().y">
        <div class="context-menu-header">{{ contextMenu().minerName }}</div>
        <div class="context-menu-divider"></div>
        <button class="context-menu-item" (click)="viewConsole()">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
          </svg>
          <span>View Console</span>
        </button>
        <button class="context-menu-item" (click)="viewStats()">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
          </svg>
          <span>View Stats</span>
        </button>
        <button class="context-menu-item" (click)="viewHashrateHistory()">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 12l3-3 3 3 4-4M8 21l4-4 4 4M3 4h18M4 4h16v12a1 1 0 01-1 1H5a1 1 0 01-1-1V4z"/>
          </svg>
          <span>Hashrate History</span>
        </button>
        <div class="context-menu-divider"></div>
        <button class="context-menu-item" (click)="editFromContext()">
          <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
          </svg>
          <span>Edit Configuration</span>
        </button>
        <div class="context-menu-divider"></div>
        <button
          class="context-menu-item danger"
          [class.loading]="isLoading('stop-' + contextMenu().minerName)"
          [disabled]="isLoading('stop-' + contextMenu().minerName)"
          (click)="stopFromContext()">
          @if (isLoading('stop-' + contextMenu().minerName)) {
            <svg class="spinner" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="12" r="10" stroke-width="3" stroke-dasharray="31.4 31.4" stroke-linecap="round"/>
            </svg>
          } @else {
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 10a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"/>
            </svg>
          }
          <span>Stop Worker</span>
        </button>
      </div>
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
      align-items: center;
      gap: 0.25rem;
      opacity: 0;
      transition: opacity 0.15s ease;
    }

    .miner-item:hover .miner-actions,
    .miner-actions.show {
      opacity: 1;
    }

    /* Stopping indicator */
    .stopping-indicator {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      color: var(--color-accent-400);
    }

    .stopping-indicator .spinner {
      width: 14px;
      height: 14px;
    }

    .stopping-text {
      font-size: 0.6875rem;
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }

    .miner-item.stopping {
      background: rgb(0 212 255 / 0.05);
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

    /* Context Menu Styles */
    .context-menu-backdrop {
      position: fixed;
      inset: 0;
      z-index: 200;
    }

    .context-menu {
      position: fixed;
      min-width: 180px;
      background: var(--color-surface-100);
      border: 1px solid rgb(37 37 66 / 0.5);
      border-radius: 0.5rem;
      box-shadow: 0 10px 40px -5px rgba(0, 0, 0, 0.5);
      z-index: 201;
      overflow: hidden;
      animation: contextMenuIn 0.15s ease;
    }

    @keyframes contextMenuIn {
      from {
        opacity: 0;
        transform: scale(0.95);
      }
      to {
        opacity: 1;
        transform: scale(1);
      }
    }

    .context-menu-header {
      padding: 0.5rem 0.75rem;
      font-size: 0.75rem;
      font-weight: 600;
      color: var(--color-accent-400);
      background: rgb(0 212 255 / 0.05);
      border-bottom: 1px solid rgb(37 37 66 / 0.3);
      text-transform: uppercase;
      letter-spacing: 0.03em;
    }

    .context-menu-divider {
      height: 1px;
      background: rgb(37 37 66 / 0.3);
      margin: 0.25rem 0;
    }

    .context-menu-item {
      display: flex;
      align-items: center;
      gap: 0.625rem;
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

    .context-menu-item:hover {
      background: rgb(37 37 66 / 0.4);
      color: white;
    }

    .context-menu-item.danger:hover {
      background: rgb(239 68 68 / 0.15);
      color: var(--color-danger-400);
    }

    .context-menu-item svg {
      width: 16px;
      height: 16px;
      flex-shrink: 0;
    }

    .context-menu-item.danger svg {
      color: var(--color-danger-500);
    }

    /* Spinner animation */
    .spinner {
      animation: spin 1s linear infinite;
    }

    @keyframes spin {
      from { transform: rotate(0deg); }
      to { transform: rotate(360deg); }
    }

    /* Loading state styles */
    .action-btn.loading,
    .dropdown-item.loading,
    .context-menu-item.loading {
      opacity: 0.7;
      cursor: not-allowed;
      pointer-events: none;
    }

    .action-btn:disabled,
    .dropdown-item:disabled,
    .context-menu-item:disabled {
      cursor: not-allowed;
      pointer-events: none;
    }

    .action-btn.loading .spinner,
    .dropdown-item.loading .spinner,
    .context-menu-item.loading .spinner {
      color: var(--color-accent-400);
    }

    /* Start item wrapper for cancel button positioning */
    .start-item-wrapper {
      position: relative;
      display: flex;
      align-items: center;
    }

    .start-item-wrapper .start-item {
      flex: 1;
    }

    /* Cancel button - appears on hover over loading items */
    .cancel-btn {
      position: absolute;
      right: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      width: 20px;
      height: 20px;
      background: var(--color-danger-500);
      border: none;
      border-radius: 50%;
      color: white;
      cursor: pointer;
      opacity: 0;
      transform: scale(0.8);
      transition: all 0.15s ease;
      z-index: 10;
    }

    .cancel-btn svg {
      width: 12px;
      height: 12px;
    }

    .start-item-wrapper:hover .cancel-btn,
    .miner-item:hover .cancel-btn {
      opacity: 1;
      transform: scale(1);
    }

    .cancel-btn:hover {
      background: var(--color-danger-400);
      transform: scale(1.1);
    }

    .cancel-btn:active {
      transform: scale(0.95);
    }
  `]
})
export class MinerSwitcherComponent implements OnDestroy {
  private minerService = inject(MinerService);
  private ws = inject(WebSocketService);
  private wsSubscriptions: Subscription[] = [];

  // Output for edit action (navigate to profiles page)
  editProfile = output<string>();

  // Output for navigation actions
  navigateToConsole = output<string>();
  navigateToStats = output<string>();

  dropdownOpen = signal(false);
  contextMenu = signal<ContextMenuState>({ visible: false, x: 0, y: 0, minerName: '' });

  // Track loading states for actions (e.g., "stop-minerName", "start-profileId")
  private loadingActions = signal<Set<string>>(new Set());

  // Track pending start actions to match profile IDs with miner names
  private pendingStarts = new Map<string, string>(); // profileId -> expected miner type

  constructor() {
    this.subscribeToWebSocketEvents();
  }

  ngOnDestroy(): void {
    this.wsSubscriptions.forEach(sub => sub.unsubscribe());
  }

  /**
   * Subscribe to WebSocket events to clear loading states when actions complete
   */
  private subscribeToWebSocketEvents(): void {
    // When a miner starts, clear the loading state for its profile
    const startedSub = this.ws.minerStarted$.subscribe(data => {
      console.log('[MinerSwitcher] Miner started event:', data.name);
      // Clear any start loading states that might match this miner
      this.loadingActions.update(set => {
        const newSet = new Set(set);
        // Clear all start-* loading states since we got a miner.started event
        for (const key of newSet) {
          if (key.startsWith('start-')) {
            newSet.delete(key);
          }
        }
        return newSet;
      });
      this.pendingStarts.clear();
    });
    this.wsSubscriptions.push(startedSub);

    // When a miner stops, clear the loading state
    const stoppedSub = this.ws.minerStopped$.subscribe(data => {
      console.log('[MinerSwitcher] Miner stopped event:', data.name);
      const actionKey = `stop-${data.name}`;
      this.setLoading(actionKey, false);

      // If this was the selected miner, switch to all view
      if (this.selectedMinerName() === data.name) {
        this.minerService.selectAllMiners();
      }

      // Close context menu if it was for this miner
      if (this.contextMenu().minerName === data.name) {
        this.closeContextMenu();
      }
    });
    this.wsSubscriptions.push(stoppedSub);

    // On error, clear relevant loading states
    const errorSub = this.ws.minerError$.subscribe(data => {
      console.log('[MinerSwitcher] Miner error event:', data.name, data.error);
      // Clear both start and stop loading states for this miner
      this.loadingActions.update(set => {
        const newSet = new Set(set);
        newSet.delete(`stop-${data.name}`);
        // Also clear any pending starts
        for (const key of newSet) {
          if (key.startsWith('start-')) {
            newSet.delete(key);
          }
        }
        return newSet;
      });
    });
    this.wsSubscriptions.push(errorSub);
  }

  isLoading(actionKey: string): boolean {
    return this.loadingActions().has(actionKey);
  }

  private setLoading(actionKey: string, loading: boolean) {
    this.loadingActions.update(set => {
      const newSet = new Set(set);
      if (loading) {
        newSet.add(actionKey);
      } else {
        newSet.delete(actionKey);
      }
      return newSet;
    });
  }

  // Close context menu on Escape
  @HostListener('document:keydown.escape')
  onEscape() {
    this.closeContextMenu();
  }

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
    const actionKey = `stop-${name}`;
    if (this.isLoading(actionKey)) return;

    this.setLoading(actionKey, true);
    this.minerService.stopMiner(name).subscribe({
      next: () => {
        // Loading state will be cleared by WebSocket miner.stopped event
        // Keep spinner spinning until we get confirmation the miner actually stopped
      },
      error: () => {
        // On HTTP error, clear loading state immediately
        this.setLoading(actionKey, false);
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
    const actionKey = `start-${profileId}`;
    if (this.isLoading(actionKey)) return;

    this.setLoading(actionKey, true);
    this.pendingStarts.set(profileId, profileName);

    this.minerService.startMiner(profileId).subscribe({
      next: () => {
        // Loading state will be cleared by WebSocket miner.started event
        // Keep spinner spinning until we get confirmation the miner actually started
        this.closeDropdown();
      },
      error: () => {
        // On HTTP error, clear loading state immediately
        this.setLoading(actionKey, false);
        this.pendingStarts.delete(profileId);
      }
    });
  }

  /**
   * Cancel an in-progress action (clears loading state)
   * For start actions, the miner may still start - this just clears the UI state
   */
  cancelAction(event: Event, actionKey: string) {
    event.stopPropagation();
    event.preventDefault();

    // Clear the loading state
    this.setLoading(actionKey, false);

    // If it was a start action, clear pending starts
    if (actionKey.startsWith('start-')) {
      const profileId = actionKey.replace('start-', '');
      this.pendingStarts.delete(profileId);
    }
  }

  getHashrate(miner: any): number {
    return miner.full_stats?.hashrate?.total?.[0] || 0;
  }

  formatHashrate(hashrate: number): string {
    if (hashrate >= 1000000) return (hashrate / 1000000).toFixed(1) + ' MH/s';
    if (hashrate >= 1000) return (hashrate / 1000).toFixed(1) + ' kH/s';
    return hashrate.toFixed(0) + ' H/s';
  }

  // Context Menu Methods
  openContextMenu(event: MouseEvent, minerName: string) {
    event.preventDefault();
    event.stopPropagation();
    this.contextMenu.set({
      visible: true,
      x: event.clientX,
      y: event.clientY,
      minerName
    });
  }

  closeContextMenu() {
    this.contextMenu.set({ visible: false, x: 0, y: 0, minerName: '' });
  }

  viewConsole() {
    const minerName = this.contextMenu().minerName;
    this.minerService.selectMiner(minerName);
    this.navigateToConsole.emit(minerName);
    this.closeContextMenu();
    this.closeDropdown();
  }

  viewStats() {
    const minerName = this.contextMenu().minerName;
    this.minerService.selectMiner(minerName);
    this.navigateToStats.emit(minerName);
    this.closeContextMenu();
    this.closeDropdown();
  }

  viewHashrateHistory() {
    const minerName = this.contextMenu().minerName;
    this.minerService.selectMiner(minerName);
    // Navigate to dashboard which shows hashrate history
    this.navigateToStats.emit(minerName);
    this.closeContextMenu();
    this.closeDropdown();
  }

  editFromContext() {
    const minerName = this.contextMenu().minerName;
    const profile = this.minerService.getProfileForMiner(minerName);
    if (profile) {
      this.editProfile.emit(profile.id);
    }
    this.closeContextMenu();
    this.closeDropdown();
  }

  stopFromContext() {
    const minerName = this.contextMenu().minerName;
    const actionKey = `stop-${minerName}`;
    if (this.isLoading(actionKey)) return;

    this.setLoading(actionKey, true);
    this.minerService.stopMiner(minerName).subscribe({
      next: () => {
        // Loading state and context menu will be cleared by WebSocket miner.stopped event
        // Keep spinner spinning until we get confirmation the miner actually stopped
      },
      error: () => {
        // On HTTP error, clear loading state immediately
        this.setLoading(actionKey, false);
        this.closeContextMenu();
      }
    });
  }
}
