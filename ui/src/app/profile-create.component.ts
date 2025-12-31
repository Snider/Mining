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

  /**
   * Validates input for potential security issues (shell injection, etc.)
   */
  private validateInput(value: string, fieldName: string, maxLength: number): string | null {
    if (!value || value.length === 0) {
      return `${fieldName} is required`;
    }
    if (value.length > maxLength) {
      return `${fieldName} is too long (max ${maxLength} characters)`;
    }
    // Check for shell metacharacters that could enable injection
    const dangerousChars = /[;&|`$(){}\\<>'\"\n\r!]/;
    if (dangerousChars.test(value)) {
      return `${fieldName} contains invalid characters`;
    }
    return null;
  }

  /**
   * Validates pool URL format
   */
  private validatePoolUrl(url: string): string | null {
    if (!url) {
      return 'Pool URL is required';
    }
    const validPrefixes = ['stratum+tcp://', 'stratum+ssl://', 'stratum://'];
    if (!validPrefixes.some(prefix => url.startsWith(prefix))) {
      return 'Pool URL must start with stratum+tcp://, stratum+ssl://, or stratum://';
    }
    return this.validateInput(url, 'Pool URL', 256);
  }

  createProfile() {
    this.error = null;
    this.success = null;

    // Validate all inputs
    const nameError = this.validateInput(this.model.name, 'Profile name', 100);
    if (nameError) {
      this.error = nameError;
      return;
    }

    if (!this.model.minerType) {
      this.error = 'Please select a miner type';
      return;
    }

    const poolError = this.validatePoolUrl(this.model.config.pool);
    if (poolError) {
      this.error = poolError;
      return;
    }

    const walletError = this.validateInput(this.model.config.wallet, 'Wallet address', 256);
    if (walletError) {
      this.error = walletError;
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
