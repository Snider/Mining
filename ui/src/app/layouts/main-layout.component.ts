import { Component, inject, AfterViewInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router, RouterOutlet, NavigationEnd } from '@angular/router';
import { filter, map } from 'rxjs/operators';
import { toSignal } from '@angular/core/rxjs-interop';
import { SidebarComponent } from '../components/sidebar/sidebar.component';
import { StatsPanelComponent } from '../components/stats-panel/stats-panel.component';
import { MinerSwitcherComponent } from '../components/miner-switcher/miner-switcher.component';
import { ToastComponent } from '../components/toast/toast.component';
import { ApiStatusComponent } from '../components/api-status/api-status.component';

@Component({
  selector: 'app-main-layout',
  standalone: true,
  imports: [
    CommonModule,
    RouterOutlet,
    SidebarComponent,
    StatsPanelComponent,
    MinerSwitcherComponent,
    ToastComponent,
    ApiStatusComponent,
  ],
  template: `
    <app-api-status></app-api-status>
    <app-toast></app-toast>
    <div class="main-layout">
      <app-sidebar [currentRoute]="currentRoute()" (routeChange)="onRouteChange($event)"></app-sidebar>

      <div class="main-content">
        <div class="top-bar">
          <app-stats-panel></app-stats-panel>
          <app-miner-switcher (editProfile)="navigateToProfiles($event)"></app-miner-switcher>
        </div>

        <div class="page-content">
          <router-outlet></router-outlet>
        </div>
      </div>
    </div>
  `,
  styles: [`
    .main-layout {
      display: flex;
      min-height: 100vh;
      background: var(--color-surface-400);
    }

    .main-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      min-width: 0;
    }

    .top-bar {
      display: flex;
      align-items: center;
      gap: 1rem;
      background: var(--color-surface-100);
      border-bottom: 1px solid rgb(37 37 66 / 0.2);
      padding-right: 1rem;
    }

    .top-bar app-stats-panel {
      flex: 1;
    }

    .page-content {
      flex: 1;
      overflow-y: auto;
      padding: 1rem;
    }

    /* Mobile responsive styles */
    @media (max-width: 768px) {
      .top-bar {
        flex-direction: column;
        align-items: stretch;
        gap: 0.5rem;
        padding: 0.75rem;
        padding-left: 3.5rem; /* Space for hamburger menu */
      }

      .page-content {
        padding: 0.75rem;
      }
    }
  `]
})
export class MainLayoutComponent implements AfterViewInit {
  private router = inject(Router);

  // Track current route from router events
  currentRoute = toSignal(
    this.router.events.pipe(
      filter((event): event is NavigationEnd => event instanceof NavigationEnd),
      map(event => {
        // Extract route from URL like "/#/workers" or "/workers"
        const url = event.urlAfterRedirects;
        const segments = url.split('/').filter(s => s && s !== '#');
        return segments[0] || 'dashboard';
      })
    ),
    { initialValue: this.getInitialRoute() }
  );

  private getInitialRoute(): string {
    const url = this.router.url;
    const segments = url.split('/').filter(s => s && s !== '#');
    return segments[0] || 'dashboard';
  }

  ngAfterViewInit() {
    // Re-trigger navigation after router-outlet is available
    // This handles the case where router tried to navigate before outlet existed
    const route = this.getInitialRoute();
    setTimeout(() => this.router.navigate(['/', route]), 0);
  }

  onRouteChange(route: string) {
    this.router.navigate(['/', route]);
  }

  navigateToProfiles(profileId: string) {
    // TODO: Could pass profileId via query params or state
    this.router.navigate(['/', 'profiles']);
  }
}
