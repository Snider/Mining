import { Component, CUSTOM_ELEMENTS_SCHEMA, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { MinerService, MiningProfile } from './miner.service';

// Import Web Awesome components
import "@awesome.me/webawesome/dist/webawesome.js";
import '@awesome.me/webawesome/dist/components/input/input.js';
import '@awesome.me/webawesome/dist/components/select/select.js';
import '@awesome.me/webawesome/dist/components/checkbox/checkbox.js';
import '@awesome.me/webawesome/dist/components/button/button.js';
import '@awesome.me/webawesome/dist/components/card/card.js';

@Component({
  selector: 'snider-mining-profile-create',
  standalone: true,
  imports: [CommonModule, FormsModule],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  templateUrl: './profile-create.component.html',
  styleUrls: ['./profile-create.component.css']
})
export class ProfileCreateComponent {
  minerService = inject(MinerService);
  state = this.minerService.state;

  // Plain object model for the form
  model: MiningProfile = {
    id: '',
    name: '',
    minerType: '',
    config: {
      pool: '',
      wallet: '',
      tls: true,
      hugePages: true
    }
  };

  // Simple properties instead of signals
  error: string | null = null;
  success: string | null = null;

  createProfile() {
    this.error = null;
    this.success = null;

    // Basic validation check
    if (!this.model.name || !this.model.minerType || !this.model.config.pool || !this.model.config.wallet) {
      this.error = 'Please fill out all required fields.';
      return;
    }

    this.minerService.createProfile(this.model).subscribe({
      next: () => {
        this.success = 'Profile created successfully!';
        // Reset form to defaults
        this.model = {
          id: '',
          name: '',
          minerType: '',
          config: {
            pool: '',
            wallet: '',
            tls: true,
            hugePages: true
          }
        };
        setTimeout(() => this.success = null, 3000);
      },
      error: (err: HttpErrorResponse) => {
        console.error(err);
        if (err.error && err.error.error) {
          this.error = `Failed to create profile: ${err.error.error}`;
        } else if (typeof err.error === 'string' && err.error.length < 200) {
          this.error = `Failed to create profile: ${err.error}`;
        } else {
          this.error = 'An unknown error occurred while creating the profile.';
        }
      }
    });
  }
}
