import { Component, ViewEncapsulation, CUSTOM_ELEMENTS_SCHEMA, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { MinerService } from './miner.service';
import { ChartComponent } from './chart.component';
import { ProfileListComponent } from './profile-list.component';
import { ProfileCreateComponent } from './profile-create.component';
import { StatsBarComponent } from './stats-bar.component';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/tooltip/tooltip.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';
import '@awesome.me/webawesome/dist/components/spinner/spinner.js';
import '@awesome.me/webawesome/dist/components/input/input.js';
import '@awesome.me/webawesome/dist/components/select/select.js';

@Component({
  selector: 'snider-mining-dashboard',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, FormsModule, ChartComponent, ProfileListComponent, ProfileCreateComponent, StatsBarComponent],
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.css']
})
export class MiningDashboardComponent {
  minerService = inject(MinerService);
  state = this.minerService.state;

  actionInProgress = signal<string | null>(null);
  error = signal<string | null>(null);

  showProfileManager = signal(false);
  // Use a map to track the selected profile for each miner type
  selectedProfileIds = signal<Map<string, string>>(new Map());

  handleProfileSelection(minerType: string, event: Event) {
    const selectedValue = (event.target as HTMLSelectElement).value;
    this.selectedProfileIds.update(m => m.set(minerType, selectedValue));
  }

  private handleError(err: HttpErrorResponse, defaultMessage: string) {
    console.error(err);
    this.actionInProgress.set(null);
    if (err.error && err.error.error) {
      this.error.set(`${defaultMessage}: ${err.error.error}`);
    } else if (typeof err.error === 'string' && err.error.length < 200) {
      this.error.set(`${defaultMessage}: ${err.error}`);
    } else {
      this.error.set(`${defaultMessage}. Please check the console for details.`);
    }
  }

  startMiner(minerType: string): void {
    const profileId = this.selectedProfileIds().get(minerType);
    if (!profileId) {
      this.error.set('Please select a profile to start.');
      return;
    }
    this.actionInProgress.set(`start-${profileId}`);
    this.error.set(null);
    this.minerService.startMiner(profileId).subscribe({
      next: () => { this.actionInProgress.set(null); },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to start miner for profile ${profileId}`);
      }
    });
  }

  stopMiner(miner: any): void {
    const runningInstance = this.getRunningMinerInstance(miner);
    if (!runningInstance) {
      this.error.set("Cannot stop a miner that is not running.");
      return;
    }
    this.actionInProgress.set(`stop-${miner.type}`);
    this.error.set(null);
    this.minerService.stopMiner(runningInstance.name).subscribe({
      next: () => { this.actionInProgress.set(null); },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to stop ${runningInstance.name}`);
      }
    });
  }

  getRunningMinerInstance(miner: any): any {
    return this.state().runningMiners.find((m: any) => m.name.startsWith(miner.type));
  }

  isMinerRunning(miner: any): boolean {
    return !!this.getRunningMinerInstance(miner);
  }

  toggleProfileManager() {
    this.showProfileManager.set(!this.showProfileManager());
  }
}
