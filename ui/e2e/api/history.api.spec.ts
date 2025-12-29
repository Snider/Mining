import { test, expect, request } from '@playwright/test';

const API_BASE_URL = process.env['API_URL'] || 'http://localhost:9090/api/v1/mining';

test.describe('History API Endpoints', () => {
  test('GET /history/status - returns database persistence status', async () => {
    const apiContext = await request.newContext();
    const response = await apiContext.get(`${API_BASE_URL}/history/status`);

    expect(response.ok()).toBeTruthy();
    const data = await response.json();

    // Database should be enabled by default
    expect(data).toHaveProperty('enabled');
    expect(typeof data.enabled).toBe('boolean');

    // Should have retention days configured
    expect(data).toHaveProperty('retentionDays');
    expect(typeof data.retentionDays).toBe('number');
    expect(data.retentionDays).toBeGreaterThan(0);
  });

  test('GET /history/miners - returns all miners historical stats', async () => {
    const apiContext = await request.newContext();
    const response = await apiContext.get(`${API_BASE_URL}/history/miners`);

    expect(response.ok()).toBeTruthy();
    const data = await response.json();

    // Should return array or null (if no data yet)
    expect(data === null || Array.isArray(data)).toBeTruthy();

    // If there is data, verify structure
    if (Array.isArray(data) && data.length > 0) {
      const stat = data[0];
      expect(stat).toHaveProperty('minerName');
      expect(stat).toHaveProperty('totalPoints');
      expect(stat).toHaveProperty('averageRate');
      expect(stat).toHaveProperty('maxRate');
      expect(stat).toHaveProperty('minRate');
    }
  });

  test('GET /history/miners/:name - returns 404 for non-existent miner', async () => {
    const apiContext = await request.newContext();
    const response = await apiContext.get(`${API_BASE_URL}/history/miners/non-existent-miner`);

    // Should return 404 for miner with no historical data
    expect(response.status()).toBe(404);

    const data = await response.json();
    expect(data).toHaveProperty('error');
  });

  test('GET /history/miners/:name/hashrate - returns historical hashrate data', async () => {
    const apiContext = await request.newContext();

    // Query with time range parameters
    const since = new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(); // 24 hours ago
    const until = new Date().toISOString();

    const response = await apiContext.get(
      `${API_BASE_URL}/history/miners/test-miner/hashrate?since=${since}&until=${until}`
    );

    expect(response.ok()).toBeTruthy();
    const data = await response.json();

    // Should return array (possibly empty)
    expect(Array.isArray(data) || data === null).toBeTruthy();

    // If there is data, verify structure
    if (Array.isArray(data) && data.length > 0) {
      const point = data[0];
      expect(point).toHaveProperty('timestamp');
      expect(point).toHaveProperty('hashrate');
      expect(typeof point.hashrate).toBe('number');
    }
  });

  test('database persistence configuration is honored', async () => {
    const apiContext = await request.newContext();

    // Get current status
    const statusResponse = await apiContext.get(`${API_BASE_URL}/history/status`);
    expect(statusResponse.ok()).toBeTruthy();
    const status = await statusResponse.json();

    // If enabled, the miners endpoint should work
    if (status.enabled) {
      const minersResponse = await apiContext.get(`${API_BASE_URL}/history/miners`);
      expect(minersResponse.ok()).toBeTruthy();
    }

    // Retention days should be reasonable (1-365)
    expect(status.retentionDays).toBeGreaterThanOrEqual(1);
    expect(status.retentionDays).toBeLessThanOrEqual(365);
  });
});
