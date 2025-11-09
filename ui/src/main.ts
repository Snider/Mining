import { createApplication } from '@angular/platform-browser';
import { createCustomElement } from '@angular/elements';
import { MiningDashboardElementComponent } from './app/app'; // Renamed App to MiningDashboardElementComponent
import { importProvidersFrom } from '@angular/core';
import { HttpClientModule } from '@angular/common/http';
import { CommonModule } from '@angular/common';

(async () => {
  // Bootstrap a minimal Angular application to provide
  // necessary services like HttpClient to the custom element.
  const app = await createApplication({
    providers: [
      importProvidersFrom(HttpClientModule, CommonModule)
    ]
  });

  // Define your custom element
  const MiningDashboardElement = createCustomElement(MiningDashboardElementComponent, { injector: app.injector });

  // Register the custom element with the browser
  customElements.define('mde-mining-dashboard', MiningDashboardElement);

  console.log('mde-mining-dashboard custom element registered!');
})();
