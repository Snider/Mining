import { Page, Locator } from '@playwright/test';

export class DashboardPage {
  readonly page: Page;
  readonly dashboard: Locator;
  readonly statsBarContainer: Locator;
  readonly statsListContainer: Locator;
  readonly chartContainer: Locator;
  readonly noMinersMessage: Locator;
  readonly errorCard: Locator;

  constructor(page: Page) {
    this.page = page;
    this.dashboard = page.locator('snider-mining-dashboard').first();
    this.statsBarContainer = page.locator('.quick-stats').first();
    this.statsListContainer = page.locator('.stats-list-container').first();
    this.chartContainer = page.locator('.chart-container').first();
    this.noMinersMessage = page.locator('text=No miners running').first();
    this.errorCard = page.locator('.card-error').first();
  }

  async goto() {
    await this.page.goto('/');
    await this.page.waitForLoadState('networkidle');
  }

  async waitForDashboardLoad() {
    await this.dashboard.waitFor({ state: 'visible' });
  }

  async hasRunningMiners(): Promise<boolean> {
    const response = await this.page.request.get('http://localhost:9090/api/v1/mining/miners');
    const miners = await response.json();
    return miners.length > 0;
  }
}
