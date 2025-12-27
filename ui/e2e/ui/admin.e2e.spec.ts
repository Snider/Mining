import { test, expect } from '@playwright/test';
import { API_BASE } from '../fixtures/test-data';

test.describe('Admin Component', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('displays admin panel', async ({ page }) => {
    const adminPanel = page.locator('snider-mining-admin');
    await expect(adminPanel).toBeVisible();
  });

  test('shows "Manage Miners" heading', async ({ page }) => {
    await expect(page.locator('snider-mining-admin h4:has-text("Manage Miners")')).toBeVisible();
  });

  test('displays available miners', async ({ page, request }) => {
    const availableResponse = await request.get(`${API_BASE}/miners/available`);
    const available = await availableResponse.json();

    for (const miner of available) {
      await expect(
        page.locator(`snider-mining-admin .miner-item:has-text("${miner.name}")`)
      ).toBeVisible();
    }
  });

  test('shows install/uninstall buttons based on installation status', async ({ page, request }) => {
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();

    const xmrigInfo = info.installed_miners_info?.find((m: { miner_binary?: string }) =>
      m.miner_binary?.includes('xmrig')
    );

    const xmrigItem = page.locator('snider-mining-admin .miner-item:has-text("xmrig")');

    if (xmrigInfo?.is_installed) {
      await expect(xmrigItem.locator('wa-button:has-text("Uninstall")')).toBeVisible();
    } else {
      await expect(xmrigItem.locator('wa-button:has-text("Install")')).toBeVisible();
    }
  });

  test('displays antivirus whitelist section', async ({ page }) => {
    await expect(
      page.locator('snider-mining-admin h4:has-text("Antivirus Whitelist")')
    ).toBeVisible();
  });
});
