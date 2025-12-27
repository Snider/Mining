import { Page, Locator } from '@playwright/test';

export class ProfileCreatePage {
  readonly page: Page;
  readonly form: Locator;
  readonly nameInput: Locator;
  readonly minerTypeSelect: Locator;
  readonly poolInput: Locator;
  readonly walletInput: Locator;
  readonly tlsCheckbox: Locator;
  readonly hugePagesCheckbox: Locator;
  readonly createButton: Locator;
  readonly successMessage: Locator;
  readonly errorMessage: Locator;

  constructor(page: Page) {
    this.page = page;
    this.form = page.locator('snider-mining-profile-create form');
    this.nameInput = page.locator('snider-mining-profile-create wa-input[name="name"]');
    this.minerTypeSelect = page.locator('snider-mining-profile-create wa-select[name="minerType"]');
    this.poolInput = page.locator('snider-mining-profile-create wa-input[name="pool"]');
    this.walletInput = page.locator('snider-mining-profile-create wa-input[name="wallet"]');
    this.tlsCheckbox = page.locator('snider-mining-profile-create wa-checkbox[name="tls"]');
    this.hugePagesCheckbox = page.locator(
      'snider-mining-profile-create wa-checkbox[name="hugePages"]'
    );
    this.createButton = page.locator('snider-mining-profile-create wa-button[type="submit"]');
    this.successMessage = page.locator('snider-mining-profile-create .card-success');
    this.errorMessage = page.locator('snider-mining-profile-create .card-error');
  }

  async fillProfile(profile: { name: string; minerType: string; pool: string; wallet: string }) {
    // Web Awesome inputs - click and type
    await this.nameInput.click();
    await this.nameInput.pressSequentially(profile.name, { delay: 50 });

    // Select miner type
    await this.minerTypeSelect.click();
    await this.page.locator(`wa-option[value="${profile.minerType}"]`).click();

    // Fill pool
    await this.poolInput.click();
    await this.poolInput.pressSequentially(profile.pool, { delay: 50 });

    // Fill wallet
    await this.walletInput.click();
    await this.walletInput.pressSequentially(profile.wallet, { delay: 50 });
  }

  async submitForm() {
    await this.createButton.click();
  }

  async waitForSuccess() {
    await this.successMessage.waitFor({ state: 'visible', timeout: 5000 });
  }
}
