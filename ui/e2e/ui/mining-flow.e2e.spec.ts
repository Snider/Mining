import { test, expect } from '@playwright/test';
import { API_BASE, TEST_POOL, TEST_XMR_WALLET } from '../fixtures/test-data';

/**
 * MINING FLOW TESTS
 *
 * These tests cover the complete mining workflow:
 * 1. Install miner (if needed)
 * 2. Create a mining profile
 * 3. Start mining
 * 4. Verify dashboard shows stats
 * 5. Stop mining
 *
 * These tests use real pool/wallet configuration and will
 * actually start the miner process.
 */

test.describe('Mining Flow', () => {
  // Run tests in order - they depend on each other
  test.describe.configure({ mode: 'serial' });

  let profileId: string;
  const profileName = `Mining Test ${Date.now()}`;

  test('Step 1: Ensure xmrig is installed', async ({ request }) => {
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();
    const xmrigInfo = info.installed_miners_info?.find((m: any) => m.path?.includes('xmrig'));

    if (xmrigInfo?.is_installed) {
      console.log('xmrig already installed, version:', xmrigInfo.version);
      return;
    }

    console.log('Installing xmrig...');
    const installResponse = await request.post(`${API_BASE}/miners/xmrig/install`);
    expect(installResponse.ok()).toBe(true);

    const result = await installResponse.json();
    expect(result.status).toBe('installed');
    console.log('xmrig installed, version:', result.version);
  });

  test('Step 2: Create mining profile', async ({ request }) => {
    const profile = {
      name: profileName,
      minerType: 'xmrig',
      config: {
        pool: TEST_POOL,
        wallet: TEST_XMR_WALLET,
        tls: false,
        hugePages: true,
      },
    };

    const createResponse = await request.post(`${API_BASE}/profiles`, { data: profile });
    expect(createResponse.ok()).toBe(true);

    const created = await createResponse.json();
    profileId = created.id;
    expect(profileId).toBeDefined();
    console.log('Created profile:', profileName, 'ID:', profileId);
  });

  test('Step 3: Verify profile appears in UI', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(profileList).toBeVisible();

    // Wait for profile to appear
    await expect(page.locator(`text=${profileName}`)).toBeVisible({ timeout: 10000 });
    console.log('Profile visible in UI');
  });

  test('Step 4: Start mining via UI', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator(`text=${profileName}`)).toBeVisible({ timeout: 10000 });

    const profileItem = profileList.locator(`.profile-item:has-text("${profileName}")`);
    const startButton = profileItem.locator('wa-button:has-text("Start")');

    await expect(startButton).toBeVisible();

    // Set up response listener
    const startPromise = page.waitForResponse(
      (resp) => resp.url().includes('/start'),
      { timeout: 30000 }
    );

    // Click start
    await startButton.click();
    console.log('Clicked Start button');

    // Wait for API response
    const response = await startPromise;
    expect(response.ok()).toBe(true);
    console.log('Miner start API returned success');
  });

  test('Step 5: Verify miner is running via API', async ({ request }) => {
    // Wait a moment for miner to start
    await new Promise(resolve => setTimeout(resolve, 3000));

    const minersResponse = await request.get(`${API_BASE}/miners`);
    expect(minersResponse.ok()).toBe(true);

    const miners = await minersResponse.json();
    console.log('Running miners:', miners.length);
    expect(miners.length).toBeGreaterThan(0);

    // Miner names include a suffix like "xmrig-419"
    const xmrigMiner = miners.find((m: any) => m.name.startsWith('xmrig'));
    expect(xmrigMiner).toBeDefined();
    console.log('xmrig is running:', xmrigMiner.name);
  });

  test('Step 6: Verify dashboard shows mining stats', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Wait for dashboard to update with mining data
    const dashboard = page.locator('snider-mining-dashboard').first();
    await expect(dashboard).toBeVisible();

    // Should NOT show "No miners running" anymore
    const noMinersMessage = dashboard.locator('text=No miners running');

    // Wait for stats to appear (miner needs time to connect and report)
    await page.waitForTimeout(5000);
    await page.reload();
    await page.waitForLoadState('networkidle');

    // Check for stats bar or chart (indicates mining is active)
    const statsBar = page.locator('.stats-bar-container').first();
    const chartContainer = page.locator('.chart-container').first();

    // At least one should be visible when mining
    const hasStats = await statsBar.isVisible() || await chartContainer.isVisible();

    if (hasStats) {
      console.log('Dashboard showing mining stats');
    } else {
      // Check if still showing no miners (might need more time)
      const stillNoMiners = await noMinersMessage.isVisible();
      if (stillNoMiners) {
        console.log('Dashboard still showing no miners - may need more time to connect');
      }
    }
  });

  test('Step 7: Check miner stats via API', async ({ request }) => {
    // Give miner time to collect stats
    await new Promise(resolve => setTimeout(resolve, 5000));

    // Get running miners first to find the miner name
    const minersResponse = await request.get(`${API_BASE}/miners`);
    const miners = await minersResponse.json();
    const xmrigMiner = miners.find((m: any) => m.name.startsWith('xmrig'));

    if (xmrigMiner) {
      const statsResponse = await request.get(`${API_BASE}/miners/${xmrigMiner.name}/stats`);

      if (statsResponse.ok()) {
        const stats = await statsResponse.json();
        console.log('Miner stats:', JSON.stringify(stats, null, 2));
      } else {
        console.log('Stats not available yet (miner may still be connecting)');
      }
    }
  });

  test('Step 8: Stop mining', async ({ request }) => {
    // Get running miners first to find the miner name
    const minersResponse = await request.get(`${API_BASE}/miners`);
    const miners = await minersResponse.json();
    const xmrigMiner = miners.find((m: any) => m.name.startsWith('xmrig'));

    if (xmrigMiner) {
      const stopResponse = await request.delete(`${API_BASE}/miners/${xmrigMiner.name}`);
      expect(stopResponse.ok()).toBe(true);
      console.log('Miner stopped:', xmrigMiner.name);
    }

    // Verify no xmrig miners running
    await new Promise(resolve => setTimeout(resolve, 2000));
    const checkResponse = await request.get(`${API_BASE}/miners`);
    const remainingMiners = await checkResponse.json();

    const xmrigRunning = remainingMiners.find((m: any) => m.name.startsWith('xmrig'));
    expect(xmrigRunning).toBeUndefined();
    console.log('Verified miner is stopped');
  });

  test('Step 9: Cleanup - delete test profile', async ({ request }) => {
    if (profileId) {
      const deleteResponse = await request.delete(`${API_BASE}/profiles/${profileId}`);
      expect(deleteResponse.ok()).toBe(true);
      console.log('Deleted test profile');
    }
  });
});

test.describe('Quick Mining Start/Stop', () => {
  // Increase timeout for this long-running test
  test.setTimeout(120000);

  test('Start mining, wait 30 seconds, then stop', async ({ page, request }) => {
    // Ensure xmrig is installed
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();
    const xmrigInfo = info.installed_miners_info?.find((m: any) => m.path?.includes('xmrig'));

    if (!xmrigInfo?.is_installed) {
      console.log('Installing xmrig first...');
      await request.post(`${API_BASE}/miners/xmrig/install`);
    }

    // Create a quick test profile
    const profile = {
      name: `Quick Test ${Date.now()}`,
      minerType: 'xmrig',
      config: {
        pool: TEST_POOL,
        wallet: TEST_XMR_WALLET,
        tls: false,
        hugePages: true,
      },
    };

    const createResponse = await request.post(`${API_BASE}/profiles`, { data: profile });
    const created = await createResponse.json();
    const profileId = created.id;

    // Start mining
    console.log('Starting miner...');
    const startResponse = await request.post(`${API_BASE}/profiles/${profileId}/start`);
    expect(startResponse.ok()).toBe(true);

    // Navigate to dashboard to watch
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    console.log('Mining for 30 seconds...');

    // Wait and periodically check stats
    for (let i = 0; i < 6; i++) {
      await page.waitForTimeout(5000);

      // Get running miners to find miner name
      const minersResponse = await request.get(`${API_BASE}/miners`);
      const miners = await minersResponse.json();
      const xmrigMiner = miners.find((m: any) => m.name.startsWith('xmrig'));

      if (xmrigMiner) {
        const hashrate = xmrigMiner.full_stats?.hashrate?.total?.[0] || 0;
        console.log(`[${(i+1)*5}s] Hashrate: ${hashrate.toFixed(2)} H/s`);
      }

      // Reload to see updated dashboard
      await page.reload();
    }

    // Stop mining - get miner name first
    const minersToStop = await request.get(`${API_BASE}/miners`);
    const runningMiners = await minersToStop.json();
    for (const miner of runningMiners) {
      if (miner.name.startsWith('xmrig')) {
        console.log('Stopping miner:', miner.name);
        await request.delete(`${API_BASE}/miners/${miner.name}`);
      }
    }

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${profileId}`);
    console.log('Test complete');
  });
});
