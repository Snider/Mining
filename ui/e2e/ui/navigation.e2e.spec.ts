import { test, expect } from '@playwright/test';
import { MainLayoutPage } from '../page-objects/main-layout.page';

test.describe('Navigation', () => {
  let layout: MainLayoutPage;

  test.beforeEach(async ({ page }) => {
    layout = new MainLayoutPage(page);
    await layout.goto();
    await layout.waitForLayoutLoad();
  });

  test.describe('Sidebar Navigation', () => {
    test('should display sidebar with all navigation items', async () => {
      // Check all navigation buttons are visible
      await expect(layout.workersNavBtn).toBeVisible();
      await expect(layout.graphsNavBtn).toBeVisible();
      await expect(layout.consoleNavBtn).toBeVisible();
      await expect(layout.poolsNavBtn).toBeVisible();
      await expect(layout.profilesNavBtn).toBeVisible();
      await expect(layout.minersNavBtn).toBeVisible();
    });

    test('should show logo text when sidebar is expanded', async () => {
      await expect(layout.sidebarLogo).toBeVisible();
    });

    test('should show mining status indicator', async () => {
      await expect(layout.miningStatus).toBeVisible();
    });

    test('should collapse sidebar when clicking collapse button', async ({ page }) => {
      // Sidebar should start expanded
      const sidebar = page.locator('.sidebar');
      await expect(sidebar).not.toHaveClass(/collapsed/);

      // Click collapse button
      await layout.toggleSidebarCollapse();

      // Sidebar should now be collapsed
      await expect(sidebar).toHaveClass(/collapsed/);

      // Logo text should be hidden
      await expect(layout.sidebarLogo).not.toBeVisible();
    });

    test('should expand sidebar when clicking collapse button again', async ({ page }) => {
      // Collapse first
      await layout.toggleSidebarCollapse();

      // Expand
      await layout.toggleSidebarCollapse();

      // Sidebar should be expanded
      const sidebar = page.locator('.sidebar');
      await expect(sidebar).not.toHaveClass(/collapsed/);
    });

    test('should navigate to Workers page', async ({ page }) => {
      await layout.navigateToWorkers();
      await expect(page.locator('.workers-page')).toBeVisible();
    });

    test('should navigate to Graphs page', async ({ page }) => {
      await layout.navigateToGraphs();
      await expect(page.getByRole('heading', { name: 'Hashrate Over Time' })).toBeVisible();
    });

    test('should navigate to Console page', async ({ page }) => {
      await layout.navigateToConsole();
      await expect(page.locator('.console-page')).toBeVisible();
    });

    test('should navigate to Pools page', async ({ page }) => {
      await layout.navigateToPools();
      await expect(page.getByRole('heading', { name: 'Mining Pools' })).toBeVisible();
    });

    test('should navigate to Profiles page', async ({ page }) => {
      await layout.navigateToProfiles();
      await expect(page.getByRole('heading', { name: 'Mining Profiles' })).toBeVisible();
    });

    test('should navigate to Miners page', async ({ page }) => {
      await layout.navigateToMiners();
      await expect(page.getByRole('heading', { name: 'Miner Software' })).toBeVisible();
    });

    test('should highlight active navigation item', async ({ page }) => {
      // Navigate to Profiles
      await layout.navigateToProfiles();

      // Check that Profiles nav item is active
      const profilesBtn = layout.profilesNavBtn;
      await expect(profilesBtn).toHaveClass(/active/);

      // Navigate to Miners
      await layout.navigateToMiners();

      // Profiles should no longer be active
      await expect(profilesBtn).not.toHaveClass(/active/);

      // Miners should be active
      await expect(layout.minersNavBtn).toHaveClass(/active/);
    });

    test('should default to Workers page on load', async ({ page }) => {
      await expect(page.locator('.workers-page')).toBeVisible();
      await expect(layout.workersNavBtn).toHaveClass(/active/);
    });
  });

  test.describe('Stats Panel', () => {
    test('should display stats panel', async () => {
      await expect(layout.statsPanel).toBeVisible();
    });

    test('should display hashrate stat', async () => {
      await expect(layout.hashratestat).toBeVisible();
    });

    test('should display pool connection stat', async () => {
      await expect(layout.poolStat).toBeVisible();
    });

    test('stats panel should persist across navigation', async ({ page }) => {
      // Navigate through all pages and verify stats panel is visible
      const pages = [
        () => layout.navigateToWorkers(),
        () => layout.navigateToGraphs(),
        () => layout.navigateToConsole(),
        () => layout.navigateToPools(),
        () => layout.navigateToProfiles(),
        () => layout.navigateToMiners(),
      ];

      for (const navigate of pages) {
        await navigate();
        await expect(layout.statsPanel).toBeVisible();
      }
    });
  });

  test.describe('Navigation Interaction', () => {
    test('should navigate between all pages without errors', async ({ page }) => {
      // Navigate through all pages in sequence
      await layout.navigateToGraphs();
      await expect(page.getByRole('heading', { name: 'Hashrate Over Time' })).toBeVisible();

      await layout.navigateToConsole();
      await expect(page.locator('.console-page')).toBeVisible();

      await layout.navigateToPools();
      await expect(page.getByRole('heading', { name: 'Mining Pools' })).toBeVisible();

      await layout.navigateToProfiles();
      await expect(page.getByRole('heading', { name: 'Mining Profiles' })).toBeVisible();

      await layout.navigateToMiners();
      await expect(page.getByRole('heading', { name: 'Miner Software' })).toBeVisible();

      await layout.navigateToWorkers();
      await expect(page.locator('.workers-page')).toBeVisible();
    });

    test('should be able to navigate when sidebar is collapsed', async ({ page }) => {
      // Collapse sidebar
      await layout.toggleSidebarCollapse();

      // Navigate should still work
      await layout.navigateToProfiles();
      await expect(page.getByRole('heading', { name: 'Mining Profiles' })).toBeVisible();

      await layout.navigateToMiners();
      await expect(page.getByRole('heading', { name: 'Miner Software' })).toBeVisible();
    });
  });
});
