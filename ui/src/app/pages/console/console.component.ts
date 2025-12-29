import { Component, inject, computed, signal, OnInit, OnDestroy, ElementRef, ViewChild, AfterViewChecked } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from '../../miner.service';
import { HttpClient } from '@angular/common/http';
import { interval, Subscription, switchMap } from 'rxjs';

@Component({
  selector: 'app-console',
  standalone: true,
  imports: [CommonModule],
  template: `
    <div class="console-page">
      <!-- Console Header with Worker Chooser (when in "all" mode) -->
      <div class="console-header">
        @if (viewMode() === 'all') {
          <!-- Worker Chooser Dropdown -->
          <div class="worker-chooser">
            <label class="chooser-label">Worker:</label>
            <select
              class="worker-select"
              [value]="consoleSelectedMiner() || ''"
              (change)="onWorkerChange($event)">
              @for (miner of runningMiners(); track miner.name) {
                <option [value]="miner.name">{{ miner.name }}</option>
              }
            </select>
          </div>
        } @else {
          <!-- Single miner mode indicator -->
          <div class="single-miner-indicator">
            <div class="status-dot online"></div>
            <span>{{ globalSelectedMiner() }}</span>
          </div>
        }

        <!-- Miner Tabs (alternative view when multiple miners) -->
        @if (viewMode() === 'all' && runningMiners().length > 1) {
          <div class="console-tabs">
            @for (miner of runningMiners(); track miner.name) {
              <button
                class="tab-btn"
                [class.active]="consoleSelectedMiner() === miner.name"
                (click)="selectConsoleMiner(miner.name)">
                <div class="tab-status" [class.online]="true"></div>
                {{ miner.name }}
              </button>
            }
          </div>
        }

        @if (runningMiners().length === 0) {
          <div class="no-miners-msg">No active workers</div>
        }
      </div>

      <!-- Console Output -->
      <div class="console-output" #consoleOutput>
        @if (logs().length > 0) {
          @for (line of logs(); track $index) {
            <div class="log-line" [class.error]="isErrorLine(line)" [class.warning]="isWarningLine(line)">
              <span class="log-text">{{ line }}</span>
            </div>
          }
        } @else if (activeMiner()) {
          <div class="console-empty">
            <p>Waiting for logs from {{ activeMiner() }}...</p>
          </div>
        } @else {
          <div class="console-empty">
            <svg class="w-12 h-12" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 9l3 3-3 3m5 0h3M5 20h14a2 2 0 002-2V6a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
            </svg>
            <p>Start a miner to see console output</p>
          </div>
        }
      </div>

      <!-- Console Controls -->
      <div class="console-controls">
        <label class="control-checkbox">
          <input type="checkbox" [checked]="autoScroll()" (change)="toggleAutoScroll()">
          <span>Auto-scroll</span>
        </label>
        <button class="control-btn" (click)="clearLogs()" [disabled]="logs().length === 0">
          <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/>
          </svg>
          Clear
        </button>
      </div>
    </div>
  `,
  styles: [`
    .console-page {
      display: flex;
      flex-direction: column;
      height: calc(100vh - 120px);
      gap: 0;
    }

    .console-header {
      display: flex;
      align-items: center;
      gap: 1rem;
      padding: 0.5rem 0.75rem;
      background: var(--color-surface-200);
      border-radius: 0.5rem 0.5rem 0 0;
      border: 1px solid rgb(37 37 66 / 0.2);
      border-bottom: none;
    }

    .worker-chooser {
      display: flex;
      align-items: center;
      gap: 0.5rem;
    }

    .chooser-label {
      font-size: 0.75rem;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .worker-select {
      padding: 0.375rem 0.625rem;
      background: var(--color-surface-100);
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.375rem;
      color: white;
      font-size: 0.8125rem;
      cursor: pointer;
      min-width: 140px;
    }

    .worker-select:hover,
    .worker-select:focus {
      border-color: var(--color-accent-500);
      outline: none;
    }

    .single-miner-indicator {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.375rem 0.625rem;
      background: rgb(0 212 255 / 0.1);
      border-radius: 0.375rem;
      color: var(--color-accent-400);
      font-size: 0.8125rem;
      font-weight: 500;
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

    .no-miners-msg {
      padding: 0.375rem 0;
      color: #64748b;
      font-size: 0.8125rem;
    }

    .console-tabs {
      display: flex;
      align-items: center;
      gap: 0.25rem;
      margin-left: auto;
    }

    .tab-btn {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      padding: 0.375rem 0.625rem;
      background: transparent;
      border: none;
      border-radius: 0.375rem;
      color: #94a3b8;
      font-size: 0.75rem;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .tab-btn:hover {
      background: rgb(37 37 66 / 0.3);
      color: white;
    }

    .tab-btn.active {
      background: var(--color-surface-100);
      color: white;
    }

    .tab-status {
      width: 6px;
      height: 6px;
      border-radius: 50%;
      background: #64748b;
    }

    .tab-status.online {
      background: var(--color-success-500);
      box-shadow: 0 0 4px var(--color-success-500);
    }

    .console-output {
      flex: 1;
      overflow-y: auto;
      background: #0a0a12;
      border-left: 1px solid rgb(37 37 66 / 0.2);
      border-right: 1px solid rgb(37 37 66 / 0.2);
      font-family: var(--font-family-mono);
      font-size: 0.8125rem;
      line-height: 1.5;
    }

    .log-line {
      padding: 0.125rem 0.75rem;
      border-bottom: 1px solid rgb(37 37 66 / 0.05);
    }

    .log-line:hover {
      background: rgb(37 37 66 / 0.2);
    }

    .log-text {
      color: #a3e635;
      word-break: break-all;
    }

    .log-line.error .log-text {
      color: var(--color-danger-500);
    }

    .log-line.warning .log-text {
      color: var(--color-warning-500);
    }

    .console-empty {
      display: flex;
      flex-direction: column;
      align-items: center;
      justify-content: center;
      height: 100%;
      color: #64748b;
    }

    .console-empty p {
      margin-top: 0.75rem;
      font-size: 0.875rem;
    }

    .console-controls {
      display: flex;
      align-items: center;
      justify-content: space-between;
      padding: 0.5rem 0.75rem;
      background: var(--color-surface-200);
      border-radius: 0 0 0.5rem 0.5rem;
      border: 1px solid rgb(37 37 66 / 0.2);
      border-top: none;
    }

    .control-checkbox {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: #94a3b8;
      font-size: 0.8125rem;
      cursor: pointer;
    }

    .control-checkbox input {
      width: 14px;
      height: 14px;
      accent-color: var(--color-accent-500);
    }

    .control-btn {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      padding: 0.375rem 0.625rem;
      background: transparent;
      border: 1px solid rgb(37 37 66 / 0.3);
      border-radius: 0.25rem;
      color: #94a3b8;
      font-size: 0.75rem;
      cursor: pointer;
      transition: all 0.15s ease;
    }

    .control-btn:hover:not(:disabled) {
      background: rgb(37 37 66 / 0.3);
      color: white;
    }

    .control-btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
  `]
})
export class ConsoleComponent implements OnInit, OnDestroy, AfterViewChecked {
  @ViewChild('consoleOutput') consoleOutput!: ElementRef;

  private minerService = inject(MinerService);
  private http = inject(HttpClient);
  private state = this.minerService.state;
  private pollSub?: Subscription;

  runningMiners = computed(() => this.state().runningMiners);
  viewMode = this.minerService.viewMode;
  globalSelectedMiner = this.minerService.selectedMinerName;

  // Local console selection (used when in "all" mode to pick which logs to show)
  consoleSelectedMiner = signal<string | null>(null);

  // The actual miner whose logs we're showing
  activeMiner = computed(() => {
    // In single mode, use the global selection
    if (this.viewMode() === 'single') {
      return this.globalSelectedMiner();
    }
    // In all mode, use the console-specific selection
    return this.consoleSelectedMiner();
  });

  logs = signal<string[]>([]);
  autoScroll = signal(true);
  private shouldScroll = false;

  ngOnInit() {
    // Auto-select first miner for console view
    const miners = this.runningMiners();
    if (miners.length > 0) {
      this.consoleSelectedMiner.set(miners[0].name);
    }

    // Poll for logs every 2 seconds
    this.pollSub = interval(2000).pipe(
      switchMap(() => {
        const miner = this.activeMiner();
        if (!miner) return [];
        return this.http.get<{logs: string[]}>(`/api/v1/mining/miners/${miner}/logs`);
      })
    ).subscribe({
      next: (response: any) => {
        if (response?.logs) {
          this.logs.set(response.logs);
          if (this.autoScroll()) {
            this.shouldScroll = true;
          }
        }
      }
    });
  }

  ngOnDestroy() {
    this.pollSub?.unsubscribe();
  }

  ngAfterViewChecked() {
    if (this.shouldScroll && this.consoleOutput) {
      const el = this.consoleOutput.nativeElement;
      el.scrollTop = el.scrollHeight;
      this.shouldScroll = false;
    }
  }

  // Select miner in console (when in "all" mode)
  selectConsoleMiner(name: string) {
    this.consoleSelectedMiner.set(name);
    this.logs.set([]);
    this.fetchLogs(name);
  }

  private fetchLogs(minerName: string) {
    this.http.get<{logs: string[]}>(`/api/v1/mining/miners/${minerName}/logs`).subscribe({
      next: (response) => {
        if (response?.logs) {
          this.logs.set(response.logs);
          this.shouldScroll = true;
        }
      }
    });
  }

  toggleAutoScroll() {
    this.autoScroll.update(v => !v);
  }

  clearLogs() {
    this.logs.set([]);
  }

  onWorkerChange(event: Event) {
    const select = event.target as HTMLSelectElement;
    this.selectConsoleMiner(select.value);
  }

  isErrorLine(line: string): boolean {
    const lower = line.toLowerCase();
    return lower.includes('error') || lower.includes('failed') || lower.includes('fatal');
  }

  isWarningLine(line: string): boolean {
    const lower = line.toLowerCase();
    return lower.includes('warn') || lower.includes('timeout') || lower.includes('retry');
  }
}
