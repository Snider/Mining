import { Component, ViewEncapsulation, CUSTOM_ELEMENTS_SCHEMA, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from './miner.service';
import { SetupWizardComponent } from './setup-wizard.component';
import { MiningDashboardComponent } from './dashboard.component';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/card/card.js';
import '@awesome.me/webawesome/dist/components/spinner/spinner.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/icon/icon.js';

@Component({
  selector: 'snider-mining',
  standalone: true,
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  imports: [
    CommonModule,
    SetupWizardComponent,
    MiningDashboardComponent
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.css'],
  encapsulation: ViewEncapsulation.ShadowDom
})
export class SniderMining {
  minerService = inject(MinerService);
  state = this.minerService.state;

  checkSystemState() {
    this.minerService.checkSystemState();
  }
}
