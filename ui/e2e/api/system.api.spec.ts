import { test, expect } from '@playwright/test';
import { API_BASE } from '../fixtures/test-data';

test.describe('System API Endpoints', () => {
  test('GET /info - returns system information', async ({ request }) => {
    const response = await request.get(`${API_BASE}/info`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(body).toHaveProperty('os');
    expect(body).toHaveProperty('architecture');
    expect(body).toHaveProperty('go_version');
    expect(body).toHaveProperty('available_cpu_cores');
    expect(body).toHaveProperty('total_system_ram_gb');
    expect(body).toHaveProperty('installed_miners_info');
    expect(Array.isArray(body.installed_miners_info)).toBeTruthy();
  });

  test('POST /doctor - performs live miner check', async ({ request }) => {
    const response = await request.post(`${API_BASE}/doctor`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(body).toHaveProperty('installed_miners_info');
    expect(Array.isArray(body.installed_miners_info)).toBeTruthy();
  });

  test('POST /update - checks for miner updates', async ({ request }) => {
    const response = await request.post(`${API_BASE}/update`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    // Either "status" (all up to date) or "updates_available"
    expect(body.status || body.updates_available).toBeDefined();
  });
});
