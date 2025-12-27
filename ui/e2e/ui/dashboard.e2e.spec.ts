import { test, expect } from '@playwright/test';
import { DashboardPage } from '../page-objects/dashboard.page';

test.describe('Dashboard Component', () => {
  let dashboardPage: DashboardPage;

  test.beforeEach(async ({ page }) => {
    dashboardPage = new DashboardPage(page);
    await dashboardPage.goto();
  });

  test('loads and displays dashboard', async () => {
    await dashboardPage.waitForDashboardLoad();
    await expect(dashboardPage.dashboard).toBeVisible();
  });

  test('shows "no miners running" when no miners are active', async () => {
    await dashboardPage.waitForDashboardLoad();

    const hasMiners = await dashboardPage.hasRunningMiners();

    if (!hasMiners) {
      await expect(dashboardPage.noMinersMessage).toBeVisible();
    }
  });

  test('displays stats bar when miners are running', async ({ page }) => {
    await dashboardPage.waitForDashboardLoad();

    const hasMiners = await dashboardPage.hasRunningMiners();

    test.skip(!hasMiners, 'No miners running - skipping stats display test');

    await expect(dashboardPage.statsBarContainer).toBeVisible();
  });

  test('displays chart container when miners are running', async ({ page }) => {
    await dashboardPage.waitForDashboardLoad();

    const hasMiners = await dashboardPage.hasRunningMiners();

    test.skip(!hasMiners, 'No miners running - skipping chart display test');

    await expect(dashboardPage.chartContainer).toBeVisible();
  });
});
