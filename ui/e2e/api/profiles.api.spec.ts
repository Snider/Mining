import { test, expect } from '@playwright/test';
import { API_BASE, testProfile } from '../fixtures/test-data';

test.describe('Profiles API CRUD', () => {
  let createdProfileId: string;

  test.beforeAll(async ({ request }) => {
    // Clean up any existing test profiles
    const profiles = await request.get(`${API_BASE}/profiles`);
    if (profiles.ok()) {
      const profileList = await profiles.json();
      for (const profile of profileList) {
        if (profile.name?.startsWith('Test')) {
          await request.delete(`${API_BASE}/profiles/${profile.id}`);
        }
      }
    }
  });

  test.afterAll(async ({ request }) => {
    // Clean up created profile if exists
    if (createdProfileId) {
      await request.delete(`${API_BASE}/profiles/${createdProfileId}`);
    }
  });

  test('GET /profiles - returns list of profiles', async ({ request }) => {
    const response = await request.get(`${API_BASE}/profiles`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(Array.isArray(body)).toBeTruthy();
  });

  test('POST /profiles - creates a new profile', async ({ request }) => {
    const response = await request.post(`${API_BASE}/profiles`, {
      data: testProfile,
    });

    expect(response.status()).toBe(201);
    const body = await response.json();

    expect(body).toHaveProperty('id');
    expect(body.name).toBe(testProfile.name);
    expect(body.minerType).toBe(testProfile.minerType);

    createdProfileId = body.id;
  });

  test('GET /profiles/:id - retrieves created profile', async ({ request }) => {
    // Skip if no profile was created
    test.skip(!createdProfileId, 'No profile created in previous test');

    const response = await request.get(`${API_BASE}/profiles/${createdProfileId}`);

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(body.id).toBe(createdProfileId);
    expect(body.name).toBe(testProfile.name);
  });

  test('PUT /profiles/:id - updates a profile', async ({ request }) => {
    test.skip(!createdProfileId, 'No profile created in previous test');

    const updatedProfile = {
      ...testProfile,
      name: 'Test Profile Updated',
    };

    const response = await request.put(`${API_BASE}/profiles/${createdProfileId}`, {
      data: updatedProfile,
    });

    expect(response.ok()).toBeTruthy();
    const body = await response.json();

    expect(body.name).toBe('Test Profile Updated');
  });

  test('GET /profiles/:id - returns 404 for non-existent profile', async ({ request }) => {
    const response = await request.get(`${API_BASE}/profiles/non-existent-id`);

    expect(response.status()).toBe(404);
  });

  test('DELETE /profiles/:id - deletes a profile', async ({ request }) => {
    // Create a profile specifically for deletion test
    const createResponse = await request.post(`${API_BASE}/profiles`, {
      data: { ...testProfile, name: 'Test Profile To Delete' },
    });
    const profile = await createResponse.json();

    const response = await request.delete(`${API_BASE}/profiles/${profile.id}`);

    expect(response.ok()).toBeTruthy();

    // Verify deletion
    const getResponse = await request.get(`${API_BASE}/profiles/${profile.id}`);
    expect(getResponse.status()).toBe(404);
  });

  test('POST /profiles/:id/start - handles starting with profile', async ({ request }) => {
    // Create a profile for this test
    const createResponse = await request.post(`${API_BASE}/profiles`, {
      data: { ...testProfile, name: 'Test Profile For Start' },
    });
    const profile = await createResponse.json();

    try {
      // Try to start - may fail if XMRig is not installed
      const startResponse = await request.post(`${API_BASE}/profiles/${profile.id}/start`);

      // Either succeeds (200) or fails gracefully (500 with error)
      expect([200, 500]).toContain(startResponse.status());

      if (startResponse.status() === 200) {
        // If started, stop the miner
        const miners = await request.get(`${API_BASE}/miners`);
        const minerList = await miners.json();
        for (const miner of minerList) {
          if (miner.name?.includes('xmrig')) {
            await request.delete(`${API_BASE}/miners/${miner.name}`);
          }
        }
      }
    } finally {
      // Cleanup profile
      await request.delete(`${API_BASE}/profiles/${profile.id}`);
    }
  });
});
