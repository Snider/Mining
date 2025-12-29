import { Page, Locator } from '@playwright/test';

export class WorkersPage {
  readonly page: Page;

  // Profile selector
  readonly profileSelect: Locator;
  readonly startButton: Locator;
  readonly stopAllButton: Locator;

  // Empty state
  readonly emptyState: Locator;
  readonly emptyStateIcon: Locator;
  readonly emptyStateTitle: Locator;
  readonly emptyStateDescription: Locator;

  // Workers table
  readonly workersTable: Locator;
  readonly workersTableRows: Locator;

  // Firewall warning
  readonly firewallWarning: Locator;
  readonly dismissWarningButton: Locator;

  // Terminal modal
  readonly terminalModal: Locator;

  constructor(page: Page) {
    this.page = page;

    // Profile selector
    this.profileSelect = page.locator('.profile-select');
    this.startButton = page.getByRole('button', { name: 'Start' }).first();
    this.stopAllButton = page.getByRole('button', { name: 'Stop All' });

    // Empty state
    this.emptyState = page.locator('.empty-state');
    this.emptyStateIcon = page.locator('.empty-icon');
    this.emptyStateTitle = page.getByRole('heading', { name: 'No Active Workers' });
    this.emptyStateDescription = page.getByText('Select a profile and start mining to see workers here.');

    // Workers table
    this.workersTable = page.locator('.workers-table');
    this.workersTableRows = page.locator('.workers-table tbody tr');

    // Firewall warning
    this.firewallWarning = page.locator('.warning-banner');
    this.dismissWarningButton = page.locator('.dismiss-btn');

    // Terminal modal
    this.terminalModal = page.locator('app-terminal-modal');
  }

  async isVisible(): Promise<boolean> {
    return await this.page.locator('.workers-page').isVisible();
  }

  async selectProfile(profileName: string) {
    await this.profileSelect.selectOption({ label: profileName });
  }

  async getSelectedProfile(): Promise<string> {
    return await this.profileSelect.inputValue();
  }

  async getProfileOptions(): Promise<string[]> {
    const options = this.profileSelect.locator('option:not([disabled])');
    return await options.allTextContents();
  }

  async startMining() {
    await this.startButton.click();
  }

  async stopAllMiners() {
    await this.stopAllButton.click();
  }

  async isStartButtonEnabled(): Promise<boolean> {
    return await this.startButton.isEnabled();
  }

  async hasRunningWorkers(): Promise<boolean> {
    return await this.workersTable.isVisible();
  }

  async getWorkerCount(): Promise<number> {
    if (!(await this.hasRunningWorkers())) {
      return 0;
    }
    return await this.workersTableRows.count();
  }

  async getWorkerNames(): Promise<string[]> {
    const nameCells = this.workersTableRows.locator('.worker-name span').first();
    return await nameCells.allTextContents();
  }

  async clickWorkerTerminal(workerName: string) {
    const row = this.workersTableRows.filter({ hasText: workerName });
    const terminalBtn = row.locator('.icon-btn').first();
    await terminalBtn.click();
  }

  async clickStopWorker(workerName: string) {
    const row = this.workersTableRows.filter({ hasText: workerName });
    const stopBtn = row.locator('.icon-btn-danger');
    await stopBtn.click();
  }

  async dismissFirewallWarning() {
    await this.dismissWarningButton.click();
  }

  async isFirewallWarningVisible(): Promise<boolean> {
    return await this.firewallWarning.isVisible();
  }

  async isTerminalModalVisible(): Promise<boolean> {
    return await this.terminalModal.isVisible();
  }

  async closeTerminalModal() {
    const closeBtn = this.terminalModal.locator('.close-btn');
    await closeBtn.click();
  }
}
