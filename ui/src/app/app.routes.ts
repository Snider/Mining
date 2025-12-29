import { Routes } from '@angular/router';
import { MainLayoutComponent } from './layouts/main-layout.component';
import { WorkersComponent } from './pages/workers/workers.component';
import { GraphsComponent } from './pages/graphs/graphs.component';
import { ConsoleComponent } from './pages/console/console.component';
import { PoolsComponent } from './pages/pools/pools.component';
import { ProfilesComponent } from './pages/profiles/profiles.component';
import { MinersComponent } from './pages/miners/miners.component';
import { SystemTrayComponent } from './pages/system-tray/system-tray.component';

export const routes: Routes = [
  // System tray is standalone without layout
  { path: 'system-tray', component: SystemTrayComponent },

  // All other routes use the main layout
  {
    path: '',
    component: MainLayoutComponent,
    children: [
      { path: '', redirectTo: 'workers', pathMatch: 'full' },
      { path: 'workers', component: WorkersComponent },
      { path: 'graphs', component: GraphsComponent },
      { path: 'console', component: ConsoleComponent },
      { path: 'pools', component: PoolsComponent },
      { path: 'profiles', component: ProfilesComponent },
      { path: 'miners', component: MinersComponent },
    ]
  },
];
