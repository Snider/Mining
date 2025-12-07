import { createApplication } from '@angular/platform-browser';
import { createCustomElement } from '@angular/elements';
import { MiningDashboardElementComponent } from './app/app';
import { MiningAdminComponent } from './app/admin.component';
import { appConfig } from './app/app.config';

(async () => {
  const app = await createApplication(appConfig);

  // Define the dashboard element as the primary application root
  const DashboardElement = createCustomElement(MiningDashboardElementComponent, { injector: app.injector });
  customElements.define('snider-mining-dashboard', DashboardElement);
  console.log('snider-mining-dashboard custom element registered!');

  // // Define the admin element as a separate, secondary element
  const AdminElement = createCustomElement(MiningAdminComponent, { injector: app.injector });
  customElements.define('snider-mining-admin', AdminElement);
  console.log('snider-mining-admin custom element registered!');
})();
