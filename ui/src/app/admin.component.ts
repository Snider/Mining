import { Component, CUSTOM_ELEMENTS_SCHEMA, inject, signal, computed } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpErrorResponse } from '@angular/common/http';
import { MinerService } from './miner.service';


@Component({
  selector: 'snider-mining-admin',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule],
  templateUrl: './admin.component.html',
  styleUrls: ['./admin.component.css'],
})
export class MiningAdminComponent {
  minerService = inject(MinerService);
  state = this.minerService.state;
  actionInProgress = signal<string | null>(null);
  error = signal<string | null>(null);

  whitelistPaths = computed(() => {
    const paths = new Set<string>();
    this.state().installedMiners.forEach(miner => {
      if (miner.miner_binary) paths.add(miner.miner_binary);
      if (miner.config_path) paths.add(miner.config_path);
    });
    this.state().runningMiners.forEach(miner => {
      if ((miner as any).configPath) paths.add((miner as any).configPath);
    });
    return Array.from(paths);
  });

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
