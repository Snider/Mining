import { Routes } from '@angular/router';
import { WorkersComponent } from './pages/workers/workers.component';
import { GraphsComponent } from './pages/graphs/graphs.component';
import { ConsoleComponent } from './pages/console/console.component';
import { PoolsComponent } from './pages/pools/pools.component';
import { ProfilesComponent } from './pages/profiles/profiles.component';
import { MinersComponent } from './pages/miners/miners.component';
import { NodesComponent } from './pages/nodes/nodes.component';
import { SystemTrayComponent } from './pages/system-tray/system-tray.component';

export const routes: Routes = [
  // System tray is standalone without layout
  { path: 'system-tray', component: SystemTrayComponent },

  // Main app routes - MainLayoutComponent is rendered directly and contains router-outlet
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: 'dashboard', component: GraphsComponent },
  { path: 'workers', component: WorkersComponent },
  { path: 'console', component: ConsoleComponent },
  { path: 'pools', component: PoolsComponent },
  { path: 'profiles', component: ProfilesComponent },
  { path: 'miners', component: MinersComponent },
  { path: 'nodes', component: NodesComponent },
];
