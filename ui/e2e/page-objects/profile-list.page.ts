import { Page, Locator } from '@playwright/test';

export class ProfileListPage {
  readonly page: Page;
  readonly container: Locator;
  readonly profileItems: Locator;
  readonly noProfilesMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.container = page.locator('snider-mining-profile-list');
    this.profileItems = page.locator('snider-mining-profile-list .profile-item');
    this.noProfilesMessage = page.locator('snider-mining-profile-list >> text=No profiles created yet');
  }

  async getProfileCount(): Promise<number> {
    return await this.profileItems.count();
  }

  async getProfileByName(name: string): Locator {
    return this.container.locator(`.profile-item:has-text("${name}")`);
  }

  async clickStartButton(profileName: string) {
    const profileItem = await this.getProfileByName(profileName);
    await profileItem.locator('wa-button:has-text("Start")').click();
  }

  async clickEditButton(profileName: string) {
    const profileItem = await this.getProfileByName(profileName);
    await profileItem.locator('wa-button:has-text("Edit")').click();
  }

  async clickDeleteButton(profileName: string) {
    const profileItem = await this.getProfileByName(profileName);
    await profileItem.locator('wa-button:has-text("Delete")').click();
  }

  async clickSaveButton() {
    await this.container.locator('wa-button:has-text("Save")').click();
  }

  async clickCancelButton() {
    await this.container.locator('wa-button:has-text("Cancel")').click();
  }

  async waitForProfileListLoad() {
    await this.container.waitFor({ state: 'visible' });
  }
}
