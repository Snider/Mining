import { test, expect } from '@playwright/test';
import { API_BASE } from '../fixtures/test-data';

test.describe('Setup Wizard Component', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('displays setup wizard', async ({ page }) => {
    const wizard = page.locator('snider-mining-setup-wizard');
    await expect(wizard).toBeVisible();
  });

  test('shows setup required header', async ({ page }) => {
    await expect(
      page.locator('snider-mining-setup-wizard .header-title:has-text("Setup Required")')
    ).toBeVisible();
  });

  test('shows available miners heading', async ({ page }) => {
    await expect(
      page.locator('snider-mining-setup-wizard h4:has-text("Available Miners")')
    ).toBeVisible();
  });

  test('displays available miners for installation', async ({ page, request }) => {
    const availableResponse = await request.get(`${API_BASE}/miners/available`);
    const available = await availableResponse.json();

    for (const miner of available) {
      await expect(
        page.locator(`snider-mining-setup-wizard .miner-item:has-text("${miner.name}")`)
      ).toBeVisible();
    }
  });

  test('shows install button for non-installed miners', async ({ page, request }) => {
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();

    const xmrigInfo = info.installed_miners_info?.find((m: { miner_binary?: string }) =>
      m.miner_binary?.includes('xmrig')
    );

    const xmrigItem = page.locator('snider-mining-setup-wizard .miner-item:has-text("xmrig")');

    if (!xmrigInfo?.is_installed) {
      await expect(xmrigItem.locator('wa-button:has-text("Install")')).toBeVisible();
    }
  });
});
