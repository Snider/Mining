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

  deleteProfile(profileId: string) {
    this.minerService.deleteProfile(profileId).subscribe();
  }

  editProfile(profile: MiningProfile) {
    this.editingProfile = { ...profile, config: { ...profile.config } };
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
