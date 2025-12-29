import { Component, inject, signal, OnInit, OnDestroy, WritableSignal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NodeService, Peer, NodeIdentity } from '../../node.service';

@Component({
  selector: 'app-nodes',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="nodes-page">
      <!-- Local Node Identity Section -->
      <section class="node-identity-section">
        <h2 class="section-title">
          <svg class="section-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 3v2m6-2v2M9 19v2m6-2v2M5 9H3m2 6H3m18-6h-2m2 6h-2M7 19h10a2 2 0 002-2V7a2 2 0 00-2-2H7a2 2 0 00-2 2v10a2 2 0 002 2zM9 9h6v6H9V9z"/>
          </svg>
          Local Node
        </h2>

        @if (nodeService.initialized()) {
          <div class="identity-card">
            <div class="identity-info">
              <div class="identity-name">
                <div class="status-dot online"></div>
                <span>{{ nodeService.identity()?.name }}</span>
                <span class="role-badge" [class]="nodeService.identity()?.role">
                  {{ nodeService.identity()?.role }}
                </span>
              </div>
              <div class="identity-id">
                <span class="label">Node ID:</span>
                <code>{{ nodeService.identity()?.id }}</code>
                <button class="copy-btn" (click)="copyToClipboard(nodeService.identity()?.id || '')" title="Copy ID">
                  <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"/>
                  </svg>
                </button>
              </div>
            </div>
            <div class="identity-stats">
              <div class="stat">
                <span class="stat-value">{{ nodeService.peers().length }}</span>
                <span class="stat-label">Peers</span>
              </div>
              <div class="stat">
                <span class="stat-value">{{ nodeService.onlinePeers().length }}</span>
                <span class="stat-label">Online</span>
              </div>
            </div>
          </div>
        } @else {
          <div class="init-card">
            <h3>Initialize Node Identity</h3>
            <p>Set up your node to enable P2P communication with remote mining rigs.</p>
            <form class="init-form" (ngSubmit)="initializeNode()">
              <div class="form-group">
                <label for="nodeName">Node Name</label>
                <input
                  id="nodeName"
                  type="text"
                  [(ngModel)]="newNodeName"
                  name="nodeName"
                  placeholder="e.g., control-center"
                  required>
              </div>
              <div class="form-group">
                <label for="nodeRole">Role</label>
                <select id="nodeRole" [(ngModel)]="newNodeRole" name="nodeRole">
                  <option value="dual">Dual (Controller + Worker)</option>
                  <option value="controller">Controller Only</option>
                  <option value="worker">Worker Only</option>
                </select>
              </div>
              <button type="submit" class="btn btn-primary" [disabled]="!newNodeName || actionInProgress() === 'init-node'">
                @if (actionInProgress() === 'init-node') {
                  <div class="spinner-sm"></div>
                  Initializing...
                } @else {
                  Initialize Node
                }
              </button>
            </form>
          </div>
        }
      </section>

      <!-- Peers Section -->
      @if (nodeService.initialized()) {
        <section class="peers-section">
          <div class="section-header">
            <h2 class="section-title">
              <svg class="section-icon" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
              </svg>
              Connected Peers
            </h2>
            <button class="btn btn-secondary" (click)="showAddPeerModal = true" [disabled]="actionInProgress() !== null">
              <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
              </svg>
              Add Peer
            </button>
          </div>

          @if (nodeService.peers().length > 0) {
            <div class="peers-table-container">
              <table class="peers-table">
                <thead>
                  <tr>
                    <th>Peer</th>
                    <th>Address</th>
                    <th class="text-center">Role</th>
                    <th class="text-right">Ping</th>
                    <th class="text-right">Score</th>
                    <th class="text-center">Last Seen</th>
                    <th class="text-center">Actions</th>
                  </tr>
                </thead>
                <tbody>
                  @for (peer of nodeService.peers(); track peer.id) {
                    <tr>
                      <td>
                        <div class="peer-name">
                          <div class="status-dot" [class.online]="peer.status === 'online'" [class.offline]="peer.status !== 'online'"></div>
                          <span>{{ peer.name }}</span>
                        </div>
                      </td>
                      <td>
                        <code class="address-code">{{ peer.address }}</code>
                      </td>
                      <td class="text-center">
                        <span class="role-badge" [class]="peer.role">{{ peer.role }}</span>
                      </td>
                      <td class="text-right tabular-nums">
                        @if (peer.pingMs > 0) {
                          <span [class.text-success-500]="peer.pingMs < 50"
                                [class.text-warning-500]="peer.pingMs >= 50 && peer.pingMs < 200"
                                [class.text-danger-500]="peer.pingMs >= 200">
                            {{ peer.pingMs.toFixed(0) }}ms
                          </span>
                        } @else {
                          <span class="text-muted">-</span>
                        }
                      </td>
                      <td class="text-right tabular-nums">
                        <span [class.text-success-500]="peer.score >= 90"
                              [class.text-warning-500]="peer.score >= 50 && peer.score < 90"
                              [class.text-danger-500]="peer.score < 50">
                          {{ peer.score.toFixed(0) }}
                        </span>
                      </td>
                      <td class="text-center">
                        <span class="text-muted">{{ formatLastSeen(peer.lastSeen) }}</span>
                      </td>
                      <td class="actions-cell">
                        <button class="icon-btn" title="Ping" (click)="pingPeer(peer.id)" [disabled]="actionInProgress() === 'ping-' + peer.id">
                          @if (actionInProgress() === 'ping-' + peer.id) {
                            <div class="spinner-sm"></div>
                          } @else {
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8.111 16.404a5.5 5.5 0 017.778 0M12 20h.01m-7.08-7.071c3.904-3.905 10.236-3.905 14.142 0M1.394 9.393c5.857-5.857 15.355-5.857 21.213 0"/>
                            </svg>
                          }
                        </button>
                        <button class="icon-btn" title="View Stats" (click)="viewPeerStats(peer)" [disabled]="actionInProgress() !== null">
                          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"/>
                          </svg>
                        </button>
                        <button class="icon-btn icon-btn-danger" title="Remove" (click)="removePeer(peer.id)" [disabled]="actionInProgress() === 'remove-' + peer.id">
                          @if (actionInProgress() === 'remove-' + peer.id) {
                            <div class="spinner-sm"></div>
                          } @else {
                            <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
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
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0z"/>
              </svg>
              <h3>No Peers Connected</h3>
              <p>Add peers to manage remote mining rigs from this dashboard.</p>
              <button class="btn btn-primary" (click)="showAddPeerModal = true" [disabled]="actionInProgress() !== null">
                <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6v6m0 0v6m0-6h6m-6 0H6"/>
                </svg>
                Add First Peer
              </button>
            </div>
          }
        </section>
      }

      <!-- Add Peer Modal -->
      @if (showAddPeerModal) {
        <div class="modal-overlay" (click)="showAddPeerModal = false">
          <div class="modal" (click)="$event.stopPropagation()">
            <div class="modal-header">
              <h3>Add Peer</h3>
              <button class="close-btn" (click)="showAddPeerModal = false">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <form class="modal-body" (ngSubmit)="addPeer()">
              <div class="form-group">
                <label for="peerAddress">Peer Address</label>
                <input
                  id="peerAddress"
                  type="text"
                  [(ngModel)]="newPeerAddress"
                  name="peerAddress"
                  placeholder="e.g., 192.168.1.100:9091"
                  required>
                <span class="hint">Enter the IP address and port of the remote node</span>
              </div>
              <div class="form-group">
                <label for="peerName">Peer Name (optional)</label>
                <input
                  id="peerName"
                  type="text"
                  [(ngModel)]="newPeerName"
                  name="peerName"
                  placeholder="e.g., rig-alpha">
              </div>
              <div class="modal-actions">
                <button type="button" class="btn btn-secondary" (click)="showAddPeerModal = false" [disabled]="actionInProgress() === 'add-peer'">Cancel</button>
                <button type="submit" class="btn btn-primary" [disabled]="!newPeerAddress || actionInProgress() === 'add-peer'">
                  @if (actionInProgress() === 'add-peer') {
                    <div class="spinner-sm"></div>
                    Adding...
                  } @else {
                    Add Peer
                  }
                </button>
              </div>
            </form>
          </div>
        </div>
      }

      <!-- Peer Stats Modal -->
      @if (selectedPeer) {
        <div class="modal-overlay" (click)="selectedPeer = null">
          <div class="modal modal-wide" (click)="$event.stopPropagation()">
            <div class="modal-header">
              <h3>{{ selectedPeer.name }} - Stats</h3>
              <button class="close-btn" (click)="selectedPeer = null">
                <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="modal-body">
              @if (selectedPeerStats) {
                <div class="peer-stats-grid">
                  @for (miner of selectedPeerStats.miners; track miner.name) {
                    <div class="miner-stat-card">
                      <div class="miner-header">
                        <div class="status-dot online"></div>
                        <span class="miner-name">{{ miner.name }}</span>
                        <span class="algo-badge">{{ miner.algorithm }}</span>
                      </div>
                      <div class="miner-stats">
                        <div class="stat-item">
                          <span class="stat-value">{{ formatHashrate(miner.hashrate) }}</span>
                          <span class="stat-label">Hashrate</span>
                        </div>
                        <div class="stat-item">
                          <span class="stat-value">{{ miner.shares }}</span>
                          <span class="stat-label">Shares</span>
                        </div>
                        <div class="stat-item">
                          <span class="stat-value">{{ formatUptime(miner.uptime) }}</span>
                          <span class="stat-label">Uptime</span>
                        </div>
                      </div>
                    </div>
                  }
                </div>
              } @else {
                <div class="loading-stats">
                  <div class="spinner"></div>
                  <span>Loading stats...</span>
                </div>
              }
            </div>
          </div>
        </div>
      }
    </div>
  `,
  styles: [`
    .nodes-page {
      display: flex;
      flex-direction: column;
      gap: 1.5rem;
    }

    .section-title {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 1rem;
      font-weight: 600;
      color: white;
      margin: 0 0 1rem 0;
    }

    .section-icon {
      width: 20px;
      height: 20px;
      flex-shrink: 0;
    }

    .w-4 { width: 16px; }
    .h-4 { height: 16px; }
    .w-5 { width: 20px; }
    .h-5 { height: 20px; }

    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 1rem;
    }

    .section-header .section-title {
      margin: 0;
    }

    /* Identity Card */
    .identity-card {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1rem 1.5rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
    }

    .identity-info {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .identity-name {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 1.125rem;
      font-weight: 600;
      color: white;
    }

    .identity-id {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 0.875rem;
    }

    .identity-id .label {
      color: #64748b;
    }

    .identity-id code {
      padding: 0.25rem 0.5rem;
      background: rgb(37 37 66 / 0.5);
      border-radius: 0.25rem;
      font-family: var(--font-family-mono);
      font-size: 0.75rem;
      color: var(--color-accent-400);
    }

    .copy-btn {
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

    .copy-btn:hover {
      background: rgb(37 37 66 / 0.5);
      color: white;
    }

    .identity-stats {
      display: flex;
      gap: 2rem;
    }

    .stat {
      display: flex;
      flex-direction: column;
      align-items: center;
    }

    .stat-value {
      font-size: 1.5rem;
      font-weight: 700;
      color: white;
    }

    .stat-label {
      font-size: 0.75rem;
      color: #64748b;
      text-transform: uppercase;
    }

    /* Role Badge */
    .role-badge {
      padding: 0.125rem 0.5rem;
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      text-transform: uppercase;
      font-weight: 600;
    }

    .role-badge.dual {
      background: rgb(0 212 255 / 0.1);
      color: var(--color-accent-500);
    }

    .role-badge.controller {
      background: rgb(168 85 247 / 0.1);
      color: rgb(168 85 247);
    }

    .role-badge.worker {
      background: rgb(34 197 94 / 0.1);
      color: var(--color-success-500);
    }

    /* Init Card */
    .init-card {
      padding: 2rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      text-align: center;
    }

    .init-card h3 {
      margin: 0 0 0.5rem 0;
      font-size: 1.125rem;
      color: white;
    }

    .init-card p {
      margin: 0 0 1.5rem 0;
      color: #64748b;
      font-size: 0.875rem;
    }

    .init-form {
      display: flex;
      flex-direction: column;
      gap: 1rem;
      max-width: 320px;
      margin: 0 auto;
    }

    .form-group {
      display: flex;
      flex-direction: column;
      gap: 0.375rem;
      text-align: left;
    }

    .form-group label {
      font-size: 0.75rem;
      font-weight: 600;
      color: #94a3b8;
      text-transform: uppercase;
    }

    .form-group input,
    .form-group select {
      padding: 0.5rem 0.75rem;
      background: var(--color-surface-200);
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      color: white;
      font-size: 0.875rem;
    }

    .form-group input:focus,
    .form-group select:focus {
      outline: none;
      border-color: var(--color-accent-500);
    }

    .form-group .hint {
      font-size: 0.75rem;
      color: #64748b;
    }

    /* Buttons */
    .btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
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

    .btn-secondary {
      background: rgb(37 37 66 / 0.5);
      color: white;
    }

    .btn-secondary:hover:not(:disabled) {
      background: rgb(37 37 66 / 0.8);
    }

    /* Peers Table */
    .peers-table-container {
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      overflow: hidden;
    }

    .peers-table {
      width: 100%;
      border-collapse: collapse;
    }

    .peers-table th {
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

    .peers-table td {
      padding: 0.75rem 1rem;
      font-size: 0.875rem;
      color: #e2e8f0;
      border-bottom: 1px solid rgb(37 37 66 / 0.1);
    }

    .peers-table tbody tr:hover {
      background: rgb(37 37 66 / 0.2);
    }

    .text-right { text-align: right; }
    .text-center { text-align: center; }
    .tabular-nums { font-variant-numeric: tabular-nums; }
    .text-muted { color: #64748b; }
    .text-success-500 { color: var(--color-success-500); }
    .text-warning-500 { color: var(--color-warning-500); }
    .text-danger-500 { color: var(--color-danger-500); }

    .peer-name {
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

    .status-dot.offline {
      background: #64748b;
    }

    .address-code {
      padding: 0.125rem 0.375rem;
      background: rgb(37 37 66 / 0.5);
      border-radius: 0.25rem;
      font-family: var(--font-family-mono);
      font-size: 0.75rem;
      color: #94a3b8;
    }

    .actions-cell {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.25rem;
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

    /* Empty State */
    .empty-state {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      padding: 3rem 2rem;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      text-align: center;
    }

    .empty-icon {
      color: #475569;
    }

    .empty-state h3 {
      margin: 1rem 0 0.5rem;
      font-size: 1.125rem;
      font-weight: 600;
      color: white;
    }

    .empty-state p {
      margin: 0 0 1.5rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    /* Modal */
    .modal-overlay {
      position: fixed;
      inset: 0;
      display: flex;
      align-items: center;
      justify-content: center;
      background: rgba(0, 0, 0, 0.7);
      z-index: 100;
    }

    .modal {
      width: 100%;
      max-width: 400px;
      background: var(--color-surface-100);
      border-radius: 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.3);
      box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5);
    }

    .modal-wide {
      max-width: 600px;
    }

    .modal-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 1rem 1.5rem;
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
    }

    .modal-header h3 {
      margin: 0;
      font-size: 1rem;
      font-weight: 600;
      color: white;
    }

    .close-btn {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 28px;
      height: 28px;
      background: transparent;
      border: none;
      border-radius: 0.25rem;
      color: #64748b;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .close-btn:hover {
      background: rgb(37 37 66 / 0.5);
      color: white;
    }

    .modal-body {
      padding: 1.5rem;
    }

    .modal-actions {
      display: flex;
      justify-content: flex-end;
      gap: 0.75rem;
      margin-top: 1.5rem;
    }

    /* Peer Stats */
    .peer-stats-grid {
      display: grid;
      gap: 1rem;
    }

    .miner-stat-card {
      padding: 1rem;
      background: var(--color-surface-200);
      border-radius: 0.5rem;
    }

    .miner-header {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      margin-bottom: 0.75rem;
    }

    .miner-name {
      font-weight: 600;
      color: white;
    }

    .algo-badge {
      padding: 0.125rem 0.375rem;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.25rem;
      font-size: 0.6875rem;
      color: var(--color-accent-500);
      text-transform: uppercase;
    }

    .miner-stats {
      display: flex;
      gap: 1.5rem;
    }

    .stat-item {
      display: flex;
      flex-direction: column;
    }

    .stat-item .stat-value {
      font-size: 1rem;
      font-weight: 600;
      color: white;
    }

    .stat-item .stat-label {
      font-size: 0.6875rem;
      color: #64748b;
      text-transform: uppercase;
    }

    .loading-stats {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.75rem;
      padding: 2rem;
      color: #64748b;
    }

    .spinner {
      width: 20px;
      height: 20px;
      border: 2px solid rgb(37 37 66 / 0.5);
      border-top-color: var(--color-accent-500);
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }

    .spinner-sm {
      width: 14px;
      height: 14px;
      border: 2px solid rgb(255 255 255 / 0.3);
      border-top-color: currentColor;
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }

    @keyframes spin {
      to { transform: rotate(360deg); }
    }
  `]
})
export class NodesComponent implements OnInit, OnDestroy {
  nodeService = inject(NodeService);
  actionInProgress = signal<string | null>(null);

  // Init form
  newNodeName = '';
  newNodeRole: 'controller' | 'worker' | 'dual' = 'dual';

  // Add peer modal
  showAddPeerModal = false;
  newPeerAddress = '';
  newPeerName = '';

  // Peer stats modal
  selectedPeer: Peer | null = null;
  selectedPeerStats: { miners: any[] } | null = null;

  ngOnInit() {
    this.nodeService.startPolling();
  }

  ngOnDestroy() {
    this.nodeService.stopPolling();
  }

  initializeNode() {
    if (!this.newNodeName) return;

    this.actionInProgress.set('init-node');
    this.nodeService.initNode(this.newNodeName, this.newNodeRole).subscribe({
      next: () => {
        this.actionInProgress.set(null);
        this.newNodeName = '';
        this.newNodeRole = 'dual';
      },
      error: (err) => {
        this.actionInProgress.set(null);
        console.error('Failed to initialize node:', err);
      }
    });
  }

  addPeer() {
    if (!this.newPeerAddress) return;

    this.actionInProgress.set('add-peer');
    this.nodeService.addPeer(this.newPeerAddress, this.newPeerName || undefined).subscribe({
      next: () => {
        this.actionInProgress.set(null);
        this.showAddPeerModal = false;
        this.newPeerAddress = '';
        this.newPeerName = '';
      },
      error: (err) => {
        this.actionInProgress.set(null);
        console.error('Failed to add peer:', err);
      }
    });
  }

  removePeer(peerId: string) {
    this.actionInProgress.set(`remove-${peerId}`);
    this.nodeService.removePeer(peerId).subscribe({
      next: () => this.actionInProgress.set(null),
      error: (err) => {
        this.actionInProgress.set(null);
        console.error('Failed to remove peer:', err);
      }
    });
  }

  pingPeer(peerId: string) {
    this.actionInProgress.set(`ping-${peerId}`);
    this.nodeService.pingPeer(peerId).subscribe({
      next: () => this.actionInProgress.set(null),
      error: (err) => {
        this.actionInProgress.set(null);
        console.error('Failed to ping peer:', err);
      }
    });
  }

  viewPeerStats(peer: Peer) {
    this.selectedPeer = peer;
    this.selectedPeerStats = null;

    this.nodeService.getPeerStats(peer.id).subscribe({
      next: (stats) => {
        this.selectedPeerStats = stats;
      },
      error: (err) => {
        console.error('Failed to get peer stats:', err);
        this.selectedPeerStats = { miners: [] };
      }
    });
  }

  copyToClipboard(text: string) {
    navigator.clipboard.writeText(text);
  }

  formatLastSeen(lastSeen: string): string {
    if (!lastSeen) return 'Never';
    const date = new Date(lastSeen);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSecs = Math.floor(diffMs / 1000);

    if (diffSecs < 60) return `${diffSecs}s ago`;
    if (diffSecs < 3600) return `${Math.floor(diffSecs / 60)}m ago`;
    if (diffSecs < 86400) return `${Math.floor(diffSecs / 3600)}h ago`;
    return `${Math.floor(diffSecs / 86400)}d ago`;
  }

  formatHashrate(hashrate: number): string {
    if (hashrate >= 1000000000) return `${(hashrate / 1000000000).toFixed(2)} GH/s`;
    if (hashrate >= 1000000) return `${(hashrate / 1000000).toFixed(2)} MH/s`;
    if (hashrate >= 1000) return `${(hashrate / 1000).toFixed(2)} kH/s`;
    return `${hashrate.toFixed(0)} H/s`;
  }

  formatUptime(seconds: number): string {
    if (seconds < 60) return `${seconds}s`;
    if (seconds < 3600) {
      const mins = Math.floor(seconds / 60);
      return `${mins}m`;
    }
    const hours = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${mins}m`;
  }
}
