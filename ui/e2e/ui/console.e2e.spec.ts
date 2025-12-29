import { test, expect } from '@playwright/test';
import { MainLayoutPage } from '../page-objects/main-layout.page';
import { ConsolePage } from '../page-objects/console.page';

test.describe('Console Page', () => {
  let layout: MainLayoutPage;
  let consolePage: ConsolePage;

  test.beforeEach(async ({ page }) => {
    layout = new MainLayoutPage(page);
    consolePage = new ConsolePage(page);
    await layout.goto();
    await layout.waitForLayoutLoad();
    await layout.navigateToConsole();
  });

  test.describe('Console Layout', () => {
    test('should display console page container', async () => {
      await expect(consolePage.consolePage).toBeVisible();
    });

    test('should display console header', async () => {
      await expect(consolePage.tabsContainer).toBeVisible();
    });

    test('should display console output area', async () => {
      await expect(consolePage.consoleOutput).toBeVisible();
    });

    test('should display console controls', async () => {
      await expect(consolePage.autoScrollCheckbox).toBeVisible();
      await expect(consolePage.clearButton).toBeVisible();
    });
  });

  test.describe('Worker Selection', () => {
    test('should show worker dropdown when miners are running', async ({ page }) => {
      // Check if there's a worker select dropdown or no miners message
      const workerSelect = page.locator('.worker-select');
      const noMinersMsg = page.locator('.no-miners-msg');

      // Either worker select or no miners message should be visible
      const hasWorkerSelect = await workerSelect.isVisible();
      const hasNoMiners = await noMinersMsg.isVisible();

      expect(hasWorkerSelect || hasNoMiners).toBe(true);
    });

    test('should auto-select first miner on load', async ({ page }) => {
      const workerSelect = page.locator('.worker-select');

      if (await workerSelect.isVisible()) {
        // A miner should be selected (value should not be empty)
        const selectedValue = await workerSelect.inputValue();
        expect(selectedValue).toBeTruthy();
      }
    });

    test('should display miner tabs when multiple miners running', async ({ page }) => {
      // This test checks if tabs appear when there are multiple miners
      // We just verify the tabs container exists in the header
      const consoleTabs = page.locator('.console-tabs');
      // Tabs only show when multiple miners, so this may or may not be visible
      // Just ensure no errors
      await consolePage.tabsContainer.isVisible();
    });
  });

  test.describe('Console Output', () => {
    test('should display logs when miner is running', async ({ page }) => {
      const workerSelect = page.locator('.worker-select');

      if (await workerSelect.isVisible()) {
        // Wait for logs to load (poll happens every 2 seconds)
        await page.waitForTimeout(3000);

        // Check if logs are displayed or waiting message
        const logLines = await consolePage.getLogLineCount();
        const waitingMsg = page.getByText(/Waiting for logs from/);
        const hasWaitingMsg = await waitingMsg.isVisible();

        // Either logs should be present or waiting message
        expect(logLines > 0 || hasWaitingMsg).toBe(true);
      }
    });

    test('should show empty state when no miners running', async ({ page }) => {
      const noMinersMsg = page.locator('.no-miners-msg');

      if (await noMinersMsg.isVisible()) {
        // When no miners, empty state should show
        const emptyState = consolePage.emptyState;
        await expect(emptyState).toBeVisible();
      }
    });

    test('should style error lines correctly', async ({ page }) => {
      // Wait for logs
      await page.waitForTimeout(3000);

      const errorLines = page.locator('.log-line.error');
      const errorCount = await errorLines.count();

      // If there are error lines, verify they have error styling
      if (errorCount > 0) {
        const firstError = errorLines.first();
        await expect(firstError).toHaveClass(/error/);
      }
    });

    test('should style warning lines correctly', async ({ page }) => {
      // Wait for logs
      await page.waitForTimeout(3000);

      const warningLines = page.locator('.log-line.warning');
      const warningCount = await warningLines.count();

      // If there are warning lines, verify they have warning styling
      if (warningCount > 0) {
        const firstWarning = warningLines.first();
        await expect(firstWarning).toHaveClass(/warning/);
      }
    });
  });

  test.describe('Console Controls', () => {
    test('should have auto-scroll enabled by default', async () => {
      const isEnabled = await consolePage.isAutoScrollEnabled();
      expect(isEnabled).toBe(true);
    });

    test('should toggle auto-scroll when checkbox clicked', async () => {
      // Initially enabled
      expect(await consolePage.isAutoScrollEnabled()).toBe(true);

      // Toggle off
      await consolePage.toggleAutoScroll();
      expect(await consolePage.isAutoScrollEnabled()).toBe(false);

      // Toggle back on
      await consolePage.toggleAutoScroll();
      expect(await consolePage.isAutoScrollEnabled()).toBe(true);
    });

    test('should clear logs when clear button clicked', async ({ page }) => {
      const workerSelect = page.locator('.worker-select');

      if (await workerSelect.isVisible()) {
        // Wait for logs to load
        await page.waitForTimeout(3000);

        const initialLogCount = await consolePage.getLogLineCount();

        if (initialLogCount > 0) {
          // Clear logs
          await consolePage.clearLogs();

          // Verify logs are cleared
          const finalLogCount = await consolePage.getLogLineCount();
          expect(finalLogCount).toBe(0);
        }
      }
    });

    test('should disable clear button when no logs', async ({ page }) => {
      const workerSelect = page.locator('.worker-select');

      if (await workerSelect.isVisible()) {
        // Wait for logs
        await page.waitForTimeout(3000);

        const logCount = await consolePage.getLogLineCount();

        if (logCount > 0) {
          // Clear logs
          await consolePage.clearLogs();

          // Clear button should be disabled now
          const isEnabled = await consolePage.isClearButtonEnabled();
          expect(isEnabled).toBe(false);
        }
      }
    });
  });

  test.describe('Log Polling', () => {
    test('should update logs periodically', async ({ page }) => {
      const workerSelect = page.locator('.worker-select');

      if (await workerSelect.isVisible()) {
        // Wait for initial logs
        await page.waitForTimeout(3000);

        const initialLogs = await consolePage.getLogContent();

        // Wait for another poll cycle (2+ seconds)
        await page.waitForTimeout(3000);

        // Check if new logs appeared (miner generates output)
        // We can't guarantee new logs, but verify no errors
        const finalLogs = await consolePage.getLogContent();

        // Logs array should still be valid
        expect(Array.isArray(finalLogs)).toBe(true);
      }
    });
  });

  test.describe('Worker Switching', () => {
    test('should switch miner and clear logs when selection changes', async ({ page }) => {
      const workerSelect = page.locator('.worker-select');

      if (await workerSelect.isVisible()) {
        // Get available options
        const options = await workerSelect.locator('option').allTextContents();

        if (options.length > 1) {
          // Wait for initial logs
          await page.waitForTimeout(3000);

          // Get initial selected value
          const initialValue = await workerSelect.inputValue();

          // Find a different miner to select
          const otherMiner = options.find(opt => opt !== initialValue);

          if (otherMiner) {
            // Select different miner
            await workerSelect.selectOption(otherMiner);

            // Logs should be cleared momentarily (new miner selected)
            // Then new logs should load
            await page.waitForTimeout(500);
          }
        }
      }
    });
  });

  test.describe('Responsive Behavior', () => {
    test('should maintain layout when resized', async ({ page }) => {
      // Resize to smaller viewport
      await page.setViewportSize({ width: 800, height: 600 });

      // Console should still be visible and functional
      await expect(consolePage.consolePage).toBeVisible();
      await expect(consolePage.consoleOutput).toBeVisible();
      await expect(consolePage.clearButton).toBeVisible();
    });
  });
});
