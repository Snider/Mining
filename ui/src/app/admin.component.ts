import { Component, OnInit, Input, CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import { HttpClient, HttpClientModule, HttpErrorResponse } from '@angular/common/http';
import { CommonModule } from '@angular/common';
import { of, forkJoin } from 'rxjs';
import { map, catchError } from 'rxjs/operators';

// Define interfaces for our data structures
interface InstallationDetails {
  is_installed: boolean;
  version: string;
  path: string;
  miner_binary: string;
  config_path?: string;
  type?: string;
}

interface AvailableMiner {
  name: string;
  description: string;
}

@Component({
  selector: 'snider-mining-admin',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [CommonModule, HttpClientModule],
  templateUrl: './admin.component.html',
  styleUrls: ['./admin.component.css'],
})
export class MiningAdminComponent implements OnInit {
  @Input() needsSetup: boolean = false; // Input to trigger setup mode

  apiBaseUrl: string = 'http://localhost:9090/api/v1/mining';

  error: string | null = null;
  actionInProgress: string | null = null;
  manageableMiners: any[] = [];
  whitelistPaths: string[] = [];

  constructor(private http: HttpClient) {}

  ngOnInit(): void {
    this.getAdminData();
  }

  private handleError(err: HttpErrorResponse, defaultMessage: string) {
    console.error(err);
    this.actionInProgress = null;
    if (err.error && err.error.error) {
      this.error = `${defaultMessage}: ${err.error.error}`;
    } else {
      this.error = `${defaultMessage}. Please check the console for details.`;
    }
  }

  getAdminData(): void {
    this.error = null;
    forkJoin({
      available: this.http.get<AvailableMiner[]>(`${this.apiBaseUrl}/miners/available`),
      info: this.http.get<any>(`${this.apiBaseUrl}/info`).pipe(catchError(() => of({}))) // Gracefully handle info error
    }).pipe(
      map(({ available, info }) => {
        const installedMap = new Map<string, InstallationDetails>(
          (info.installed_miners_info || []).map((m: InstallationDetails) => [this.getMinerType(m), m])
        );

        this.manageableMiners = available.map(availMiner => ({
          ...availMiner,
          is_installed: installedMap.get(availMiner.name)?.is_installed ?? false,
        }));

        const installedMiners = (info.installed_miners_info || []).filter((m: InstallationDetails) => m.is_installed);
        this.updateWhitelistPaths(installedMiners, []);
      }),
      catchError(err => {
        this.handleError(err, 'Could not load miner information');
        return of(null);
      })
    ).subscribe();
  }

  private updateWhitelistPaths(installed: InstallationDetails[], running: any[]) {
    const paths = new Set<string>();
    installed.forEach(miner => {
      if (miner.miner_binary) paths.add(miner.miner_binary);
      if (miner.config_path) paths.add(miner.config_path);
    });
    running.forEach(miner => {
      if (miner.configPath) paths.add(miner.configPath);
    });
    this.whitelistPaths = Array.from(paths);
  }

  installMiner(minerType: string): void {
    this.actionInProgress = `install-${minerType}`;
    this.error = null;
    this.http.post(`${this.apiBaseUrl}/miners/${minerType}/install`, {}).subscribe({
      next: () => {
        setTimeout(() => {
          this.getAdminData();
          // A simple way to signal completion is to reload the page
          // so the main dashboard component re-evaluates its state.
          if (this.needsSetup) {
            window.location.reload();
          }
        }, 1000);
      },
      error: (err: HttpErrorResponse) => this.handleError(err, `Failed to install ${minerType}`)
    });
  }

  uninstallMiner(minerType: string): void {
    this.actionInProgress = `uninstall-${minerType}`;
    this.error = null;
    this.http.delete(`${this.apiBaseUrl}/miners/${minerType}/uninstall`).subscribe({
      next: () => {
        setTimeout(() => {
          this.getAdminData();
        }, 1000);
      },
      error: (err: HttpErrorResponse) => this.handleError(err, `Failed to uninstall ${minerType}`)
    });
  }

  getMinerType(miner: any): string {
    if (!miner.path) return 'unknown';
    const parts = miner.path.split('/').filter((p: string) => p);
    return parts.length > 1 ? parts[parts.length - 2] : parts[parts.length - 1] || 'unknown';
  }
}
