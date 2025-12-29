import { Component, CUSTOM_ELEMENTS_SCHEMA, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { HttpErrorResponse } from '@angular/common/http';
import { MinerService, MiningProfile } from './miner.service';


@Component({
  selector: 'snider-mining-profile-create',
  standalone: true,
  imports: [CommonModule, FormsModule],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  templateUrl: './profile-create.component.html'
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
  isCreating = signal(false);

  // --- Event Handlers for Custom Elements ---
  // By handling events here, we can safely cast the event target
  // to the correct type, satisfying TypeScript's strict checking.

  onNameInput(event: Event) {
    this.model.name = (event.target as HTMLInputElement).value;
  }

  onMinerTypeChange(event: Event) {
    this.model.minerType = (event.target as HTMLSelectElement).value;
  }

  onPoolInput(event: Event) {
    this.model.config.pool = (event.target as HTMLInputElement).value;
  }

  onWalletInput(event: Event) {
    this.model.config.wallet = (event.target as HTMLInputElement).value;
  }

  onTlsChange(event: Event) {
    this.model.config.tls = (event.target as HTMLInputElement).checked;
  }

  onHugePagesChange(event: Event) {
    this.model.config.hugePages = (event.target as HTMLInputElement).checked;
  }

  createProfile() {
    this.error = null;
    this.success = null;

    // Basic validation check
    if (!this.model.name || !this.model.minerType || !this.model.config.pool || !this.model.config.wallet) {
      this.error = 'Please fill out all required fields.';
      return;
    }

    this.isCreating.set(true);
    this.minerService.createProfile(this.model).subscribe({
      next: () => {
        this.isCreating.set(false);
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
        this.isCreating.set(false);
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
