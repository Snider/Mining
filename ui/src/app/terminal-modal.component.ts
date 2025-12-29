import { Component, Input, Output, EventEmitter, OnInit, OnDestroy, ElementRef, ViewChild } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MinerService } from './miner.service';
import { interval, Subscription } from 'rxjs';
import { switchMap, catchError } from 'rxjs/operators';
import { of } from 'rxjs';

@Component({
  selector: 'app-terminal-modal',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="modal-overlay" (click)="close.emit()">
      <div class="modal-content" (click)="$event.stopPropagation()">
        <div class="modal-header">
          <h3>{{ minerName }} - Terminal Output</h3>
          <button class="close-btn" (click)="close.emit()">&times;</button>
        </div>
        <div class="terminal" #terminal>
          <div class="terminal-line" *ngFor="let line of logs">{{ stripAnsi(line) }}</div>
          <div class="terminal-line" *ngIf="logs.length === 0">Waiting for output...</div>
        </div>
        <div class="modal-footer">
          <span class="status">{{ logs.length }} lines</span>
          <label class="auto-scroll-label">
            <input type="checkbox" [(ngModel)]="autoScroll" (change)="onAutoScrollChange()">
            Auto-scroll
          </label>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .modal-overlay {
      position: fixed;
      top: 0;
      left: 0;
      right: 0;
      bottom: 0;
      background: rgba(0, 0, 0, 0.7);
      display: flex;
      align-items: center;
      justify-content: center;
      z-index: 1000;
    }

    .modal-content {
      background: #1e1e1e;
      border-radius: 8px;
      width: 90%;
      max-width: 900px;
      height: 70vh;
      display: flex;
      flex-direction: column;
      box-shadow: 0 4px 20px rgba(0, 0, 0, 0.5);
    }

    .modal-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 12px 16px;
      border-bottom: 1px solid #333;
    }

    .modal-header h3 {
      margin: 0;
      color: #fff;
      font-size: 14px;
      font-weight: 500;
    }

    .close-btn {
      background: none;
      border: none;
      color: #888;
      font-size: 24px;
      cursor: pointer;
      padding: 0;
      line-height: 1;
    }

    .close-btn:hover {
      color: #fff;
    }

    .terminal {
      flex: 1;
      overflow-y: auto;
      padding: 12px;
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 12px;
      line-height: 1.5;
      background: #0d0d0d;
      color: #00ff00;
    }

    .terminal-line {
      white-space: pre-wrap;
      word-break: break-all;
    }

    .modal-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 8px 16px;
      border-top: 1px solid #333;
      color: #888;
      font-size: 12px;
    }

    .auto-scroll-label {
      display: flex;
      align-items: center;
      gap: 6px;
      cursor: pointer;
      color: #888;
    }

    .auto-scroll-label input {
      cursor: pointer;
    }
  `]
})
export class TerminalModalComponent implements OnInit, OnDestroy {
  @Input() minerName: string = '';
  @Output() close = new EventEmitter<void>();
  @ViewChild('terminal') terminalEl!: ElementRef;

  logs: string[] = [];
  autoScroll = true;
  private pollSubscription?: Subscription;

  constructor(private minerService: MinerService) {}

  ngOnInit() {
    this.startPolling();
  }

  ngOnDestroy() {
    this.stopPolling();
  }

  private startPolling() {
    // Initial fetch
    this.fetchLogs();

    // Poll every 2 seconds
    this.pollSubscription = interval(2000).pipe(
      switchMap(() => this.minerService.getMinerLogs(this.minerName)),
      catchError(() => of([]))
    ).subscribe(logs => {
      this.logs = logs;
      if (this.autoScroll) {
        this.scrollToBottom();
      }
    });
  }

  private stopPolling() {
    this.pollSubscription?.unsubscribe();
  }

  private fetchLogs() {
    this.minerService.getMinerLogs(this.minerName).pipe(
      catchError(() => of([]))
    ).subscribe(logs => {
      this.logs = logs;
      if (this.autoScroll) {
        setTimeout(() => this.scrollToBottom(), 0);
      }
    });
  }

  private scrollToBottom() {
    if (this.terminalEl) {
      const el = this.terminalEl.nativeElement;
      el.scrollTop = el.scrollHeight;
    }
  }

  onAutoScrollChange() {
    if (this.autoScroll) {
      this.scrollToBottom();
    }
  }

  // Strip ANSI escape codes for cleaner display
  stripAnsi(text: string): string {
    return text.replace(/\x1b\[[0-9;]*m/g, '');
  }
}
