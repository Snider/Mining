import { Component, ViewEncapsulation, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MinerService } from './miner.service';
import { SetupWizardComponent } from './setup-wizard.component';
import { MainLayoutComponent } from './layouts/main-layout.component';

@Component({
  selector: 'snider-mining',
  standalone: true,
  imports: [
    CommonModule,
    SetupWizardComponent,
    MainLayoutComponent
  ],
  templateUrl: './app.html',
  styleUrls: ['./app.css'],
  encapsulation: ViewEncapsulation.ShadowDom
})
export class SniderMining {
  minerService = inject(MinerService);
  state = this.minerService.state;

  forceRefreshState() {
    this.minerService.forceRefreshState();
  }
}
