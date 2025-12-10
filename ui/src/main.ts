import { createApplication } from '@angular/platform-browser';
import { createCustomElement } from '@angular/elements';
import { SniderMining } from './app/app';
import { MiningAdminComponent } from './app/admin.component';
import { SetupWizardComponent } from './app/setup-wizard.component';
import { MiningDashboardComponent } from './app/dashboard.component';
import { ChartComponent } from './app/chart.component';
import { ProfileListComponent } from './app/profile-list.component';
import { ProfileCreateComponent } from './app/profile-create.component';
import { appConfig } from './app/app.config';

(async () => {
  const app = await createApplication(appConfig);

  // Define the main app element
  const AppElement = createCustomElement(SniderMining, { injector: app.injector });
  customElements.define('snider-mining', AppElement);

  // Define the setup wizard element
  const SetupWizardElement = createCustomElement(SetupWizardComponent, { injector: app.injector });
  customElements.define('snider-mining-setup-wizard', SetupWizardElement);

  // Define the dashboard view element
  const DashboardElement = createCustomElement(MiningDashboardComponent, { injector: app.injector });
  customElements.define('snider-mining-dashboard', DashboardElement);

  // Define the chart element
  const ChartElement = createCustomElement(ChartComponent, { injector: app.injector });
  customElements.define('snider-mining-chart', ChartElement);

  // Define the admin element as a separate, secondary element
  const AdminElement = createCustomElement(MiningAdminComponent, { injector: app.injector });
  customElements.define('snider-mining-admin', AdminElement);

  // Define the profile list element
  const ProfileListElement = createCustomElement(ProfileListComponent, { injector: app.injector });
  customElements.define('snider-mining-profile-list', ProfileListElement);

  // Define the profile create element
  const ProfileCreateElement = createCustomElement(ProfileCreateComponent, { injector: app.injector });
  customElements.define('snider-mining-profile-create', ProfileCreateElement);

  console.log('All Snider Mining custom elements registered.');
})();
