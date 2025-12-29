import { Page, Locator } from '@playwright/test';

export class ProfilesPageNew {
  readonly page: Page;

  // Main container
  readonly profilesPage: Locator;

  // Header
  readonly pageTitle: Locator;
  readonly pageDescription: Locator;
  readonly newProfileButton: Locator;

  // Create form
  readonly createFormContainer: Locator;
  readonly profileCreateComponent: Locator;

  // Profile cards
  readonly profilesGrid: Locator;
  readonly profileCards: Locator;

  // Empty state
  readonly emptyState: Locator;
  readonly emptyStateTitle: Locator;
  readonly createFirstProfileButton: Locator;

  constructor(page: Page) {
    this.page = page;

    // Main container
    this.profilesPage = page.locator('.profiles-page');

    // Header
    this.pageTitle = page.getByRole('heading', { name: 'Mining Profiles' });
    this.pageDescription = page.getByText('Manage your mining configurations');
    this.newProfileButton = page.getByRole('button', { name: 'New Profile' });

    // Create form
    this.createFormContainer = page.locator('.create-form-container');
    this.profileCreateComponent = page.locator('snider-mining-profile-create');

    // Profile cards
    this.profilesGrid = page.locator('.profiles-grid');
    this.profileCards = page.locator('.profile-card');

    // Empty state
    this.emptyState = page.locator('.empty-state');
    this.emptyStateTitle = page.getByRole('heading', { name: 'No Profiles Yet' });
    this.createFirstProfileButton = page.getByRole('button', { name: 'Create Your First Profile' });
  }

  async isVisible(): Promise<boolean> {
    // Use the page title as indicator since CSS classes may not pierce shadow DOM
    return await this.pageTitle.isVisible();
  }

  async clickNewProfile() {
    await this.newProfileButton.click();
  }

  async isCreateFormVisible(): Promise<boolean> {
    return await this.createFormContainer.isVisible();
  }

  async hasProfiles(): Promise<boolean> {
    return await this.profileCards.count() > 0;
  }

  async getProfileCount(): Promise<number> {
    return await this.profileCards.count();
  }

  async getProfileNames(): Promise<string[]> {
    return await this.profileCards.locator('.profile-info h3').allTextContents();
  }

  async getProfileMinerTypes(): Promise<string[]> {
    return await this.profileCards.locator('.profile-miner').allTextContents();
  }

  async getProfileCard(profileName: string): Locator {
    return this.profileCards.filter({ hasText: profileName });
  }

  async isProfileRunning(profileName: string): Promise<boolean> {
    const card = await this.getProfileCard(profileName);
    const runningBadge = card.locator('.running-badge');
    return await runningBadge.isVisible();
  }

  async getProfilePool(profileName: string): Promise<string> {
    const card = await this.getProfileCard(profileName);
    const poolRow = card.locator('.detail-row').filter({ hasText: 'Pool' });
    return await poolRow.locator('.detail-value').textContent() ?? '';
  }

  async getProfileWallet(profileName: string): Promise<string> {
    const card = await this.getProfileCard(profileName);
    const walletRow = card.locator('.detail-row').filter({ hasText: 'Wallet' });
    return await walletRow.locator('.detail-value').textContent() ?? '';
  }

  async clickStartProfile(profileName: string) {
    const card = await this.getProfileCard(profileName);
    const startBtn = card.getByRole('button', { name: 'Start' });
    await startBtn.click();
  }

  async clickStopProfile(profileName: string) {
    const card = await this.getProfileCard(profileName);
    const stopBtn = card.getByRole('button', { name: 'Stop' });
    await stopBtn.click();
  }

  async clickDeleteProfile(profileName: string) {
    const card = await this.getProfileCard(profileName);
    const deleteBtn = card.locator('.action-btn.delete');
    await deleteBtn.click();
  }

  async isStartButtonVisible(profileName: string): Promise<boolean> {
    const card = await this.getProfileCard(profileName);
    const startBtn = card.getByRole('button', { name: 'Start' });
    return await startBtn.isVisible();
  }

  async isStopButtonVisible(profileName: string): Promise<boolean> {
    const card = await this.getProfileCard(profileName);
    const stopBtn = card.getByRole('button', { name: 'Stop' });
    return await stopBtn.isVisible();
  }

  async isDeleteButtonDisabled(profileName: string): Promise<boolean> {
    const card = await this.getProfileCard(profileName);
    const deleteBtn = card.locator('.action-btn.delete');
    return await deleteBtn.isDisabled();
  }

  async isEmpty(): Promise<boolean> {
    return await this.emptyStateTitle.isVisible();
  }

  async clickCreateFirstProfile() {
    await this.createFirstProfileButton.click();
  }
}
