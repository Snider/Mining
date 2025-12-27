import { test, expect } from '@playwright/test';
import { API_BASE, TEST_POOL, TEST_XMR_WALLET } from '../fixtures/test-data';

/**
 * LONG RUNNING MINING TEST
 *
 * Runs mining for 5 minutes with 30-second interval checks.
 * Logs detailed stats for analysis.
 */

test.describe('Long Running Mining Test', () => {
  test.setTimeout(600000); // 10 minute timeout

  test('Mine for 5 minutes with stats logging', async ({ page, request }) => {
    // Ensure xmrig is installed
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();
    const xmrigInfo = info.installed_miners_info?.find((m: any) => m.path?.includes('xmrig'));

    if (!xmrigInfo?.is_installed) {
      console.log('=== Installing xmrig ===');
      const installResponse = await request.post(`${API_BASE}/miners/xmrig/install`);
      expect(installResponse.ok()).toBe(true);
    }

    // Create test profile
    const profileName = `Long Test ${Date.now()}`;
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

    console.log('=== Creating profile ===');
    console.log(`Pool: ${TEST_POOL}`);
    console.log(`Wallet: ${TEST_XMR_WALLET.substring(0, 20)}...`);

    const createResponse = await request.post(`${API_BASE}/profiles`, { data: profile });
    const created = await createResponse.json();
    const profileId = created.id;

    // Start mining
    console.log('\n=== Starting miner ===');
    const startResponse = await request.post(`${API_BASE}/profiles/${profileId}/start`);
    expect(startResponse.ok()).toBe(true);

    // Wait for miner to initialize
    await page.waitForTimeout(5000);

    // Navigate to dashboard
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    console.log('\n=== Mining for 5 minutes ===\n');

    const startTime = Date.now();
    const stats: any[] = [];

    // 10 intervals of 30 seconds = 5 minutes
    for (let i = 0; i < 10; i++) {
      await page.waitForTimeout(30000);

      const elapsed = Math.round((Date.now() - startTime) / 1000);
      const minersResponse = await request.get(`${API_BASE}/miners`);
      const miners = await minersResponse.json();
      const xmrigMiner = miners.find((m: any) => m.name.startsWith('xmrig'));

      if (xmrigMiner && xmrigMiner.full_stats) {
        const s = xmrigMiner.full_stats;
        const hashrate = s.hashrate?.total?.[0] || 0;
        const shares = s.results?.shares_good || 0;
        const rejected = s.results?.shares_total - s.results?.shares_good || 0;
        const uptime = s.uptime || 0;
        const pool = s.connection?.pool || 'unknown';
        const ping = s.connection?.ping || 0;
        const diff = s.connection?.diff || 0;
        const accepted = s.connection?.accepted || 0;
        const algo = s.algo || 'unknown';
        const cpu = s.cpu?.brand || 'unknown';
        const threads = s.cpu?.threads || 0;
        const memory = s.resources?.memory?.resident_set_memory || 0;
        const memoryMB = Math.round(memory / 1024 / 1024);

        const statEntry = {
          interval: i + 1,
          elapsed,
          hashrate,
          shares,
          rejected,
          accepted,
          uptime,
          ping,
          diff,
          algo,
          memoryMB,
        };
        stats.push(statEntry);

        console.log(`--- Interval ${i + 1}/10 (${elapsed}s elapsed) ---`);
        console.log(`Hashrate: ${hashrate.toFixed(2)} H/s`);
        console.log(`Shares: ${shares} accepted, ${rejected} rejected`);
        console.log(`Pool: ${pool} (ping: ${ping}ms, diff: ${diff})`);
        console.log(`Algorithm: ${algo}`);
        console.log(`Memory: ${memoryMB} MB`);
        console.log(`CPU: ${cpu} (${threads} threads)`);

        // Check hashrate history
        if (xmrigMiner.hashrateHistory && xmrigMiner.hashrateHistory.length > 0) {
          const recentHashrates = xmrigMiner.hashrateHistory.slice(-5).map((h: any) => h.hashrate);
          console.log(`Recent hashrates: ${recentHashrates.join(', ')} H/s`);
        }
        console.log('');
      } else {
        console.log(`--- Interval ${i + 1}/10 (${elapsed}s elapsed) ---`);
        console.log('Miner data not available\n');
      }

      // Reload dashboard to see updates
      await page.reload();
      await page.waitForLoadState('networkidle');
    }

    // Final summary
    console.log('\n=== MINING SESSION SUMMARY ===');
    if (stats.length > 0) {
      const avgHashrate = stats.reduce((a, b) => a + b.hashrate, 0) / stats.length;
      const maxHashrate = Math.max(...stats.map(s => s.hashrate));
      const minHashrate = Math.min(...stats.map(s => s.hashrate));
      const totalShares = stats[stats.length - 1].shares;
      const totalRejected = stats[stats.length - 1].rejected;
      const finalUptime = stats[stats.length - 1].uptime;

      console.log(`Duration: ${finalUptime} seconds`);
      console.log(`Average Hashrate: ${avgHashrate.toFixed(2)} H/s`);
      console.log(`Max Hashrate: ${maxHashrate.toFixed(2)} H/s`);
      console.log(`Min Hashrate: ${minHashrate.toFixed(2)} H/s`);
      console.log(`Hashrate Variance: ${((maxHashrate - minHashrate) / avgHashrate * 100).toFixed(1)}%`);
      console.log(`Total Shares: ${totalShares} accepted, ${totalRejected} rejected`);
      console.log(`Share Rate: ${(totalShares / (finalUptime / 60)).toFixed(2)} shares/min`);

      // Check for anomalies
      console.log('\n=== ANOMALY CHECK ===');
      const hashrateDrops = stats.filter((s, i) => i > 0 && s.hashrate < stats[i-1].hashrate * 0.8);
      if (hashrateDrops.length > 0) {
        console.log(`WARNING: ${hashrateDrops.length} significant hashrate drops detected`);
      } else {
        console.log('No significant hashrate drops');
      }

      if (totalRejected > 0) {
        const rejectRate = (totalRejected / (totalShares + totalRejected) * 100).toFixed(2);
        console.log(`WARNING: Reject rate: ${rejectRate}%`);
      } else {
        console.log('No rejected shares');
      }

      // Memory trend
      const memoryTrend = stats[stats.length - 1].memoryMB - stats[0].memoryMB;
      if (memoryTrend > 100) {
        console.log(`WARNING: Memory increased by ${memoryTrend} MB during session`);
      } else {
        console.log(`Memory stable (change: ${memoryTrend} MB)`);
      }
    }

    // Stop mining
    console.log('\n=== Stopping miner ===');
    const minersToStop = await request.get(`${API_BASE}/miners`);
    const runningMiners = await minersToStop.json();
    for (const miner of runningMiners) {
      if (miner.name.startsWith('xmrig')) {
        await request.delete(`${API_BASE}/miners/${miner.name}`);
        console.log(`Stopped: ${miner.name}`);
      }
    }

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${profileId}`);
    console.log('Profile deleted');
    console.log('\n=== TEST COMPLETE ===');
  });
});
