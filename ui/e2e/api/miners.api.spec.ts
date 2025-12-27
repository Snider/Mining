import { test, expect } from '@playwright/test';
import { API_BASE } from '../fixtures/test-data';

test.describe('Miners API Endpoints', () => {
  test('GET /miners - returns list of running miners', async ({ request }) => {
    const response = await request.get(`${API_BASE}/miners`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(Array.isArray(body)).toBeTruthy();
  });

  test('GET /miners/available - returns available miner types', async ({ request }) => {
    const response = await request.get(`${API_BASE}/miners/available`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(Array.isArray(body)).toBeTruthy();
    expect(body.length).toBeGreaterThan(0);

    // Check xmrig is in the list
    const xmrig = body.find((m: { name: string }) => m.name === 'xmrig');
    expect(xmrig).toBeDefined();
    expect(xmrig).toHaveProperty('description');
  });

  test.describe('error handling', () => {
    test('GET /miners/:name/stats - returns 404 for non-existent miner', async ({ request }) => {
      const response = await request.get(`${API_BASE}/miners/nonexistent/stats`);

      expect(response.status()).toBe(404);
    });

    test('DELETE /miners/:name - handles stopping non-running miner', async ({ request }) => {
      const response = await request.delete(`${API_BASE}/miners/nonexistent`);

      // Should return error since miner isn't running
      expect(response.status()).toBeGreaterThanOrEqual(400);
    });

    test('GET /miners/:name/hashrate-history - returns 404 for non-existent miner', async ({
      request,
    }) => {
      const response = await request.get(`${API_BASE}/miners/nonexistent/hashrate-history`);

      expect(response.status()).toBe(404);
    });
  });
});
