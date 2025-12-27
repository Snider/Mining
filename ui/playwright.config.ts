import { defineConfig, devices } from '@playwright/test';

const isApiOnly = process.env.API_ONLY === 'true';

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: process.env.CI ? 1 : undefined,
  reporter: [['html', { open: 'never' }], ['list']],
  use: {
    baseURL: 'http://localhost:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'api',
      testMatch: /.*\.api\.spec\.ts/,
      use: {
        baseURL: 'http://localhost:9090/api/v1/mining',
      },
    },
    {
      name: 'chromium',
      testMatch: /.*\.e2e\.spec\.ts/,
      use: { ...devices['Desktop Chrome'] },
    },
    {
      name: 'firefox',
      testMatch: /.*\.e2e\.spec\.ts/,
      use: { ...devices['Desktop Firefox'] },
    },
    {
      name: 'webkit',
      testMatch: /.*\.e2e\.spec\.ts/,
      use: { ...devices['Desktop Safari'] },
    },
  ],

  webServer: isApiOnly
    ? [
        {
          command: 'cd .. && make build && ./miner-cli serve --host localhost --port 9090',
          url: 'http://localhost:9090/api/v1/mining/info',
          reuseExistingServer: true,
          timeout: 120000,
        },
      ]
    : [
        {
          command: 'cd .. && make build && ./miner-cli serve --host localhost --port 9090',
          url: 'http://localhost:9090/api/v1/mining/info',
          reuseExistingServer: !process.env.CI,
          timeout: 120000,
        },
        {
          command: 'npm run start',
          url: 'http://localhost:4200',
          reuseExistingServer: !process.env.CI,
          timeout: 120000,
        },
      ],
});
