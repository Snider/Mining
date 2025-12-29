import { Page, Locator } from '@playwright/test';

export class MainLayoutPage {
  readonly page: Page;

  // Shadow DOM root
  readonly shadowHost: Locator;

  // Sidebar elements
  readonly sidebar: Locator;
  readonly sidebarLogo: Locator;
  readonly collapseButton: Locator;
  readonly workersNavBtn: Locator;
  readonly graphsNavBtn: Locator;
  readonly consoleNavBtn: Locator;
  readonly poolsNavBtn: Locator;
  readonly profilesNavBtn: Locator;
  readonly minersNavBtn: Locator;
  readonly miningStatus: Locator;

  // Stats panel elements
  readonly statsPanel: Locator;
  readonly hashratestat: Locator;
  readonly sharesStat: Locator;
  readonly uptimeStat: Locator;
  readonly poolStat: Locator;
  readonly workersStat: Locator;

  constructor(page: Page) {
    this.page = page;
    this.shadowHost = page.locator('snider-mining');

    // Sidebar - use button role with exact name for navigation (to avoid matching "All Workers" in switcher)
    this.sidebar = page.locator('app-sidebar');
    this.sidebarLogo = page.locator('.logo-text');
    this.collapseButton = page.locator('button.collapse-btn');
    this.workersNavBtn = page.getByRole('button', { name: 'Workers', exact: true });
    this.graphsNavBtn = page.getByRole('button', { name: 'Graphs', exact: true });
    this.consoleNavBtn = page.getByRole('button', { name: 'Console', exact: true });
    this.poolsNavBtn = page.getByRole('button', { name: 'Pools', exact: true });
    this.profilesNavBtn = page.getByRole('button', { name: 'Profiles', exact: true });
    this.minersNavBtn = page.getByRole('button', { name: 'Miners', exact: true });
    this.miningStatus = page.getByText('Mining Active');

    // Stats panel
    this.statsPanel = page.locator('app-stats-panel');
    this.hashratestat = page.getByText('H/s').first();
    this.sharesStat = page.locator('.stat-card').nth(1);
    this.uptimeStat = page.locator('.stat-card').nth(2);
    this.poolStat = page.getByText('Not connected');
    this.workersStat = page.locator('.stat-card').nth(4);
  }

  async goto() {
    await this.page.goto('/');
    await this.page.waitForLoadState('networkidle');
  }

  async waitForLayoutLoad() {
    await this.shadowHost.waitFor({ state: 'visible' });
    // Wait for either main layout or setup wizard
    await this.page.waitForTimeout(1000);
  }

  async isMainLayoutVisible(): Promise<boolean> {
    try {
      const mainLayout = this.page.locator('app-main-layout');
      return await mainLayout.isVisible({ timeout: 2000 });
    } catch {
      return false;
    }
  }

  async navigateToWorkers() {
    await this.workersNavBtn.click();
  }

  async navigateToGraphs() {
    await this.graphsNavBtn.click();
  }

  async navigateToConsole() {
    await this.consoleNavBtn.click();
  }

  async navigateToPools() {
    await this.poolsNavBtn.click();
  }

  async navigateToProfiles() {
    await this.profilesNavBtn.click();
  }

  async navigateToMiners() {
    await this.minersNavBtn.click();
  }

  async toggleSidebarCollapse() {
    await this.collapseButton.click();
  }

  async isSidebarCollapsed(): Promise<boolean> {
    const sidebar = this.page.locator('.sidebar');
    const classes = await sidebar.getAttribute('class');
    return classes?.includes('collapsed') ?? false;
  }

  async getActiveNavItem(): Promise<string> {
    const activeBtn = this.page.locator('.nav-item.active');
    const label = activeBtn.locator('.nav-label');
    return await label.textContent() ?? '';
  }
}
