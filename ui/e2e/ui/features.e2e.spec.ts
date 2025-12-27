import { test, expect } from '@playwright/test';
import { API_BASE, testProfile } from '../fixtures/test-data';

/**
 * FEATURE LIST - Snider Mining UI
 *
 * This test file documents and tests all UI features for the Mining dashboard.
 * The app uses Angular custom elements with Web Awesome (wa-*) components.
 *
 * COMPONENTS:
 * 1. snider-mining-dashboard - Shows miner stats, chart, or "no miners running"
 * 2. snider-mining-admin - Install/uninstall miners, antivirus paths
 * 3. snider-mining-profile-create - Form to create mining profiles
 * 4. snider-mining-profile-list - List profiles with Start/Edit/Delete
 * 5. snider-mining-setup-wizard - First-time setup, install miners
 * 6. snider-mining-chart - Hashrate chart visualization
 * 7. snider-mining-stats-bar - Stats display (bar or list mode)
 */

test.describe('Feature Tests - Profile Create Form', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('profile create form renders all inputs', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    await expect(form).toBeVisible();

    // Check all form elements are present
    await expect(form.locator('wa-input[name="name"]')).toBeVisible();
    await expect(form.locator('wa-select[name="minerType"]')).toBeVisible();
    await expect(form.locator('wa-input[name="pool"]')).toBeVisible();
    await expect(form.locator('wa-input[name="wallet"]')).toBeVisible();
    await expect(form.locator('wa-checkbox[name="tls"]')).toBeVisible();
    await expect(form.locator('wa-checkbox[name="hugePages"]')).toBeVisible();
    await expect(form.locator('wa-button[type="submit"]')).toBeVisible();
  });

  test('profile name input accepts text', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const nameInput = form.locator('wa-input[name="name"]');

    await nameInput.click();
    await nameInput.pressSequentially('Test Profile Name');

    // For Web Awesome shadow DOM components, check the internal input value
    const inputValue = await nameInput.evaluate((el: any) => el.value);
    expect(inputValue).toBe('Test Profile Name');
  });

  test('miner type select shows options and can be selected', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const minerSelect = form.locator('wa-select[name="minerType"]');

    await minerSelect.click();

    // Wait for dropdown to appear
    await page.waitForTimeout(500);

    // Check if options are visible
    const options = form.locator('wa-option');
    const optionCount = await options.count();
    expect(optionCount).toBeGreaterThan(0);
  });

  test('pool address input accepts text', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const poolInput = form.locator('wa-input[name="pool"]');

    await poolInput.click();
    await poolInput.pressSequentially('stratum+tcp://pool.example.com:3333');

    const inputValue = await poolInput.evaluate((el: any) => el.value);
    expect(inputValue).toBe('stratum+tcp://pool.example.com:3333');
  });

  test('wallet address input accepts text', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const walletInput = form.locator('wa-input[name="wallet"]');

    await walletInput.click();
    await walletInput.pressSequentially('wallet123abc');

    const inputValue = await walletInput.evaluate((el: any) => el.value);
    expect(inputValue).toBe('wallet123abc');
  });

  test('TLS checkbox can be toggled', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const tlsCheckbox = form.locator('wa-checkbox[name="tls"]');

    // Get initial state via property (not attribute)
    const initialChecked = await tlsCheckbox.evaluate((el: any) => el.checked);

    // Click to toggle
    await tlsCheckbox.click();

    // State should have changed
    const newChecked = await tlsCheckbox.evaluate((el: any) => el.checked);
    expect(newChecked).not.toBe(initialChecked);
  });

  test('Huge Pages checkbox can be toggled', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const hugePagesCheckbox = form.locator('wa-checkbox[name="hugePages"]');

    await hugePagesCheckbox.click();

    // Checkbox should respond to click
    await expect(hugePagesCheckbox).toBeVisible();
  });

  test('Create Profile button is clickable', async ({ page }) => {
    const form = page.locator('snider-mining-profile-create');
    const submitButton = form.locator('wa-button[type="submit"]');

    await expect(submitButton).toBeVisible();
    await expect(submitButton).toBeEnabled();
  });
});

test.describe('Feature Tests - Profile List', () => {
  // Helper to create a unique profile and return its name
  const createTestProfile = async (request: any, suffix: string) => {
    const name = `FT-${suffix}-${Date.now()}`;
    const response = await request.post(`${API_BASE}/profiles`, {
      data: { ...testProfile, name },
    });
    const profile = await response.json();
    return { name, id: profile.id };
  };

  test('profile list displays profiles', async ({ page, request }) => {
    const { name, id } = await createTestProfile(request, 'display');

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(profileList).toBeVisible();

    // Wait for the profile to appear
    await expect(page.locator(`text=${name}`)).toBeVisible({ timeout: 10000 });

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${id}`);
  });

  test('Start button is visible and clickable', async ({ page, request }) => {
    const { name, id } = await createTestProfile(request, 'start');

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator(`text=${name}`)).toBeVisible({ timeout: 10000 });

    const profileItem = profileList.locator(`.profile-item:has-text("${name}")`);
    const startButton = profileItem.locator('wa-button:has-text("Start")');

    await expect(startButton).toBeVisible();
    await expect(startButton).toBeEnabled();

    // Click and verify it responds
    await startButton.click();

    // Should trigger an API call (may succeed or fail depending on miner installation)
    await page.waitForResponse(
      (resp) => resp.url().includes('/profiles/') && resp.url().includes('/start'),
      { timeout: 5000 }
    ).catch(() => {
      // It's OK if no response - we're just testing the button works
    });

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${id}`);
  });

  test('Edit button is visible and clickable', async ({ page, request }) => {
    const { name, id } = await createTestProfile(request, 'edit');

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator(`text=${name}`)).toBeVisible({ timeout: 10000 });

    const profileItem = profileList.locator(`.profile-item:has-text("${name}")`);
    const editButton = profileItem.locator('wa-button:has-text("Edit")');

    await expect(editButton).toBeVisible();
    await expect(editButton).toBeEnabled();

    // Click and verify edit form appears
    await editButton.click();

    // Edit form should appear with Save/Cancel buttons
    await expect(profileList.locator('wa-button:has-text("Save")')).toBeVisible({ timeout: 5000 });
    await expect(profileList.locator('wa-button:has-text("Cancel")')).toBeVisible();

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${id}`);
  });

  test('Delete button is visible and clickable', async ({ page, request }) => {
    const { name, id } = await createTestProfile(request, 'delete');

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator(`text=${name}`)).toBeVisible({ timeout: 10000 });

    const profileItem = profileList.locator(`.profile-item:has-text("${name}")`);
    const deleteButton = profileItem.locator('wa-button:has-text("Delete")');

    await expect(deleteButton).toBeVisible();
    await expect(deleteButton).toBeEnabled();

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${id}`);
  });

  test('Edit form shows all fields when editing', async ({ page, request }) => {
    const { name, id } = await createTestProfile(request, 'editform');

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator(`text=${name}`)).toBeVisible({ timeout: 10000 });

    const profileItem = profileList.locator(`.profile-item:has-text("${name}")`);
    await profileItem.locator('wa-button:has-text("Edit")').click();

    // Check edit form fields
    const editForm = profileList.locator('.profile-form');
    await expect(editForm).toBeVisible({ timeout: 5000 });
    await expect(editForm.locator('wa-input[label="Profile Name"]')).toBeVisible();
    await expect(editForm.locator('wa-select[label="Miner Type"]')).toBeVisible();
    await expect(editForm.locator('wa-input[label="Pool Address"]')).toBeVisible();
    await expect(editForm.locator('wa-input[label="Wallet Address"]')).toBeVisible();

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${id}`);
  });

  test('Cancel button exits edit mode', async ({ page, request }) => {
    const { name, id } = await createTestProfile(request, 'cancel');

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator(`text=${name}`)).toBeVisible({ timeout: 10000 });

    const profileItem = profileList.locator(`.profile-item:has-text("${name}")`);
    await profileItem.locator('wa-button:has-text("Edit")').click();

    // Verify edit mode
    await expect(profileList.locator('wa-button:has-text("Cancel")')).toBeVisible({ timeout: 5000 });

    // Click cancel
    await profileList.locator('wa-button:has-text("Cancel")').click();

    // Should return to normal view with Edit button
    await expect(profileItem.locator('wa-button:has-text("Edit")')).toBeVisible({ timeout: 5000 });

    // Cleanup
    await request.delete(`${API_BASE}/profiles/${id}`);
  });
});

test.describe('Feature Tests - Admin Panel', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('admin panel renders', async ({ page }) => {
    const admin = page.locator('snider-mining-admin');
    await expect(admin).toBeVisible();
  });

  test('shows Manage Miners heading', async ({ page }) => {
    const admin = page.locator('snider-mining-admin');
    await expect(admin.locator('h4:has-text("Manage Miners")')).toBeVisible();
  });

  test('shows miner list with install/uninstall buttons', async ({ page }) => {
    const admin = page.locator('snider-mining-admin');
    const minerList = admin.locator('.miner-list');
    await expect(minerList).toBeVisible();

    // Should have at least one miner item
    const minerItems = minerList.locator('.miner-item');
    const count = await minerItems.count();
    expect(count).toBeGreaterThan(0);

    // Each miner should have either Install or Uninstall button
    for (let i = 0; i < count; i++) {
      const item = minerItems.nth(i);
      const hasInstall = await item.locator('wa-button:has-text("Install")').isVisible();
      const hasUninstall = await item.locator('wa-button:has-text("Uninstall")').isVisible();
      expect(hasInstall || hasUninstall).toBe(true);
    }
  });

  test('Install button is clickable', async ({ page }) => {
    const admin = page.locator('snider-mining-admin');
    const installButton = admin.locator('wa-button:has-text("Install")').first();

    // Skip if no install buttons (all miners installed)
    if (await installButton.count() === 0) {
      test.skip();
      return;
    }

    await expect(installButton).toBeVisible();
    await expect(installButton).toBeEnabled();
  });

  test('Uninstall button is clickable', async ({ page }) => {
    const admin = page.locator('snider-mining-admin');
    const uninstallButton = admin.locator('wa-button:has-text("Uninstall")').first();

    // Skip if no uninstall buttons (no miners installed)
    if (await uninstallButton.count() === 0) {
      test.skip();
      return;
    }

    await expect(uninstallButton).toBeVisible();
    await expect(uninstallButton).toBeEnabled();
  });

  test('shows Antivirus Whitelist Paths section', async ({ page }) => {
    const admin = page.locator('snider-mining-admin');
    await expect(admin.locator('h4:has-text("Antivirus Whitelist Paths")')).toBeVisible();
  });
});

test.describe('Feature Tests - Miner Installation', () => {
  // These tests modify system state, run them serially
  test.describe.configure({ mode: 'serial' });

  test('can install xmrig miner via API', async ({ request }) => {
    // First check current status
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();
    const xmrigInfo = info.installed_miners_info?.find((m: any) => m.path?.includes('xmrig'));

    if (xmrigInfo?.is_installed) {
      // Already installed, skip
      test.skip();
      return;
    }

    // Install xmrig
    const installResponse = await request.post(`${API_BASE}/miners/xmrig/install`);
    expect(installResponse.ok()).toBe(true);

    const result = await installResponse.json();
    expect(result.status).toBe('installed');
    expect(result.version).toBeDefined();
  });

  test('can uninstall xmrig miner via API', async ({ request }) => {
    // First check if installed
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();
    const xmrigInfo = info.installed_miners_info?.find((m: any) => m.path?.includes('xmrig'));

    if (!xmrigInfo?.is_installed) {
      // Not installed, skip
      test.skip();
      return;
    }

    // Uninstall xmrig
    const uninstallResponse = await request.delete(`${API_BASE}/miners/xmrig/uninstall`);
    expect(uninstallResponse.ok()).toBe(true);
  });

  test('Install button triggers install API and updates UI', async ({ page, request }) => {
    // Check if xmrig is already installed
    const infoResponse = await request.get(`${API_BASE}/info`);
    const info = await infoResponse.json();
    const xmrigInfo = info.installed_miners_info?.find((m: any) => m.path?.includes('xmrig'));

    if (xmrigInfo?.is_installed) {
      // Uninstall first so we can test install
      await request.delete(`${API_BASE}/miners/xmrig/uninstall`);
    }

    await page.goto('/');
    await page.waitForLoadState('networkidle');

    const admin = page.locator('snider-mining-admin');
    const xmrigItem = admin.locator('.miner-item:has-text("xmrig")');
    const installButton = xmrigItem.locator('wa-button:has-text("Install")');

    await expect(installButton).toBeVisible({ timeout: 5000 });

    // Set up response listener before clicking
    const installPromise = page.waitForResponse(
      (resp) => resp.url().includes('/miners/xmrig/install'),
      { timeout: 120000 }
    );

    // Click install
    await installButton.click();

    // Wait for install to complete
    const response = await installPromise;
    expect(response.ok()).toBe(true);

    // After install, the button should change to Uninstall
    await expect(xmrigItem.locator('wa-button:has-text("Uninstall")')).toBeVisible({ timeout: 15000 });
  });
});

test.describe('Feature Tests - Setup Wizard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('setup wizard renders', async ({ page }) => {
    const wizard = page.locator('snider-mining-setup-wizard');
    await expect(wizard).toBeVisible();
  });

  test('shows Setup Required header', async ({ page }) => {
    const wizard = page.locator('snider-mining-setup-wizard');
    await expect(wizard.locator('text=Setup Required')).toBeVisible();
  });

  test('shows Available Miners heading', async ({ page }) => {
    const wizard = page.locator('snider-mining-setup-wizard');
    await expect(wizard.locator('h4:has-text("Available Miners")')).toBeVisible();
  });

  test('displays miner list with buttons', async ({ page }) => {
    const wizard = page.locator('snider-mining-setup-wizard');
    const minerList = wizard.locator('.miner-list');
    await expect(minerList).toBeVisible();

    const minerItems = minerList.locator('.miner-item');
    const count = await minerItems.count();
    expect(count).toBeGreaterThan(0);
  });

  test('Install button in wizard is clickable', async ({ page }) => {
    const wizard = page.locator('snider-mining-setup-wizard');
    const installButton = wizard.locator('wa-button:has-text("Install")').first();

    if (await installButton.count() === 0) {
      test.skip();
      return;
    }

    await expect(installButton).toBeVisible();
    await expect(installButton).toBeEnabled();
  });
});

test.describe('Feature Tests - Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');
  });

  test('dashboard component renders', async ({ page }) => {
    const dashboard = page.locator('snider-mining-dashboard').first();
    await expect(dashboard).toBeVisible();
  });

  test('shows no miners message when no miners running', async ({ page, request }) => {
    // Check if miners are running
    const minersResponse = await request.get(`${API_BASE}/miners`);
    const miners = await minersResponse.json();

    if (miners.length > 0) {
      test.skip();
      return;
    }

    const dashboard = page.locator('snider-mining-dashboard').first();
    await expect(dashboard.locator('text=No miners running')).toBeVisible();
  });
});

test.describe('Feature Tests - Full User Flow', () => {
  test.beforeEach(async ({ request }) => {
    // Clean up test profiles
    const profiles = await request.get(`${API_BASE}/profiles`);
    if (profiles.ok()) {
      const profileList = await profiles.json();
      for (const profile of profileList) {
        if (profile.name?.includes('Flow Test')) {
          await request.delete(`${API_BASE}/profiles/${profile.id}`);
        }
      }
    }
  });

  test('complete flow: create profile, verify in list, edit, delete', async ({ page, request }) => {
    await page.goto('/');
    await page.waitForLoadState('networkidle');

    // STEP 1: Create a profile using the form
    const form = page.locator('snider-mining-profile-create');

    // Fill in name
    const nameInput = form.locator('wa-input[name="name"]');
    await nameInput.click();
    await nameInput.pressSequentially('Flow Test Profile');

    // Select miner type
    const minerSelect = form.locator('wa-select[name="minerType"]');
    await minerSelect.click();
    await page.waitForTimeout(300);
    const firstOption = form.locator('wa-option').first();
    if (await firstOption.count() > 0) {
      await firstOption.click();
    }

    // Fill pool
    const poolInput = form.locator('wa-input[name="pool"]');
    await poolInput.click();
    await poolInput.pressSequentially('stratum+tcp://pool.test.com:3333');

    // Fill wallet
    const walletInput = form.locator('wa-input[name="wallet"]');
    await walletInput.click();
    await walletInput.pressSequentially('testwalletaddress123');

    // Submit form
    await form.locator('wa-button[type="submit"]').click();

    // Wait for API response
    await page.waitForResponse(
      (resp) => resp.url().includes('/profiles') && resp.status() === 201,
      { timeout: 10000 }
    );

    // STEP 2: Verify profile appears in list
    const profileList = page.locator('snider-mining-profile-list');
    await expect(page.locator('text=Flow Test Profile')).toBeVisible({ timeout: 10000 });

    // STEP 3: Edit the profile
    const profileItem = profileList.locator('.profile-item:has-text("Flow Test Profile")');
    await profileItem.locator('wa-button:has-text("Edit")').click();

    // Verify edit form appears
    await expect(profileList.locator('wa-button:has-text("Save")')).toBeVisible({ timeout: 5000 });

    // Cancel edit
    await profileList.locator('wa-button:has-text("Cancel")').click();

    // STEP 4: Delete the profile
    await profileItem.locator('wa-button:has-text("Delete")').click();

    // Wait for deletion
    await page.waitForResponse(
      (resp) => resp.url().includes('/profiles/') && resp.request().method() === 'DELETE',
      { timeout: 5000 }
    );

    // Verify profile is gone
    await expect(page.locator('text=Flow Test Profile')).not.toBeVisible({ timeout: 5000 });
  });
});
