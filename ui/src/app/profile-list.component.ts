import { Component, CUSTOM_ELEMENTS_SCHEMA, inject, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { MinerService, MiningProfile } from './miner.service';

@Component({
  selector: 'snider-mining-profile-list',
  standalone: true,
  imports: [CommonModule, FormsModule],
  schemas: [CUSTOM_ELEMENTS_SCHEMA],
  templateUrl: './profile-list.component.html',
  styleUrls: ['./profile-list.component.css']
})
export class ProfileListComponent {
  minerService = inject(MinerService);
  state = this.minerService.state;

  editingProfile: (MiningProfile & { config: any }) | null = null;

  // --- Event Handlers for Custom Elements in Edit Form ---
  onNameInput(event: Event) {
    if (this.editingProfile) {
      this.editingProfile.name = (event.target as HTMLInputElement).value;
    }
  }

  onMinerTypeChange(event: Event) {
    if (this.editingProfile) {
      this.editingProfile.minerType = (event.target as HTMLSelectElement).value;
    }
  }

  onPoolInput(event: Event) {
    if (this.editingProfile) {
      this.editingProfile.config.pool = (event.target as HTMLInputElement).value;
    }
  }

  onWalletInput(event: Event) {
    if (this.editingProfile) {
      this.editingProfile.config.wallet = (event.target as HTMLInputElement).value;
    }
  }

  onTlsChange(event: Event) {
    if (this.editingProfile) {
      this.editingProfile.config.tls = (event.target as HTMLInputElement).checked;
    }
  }

  onHugePagesChange(event: Event) {
    if (this.editingProfile) {
      this.editingProfile.config.hugePages = (event.target as HTMLInputElement).checked;
    }
  }

  startMiner(profileId: string) {
    this.minerService.startMiner(profileId).subscribe();
  }

  deleteProfile(profileId: string) {
    this.minerService.deleteProfile(profileId).subscribe();
  }

  editProfile(profile: MiningProfile) {
    // Create a deep copy to avoid mutating the original profile object during editing
    this.editingProfile = JSON.parse(JSON.stringify(profile));
  }

  updateProfile() {
    if (!this.editingProfile) return;
    this.minerService.updateProfile(this.editingProfile).subscribe(() => {
      this.editingProfile = null;
    });
  }

  cancelEdit() {
    this.editingProfile = null;
  }
}
