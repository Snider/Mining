import { Component, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { SidebarComponent } from '../components/sidebar/sidebar.component';
import { StatsPanelComponent } from '../components/stats-panel/stats-panel.component';
import { MinerSwitcherComponent } from '../components/miner-switcher/miner-switcher.component';
import { WorkersComponent } from '../pages/workers/workers.component';
import { GraphsComponent } from '../pages/graphs/graphs.component';
import { ConsoleComponent } from '../pages/console/console.component';
import { PoolsComponent } from '../pages/pools/pools.component';
import { ProfilesComponent } from '../pages/profiles/profiles.component';
import { MinersComponent } from '../pages/miners/miners.component';
import { NodesComponent } from '../pages/nodes/nodes.component';

@Component({
  selector: 'app-main-layout',
  standalone: true,
  imports: [
    CommonModule,
    SidebarComponent,
    StatsPanelComponent,
    MinerSwitcherComponent,
    WorkersComponent,
    GraphsComponent,
    ConsoleComponent,
    PoolsComponent,
    ProfilesComponent,
    MinersComponent,
    NodesComponent
  ],
  template: `
    <div class="main-layout">
      <app-sidebar [currentRoute]="currentRoute()" (routeChange)="onRouteChange($event)"></app-sidebar>

      <div class="main-content">
        <div class="top-bar">
          <app-stats-panel></app-stats-panel>
          <app-miner-switcher (editProfile)="navigateToProfiles($event)"></app-miner-switcher>
        </div>

        <div class="page-content">
          @switch (currentRoute()) {
            @case ('workers') {
              <app-workers></app-workers>
            }
            @case ('graphs') {
              <app-graphs></app-graphs>
            }
            @case ('console') {
              <app-console></app-console>
            }
            @case ('pools') {
              <app-pools></app-pools>
            }
            @case ('profiles') {
              <app-profiles></app-profiles>
            }
            @case ('miners') {
              <app-miners></app-miners>
            }
            @case ('nodes') {
              <app-nodes></app-nodes>
            }
            @default {
              <app-workers></app-workers>
            }
          }
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
  `]
})
export class MainLayoutComponent {
  currentRoute = signal('workers');
  private editingProfileId: string | null = null;

  onRouteChange(route: string) {
    this.currentRoute.set(route);
  }

  navigateToProfiles(profileId: string) {
    this.editingProfileId = profileId;
    this.currentRoute.set('profiles');
    // TODO: Could emit event to profiles page to open edit modal for this profile
  }
}
