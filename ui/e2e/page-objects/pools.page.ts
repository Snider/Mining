import { Page, Locator } from '@playwright/test';

export class PoolsPage {
  readonly page: Page;

  // Main container
  readonly poolsPage: Locator;

  // Header
  readonly pageTitle: Locator;
  readonly pageDescription: Locator;

  // Pool cards
  readonly poolsGrid: Locator;
  readonly poolCards: Locator;

  // Empty state
  readonly emptyState: Locator;
  readonly emptyStateTitle: Locator;

  constructor(page: Page) {
    this.page = page;

    // Main container
    this.poolsPage = page.locator('.pools-page');

    // Header
    this.pageTitle = page.getByRole('heading', { name: 'Mining Pools' });
    this.pageDescription = page.getByText('Active pool connections from running miners');

    // Pool cards
    this.poolsGrid = page.locator('.pools-grid');
    this.poolCards = page.locator('.pool-card');

    // Empty state
    this.emptyState = page.locator('.empty-state');
    this.emptyStateTitle = page.getByRole('heading', { name: 'No Pool Connections' });
  }

  async isVisible(): Promise<boolean> {
    // Use the page title as indicator since CSS classes may not pierce shadow DOM
    return await this.pageTitle.isVisible();
  }

  async hasPoolConnections(): Promise<boolean> {
    return await this.poolCards.count() > 0;
  }

  async getPoolCount(): Promise<number> {
    return await this.poolCards.count();
  }

  async getPoolNames(): Promise<string[]> {
    return await this.poolCards.locator('.pool-name').allTextContents();
  }

  async getPoolHosts(): Promise<string[]> {
    return await this.poolCards.locator('.pool-host').allTextContents();
  }

  async getPoolPings(): Promise<string[]> {
    return await this.poolCards.locator('.pool-ping').allTextContents();
  }

  async isPoolConnected(poolHost: string): Promise<boolean> {
    const card = this.poolCards.filter({ hasText: poolHost });
    const classes = await card.getAttribute('class');
    return classes?.includes('connected') ?? false;
  }

  async getPoolMinerBadges(poolHost: string): Promise<string[]> {
    const card = this.poolCards.filter({ hasText: poolHost });
    return await card.locator('.miner-badge').allTextContents();
  }

  async isEmpty(): Promise<boolean> {
    return await this.emptyStateTitle.isVisible();
  }
}
