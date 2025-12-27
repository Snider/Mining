import { test, expect } from '@playwright/test';
import { API_BASE, testProfile } from '../fixtures/test-data';
import { ProfileCreatePage } from '../page-objects/profile-create.page';
import { ProfileListPage } from '../page-objects/profile-list.page';

test.describe('Profile Management E2E', () => {
  test.beforeEach(async ({ request }) => {
    // Clean up test profiles before each test
    const profiles = await request.get(`${API_BASE}/profiles`);
    if (profiles.ok()) {
      const profileList = await profiles.json();
      for (const profile of profileList) {
        if (profile.name?.startsWith('Test') || profile.name?.startsWith('E2E')) {
          await request.delete(`${API_BASE}/profiles/${profile.id}`);
        }
      }
    }
  });

  test('displays profile create form', async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileCreate = new ProfileCreatePage(page);
    await expect(profileCreate.form).toBeVisible();
    await expect(profileCreate.nameInput).toBeVisible();
    await expect(profileCreate.createButton).toBeVisible();
  });

  test('can create a new profile via the form', async ({ page, request }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileCreate = new ProfileCreatePage(page);

    await profileCreate.fillProfile({
      name: 'E2E Test Profile',
      minerType: 'xmrig',
      pool: testProfile.config.pool,
      wallet: testProfile.config.wallet,
    });

    await profileCreate.submitForm();

    // Wait for API response
    await page.waitForResponse((resp) => resp.url().includes('/profiles') && resp.status() === 201);

    // Verify profile was created via API
    const profiles = await request.get(`${API_BASE}/profiles`);
    const profileList = await profiles.json();
    const createdProfile = profileList.find((p: { name: string }) => p.name === 'E2E Test Profile');
    expect(createdProfile).toBeDefined();
  });

  test('displays existing profiles in the list', async ({ page, request }) => {
    // Create a profile via API first
    await request.post(`${API_BASE}/profiles`, {
      data: { ...testProfile, name: 'E2E List Test Profile' },
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = new ProfileListPage(page);
    await profileList.waitForProfileListLoad();

    // Wait for profile to appear with explicit timeout
    const profileLocator = page.locator('text=E2E List Test Profile');
    await expect(profileLocator).toBeVisible({ timeout: 10000 });
  });

  test('can delete a profile', async ({ page, request }) => {
    // Create a profile via API first
    const createResponse = await request.post(`${API_BASE}/profiles`, {
      data: { ...testProfile, name: 'E2E Delete Test Profile' },
    });
    const profile = await createResponse.json();

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileListPage = new ProfileListPage(page);
    await profileListPage.waitForProfileListLoad();

    // Wait for the profile to appear before trying to delete
    const profileLocator = page.locator('text=E2E Delete Test Profile');
    await expect(profileLocator).toBeVisible({ timeout: 10000 });

    // Click delete button
    await profileListPage.clickDeleteButton('E2E Delete Test Profile');

    // Wait for deletion API call
    await page.waitForResponse(
      (resp) => resp.url().includes('/profiles/') && resp.request().method() === 'DELETE'
    );

    // Verify via API
    const getResponse = await request.get(`${API_BASE}/profiles/${profile.id}`);
    expect(getResponse.status()).toBe(404);
  });

  test('shows empty state when no profiles exist', async ({ page, request }) => {
    // Navigate to page first
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // Clean up ALL profiles
    const profiles = await request.get(`${API_BASE}/profiles`);
    if (profiles.ok()) {
      const profileList = await profiles.json();
      for (const profile of profileList) {
        await request.delete(`${API_BASE}/profiles/${profile.id}`);
      }
    }

    // Reload to get fresh state after cleanup
    await page.reload();
    await page.waitForLoadState('networkidle');

    const profileListPage = new ProfileListPage(page);
    await profileListPage.waitForProfileListLoad();

    // Check for empty state message
    await expect(profileListPage.noProfilesMessage).toBeVisible();
  });
});
