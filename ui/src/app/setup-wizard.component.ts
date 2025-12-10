import { Component, ViewEncapsulation, CUSTOM_ELEMENTS_SCHEMA, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { MinerService } from './miner.service';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/spinner/spinner.js';
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';

@Component({
  selector: 'snider-mining-setup-wizard',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule],
  templateUrl: './setup-wizard.component.html',
  styleUrls: ['./setup-wizard.component.css']
})
export class SetupWizardComponent {
  minerService = inject(MinerService);
  state = this.minerService.state;
  actionInProgress = signal<string | null>(null);
  error = signal<string | null>(null);

  installMiner(minerType: string): void {
    this.actionInProgress.set(`install-${minerType}`);
    this.error.set(null);
    this.minerService.installMiner(minerType).subscribe({
      next: () => { this.actionInProgress.set(null); },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to install ${minerType}`);
      }
    });
  }

  uninstallMiner(minerType: string): void {
    this.actionInProgress.set(`uninstall-${minerType}`);
    this.error.set(null);
    this.minerService.uninstallMiner(minerType).subscribe({
      next: () => { this.actionInProgress.set(null); },
      error: (err: HttpErrorResponse) => {
        this.handleError(err, `Failed to uninstall ${minerType}`);
      }
    });
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
}
