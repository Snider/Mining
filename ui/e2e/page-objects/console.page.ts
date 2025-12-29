import { Page, Locator } from '@playwright/test';

export class ConsolePage {
  readonly page: Page;

  // Main container
  readonly consolePage: Locator;

  // Tabs
  readonly tabsContainer: Locator;
  readonly minerTabs: Locator;
  readonly noMinersTab: Locator;

  // Console output
  readonly consoleOutput: Locator;
  readonly logLines: Locator;
  readonly emptyState: Locator;
  readonly emptyMessage: Locator;

  // Controls
  readonly autoScrollCheckbox: Locator;
  readonly clearButton: Locator;

  constructor(page: Page) {
    this.page = page;

    // Main container
    this.consolePage = page.locator('.console-page');

    // Header (contains worker chooser or tabs)
    this.tabsContainer = page.locator('.console-header');
    this.minerTabs = page.locator('.tab-btn');
    this.noMinersTab = page.getByText('No active workers');

    // Console output
    this.consoleOutput = page.locator('.console-output');
    this.logLines = page.locator('.log-line');
    this.emptyState = page.locator('.console-empty');
    this.emptyMessage = page.getByText('Start a miner to see console output');

    // Controls
    this.autoScrollCheckbox = page.locator('.control-checkbox input[type="checkbox"]');
    this.clearButton = page.getByRole('button', { name: 'Clear' });
  }

  async isVisible(): Promise<boolean> {
    // Use the tabs container or clear button as indicator since CSS classes may not pierce shadow DOM
    return await this.tabsContainer.isVisible() || await this.clearButton.isVisible();
  }

  async hasActiveMiners(): Promise<boolean> {
    return await this.minerTabs.count() > 0;
  }

  async getMinerTabCount(): Promise<number> {
    return await this.minerTabs.count();
  }

  async selectMinerTab(minerName: string) {
    const tab = this.minerTabs.filter({ hasText: minerName });
    await tab.click();
  }

  async getSelectedMinerTab(): Promise<string | null> {
    const activeTab = this.page.locator('.tab-btn.active');
    if (await activeTab.isVisible()) {
      return await activeTab.textContent();
    }
    return null;
  }

  async getLogLineCount(): Promise<number> {
    return await this.logLines.count();
  }

  async getLogContent(): Promise<string[]> {
    return await this.logLines.locator('.log-text').allTextContents();
  }

  async hasErrorLogs(): Promise<boolean> {
    const errorLines = this.logLines.locator('.error');
    return await errorLines.count() > 0;
  }

  async hasWarningLogs(): Promise<boolean> {
    const warningLines = this.logLines.locator('.warning');
    return await warningLines.count() > 0;
  }

  async toggleAutoScroll() {
    await this.autoScrollCheckbox.click();
  }

  async isAutoScrollEnabled(): Promise<boolean> {
    return await this.autoScrollCheckbox.isChecked();
  }

  async clearLogs() {
    await this.clearButton.click();
  }

  async isClearButtonEnabled(): Promise<boolean> {
    return await this.clearButton.isEnabled();
  }

  async isConsoleEmpty(): Promise<boolean> {
    return await this.emptyState.isVisible();
  }
}
