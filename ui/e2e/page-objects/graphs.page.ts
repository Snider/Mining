import { Page, Locator } from '@playwright/test';

export class GraphsPage {
  readonly page: Page;

  // Main container
  readonly graphsPage: Locator;

  // Chart section
  readonly chartContainer: Locator;
  readonly chartTitle: Locator;
  readonly chartEmptyState: Locator;
  readonly chartEmptyIcon: Locator;
  readonly chartEmptyMessage: Locator;

  // Stats cards
  readonly statsGrid: Locator;
  readonly statCards: Locator;
  readonly peakHashrateStat: Locator;
  readonly efficiencyStat: Locator;
  readonly avgShareTimeStat: Locator;
  readonly difficultyStat: Locator;

  constructor(page: Page) {
    this.page = page;

    // Main container
    this.graphsPage = page.locator('.graphs-page');

    // Chart section
    this.chartContainer = page.locator('.chart-card');
    this.chartTitle = page.getByRole('heading', { name: 'Hashrate Over Time' });
    this.chartEmptyState = page.locator('.chart-empty');
    this.chartEmptyIcon = page.locator('.chart-empty svg');
    this.chartEmptyMessage = page.getByText('Start mining to see hashrate graphs');

    // Stats cards
    this.statsGrid = page.locator('.stats-grid');
    this.statCards = page.locator('.stat-card');
    this.peakHashrateStat = page.getByText('Peak Hashrate').locator('..');
    this.efficiencyStat = page.locator('.stat-card').filter({ hasText: 'Efficiency' });
    this.avgShareTimeStat = page.getByText('Avg. Share Time').locator('..');
    this.difficultyStat = page.locator('.stat-card').filter({ hasText: 'Difficulty' });
  }

  async isVisible(): Promise<boolean> {
    // Use the chart title as the indicator since CSS classes may not pierce shadow DOM
    return await this.chartTitle.isVisible();
  }

  async isChartEmpty(): Promise<boolean> {
    return await this.chartEmptyMessage.isVisible();
  }

  async getPeakHashrate(): Promise<string> {
    const valueEl = this.peakHashrateStat.locator('.stat-value');
    return await valueEl.textContent() ?? '';
  }

  async getEfficiency(): Promise<string> {
    const valueEl = this.efficiencyStat.locator('.stat-value');
    return await valueEl.textContent() ?? '';
  }

  async getAvgShareTime(): Promise<string> {
    const valueEl = this.avgShareTimeStat.locator('.stat-value');
    return await valueEl.textContent() ?? '';
  }

  async getDifficulty(): Promise<string> {
    const valueEl = this.difficultyStat.locator('.stat-value');
    return await valueEl.textContent() ?? '';
  }

  async getStatsCardCount(): Promise<number> {
    // Count by checking for known stat labels
    let count = 0;
    const labels = ['Peak Hashrate', 'Efficiency', 'Avg. Share Time', 'Difficulty'];
    for (const label of labels) {
      const labelEl = this.page.getByText(label);
      if (await labelEl.isVisible()) count++;
    }
    return count;
  }
}
