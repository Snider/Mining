import { Page, Locator } from '@playwright/test';

export class MinersPage {
  readonly page: Page;

  // Header
  readonly pageTitle: Locator;
  readonly pageDescription: Locator;

  // System info header
  readonly systemInfoTitle: Locator;

  constructor(page: Page) {
    this.page = page;

    // Header
    this.pageTitle = page.getByRole('heading', { name: 'Miner Software' });
    this.pageDescription = page.getByText('Install and manage mining software');

    // System info
    this.systemInfoTitle = page.getByRole('heading', { name: 'System Information' });
  }

  async isVisible(): Promise<boolean> {
    return await this.pageTitle.isVisible();
  }

  async getMinerCount(): Promise<number> {
    // Count cards by looking for miner names (xmrig, tt-miner, etc.)
    const xmrigCard = this.page.getByRole('heading', { name: 'xmrig', exact: true });
    const ttMinerCard = this.page.getByRole('heading', { name: 'tt-miner', exact: true });
    let count = 0;
    if (await xmrigCard.isVisible()) count++;
    if (await ttMinerCard.isVisible()) count++;
    return count;
  }

  async getMinerNames(): Promise<string[]> {
    const names: string[] = [];
    // Check for common miner names
    const minerNames = ['xmrig', 'tt-miner', 'lolminer', 'trex'];
    for (const name of minerNames) {
      const heading = this.page.getByRole('heading', { name: name, exact: true });
      if (await heading.isVisible()) {
        names.push(name);
      }
    }
    return names;
  }

  async isMinerInstalled(minerName: string): Promise<boolean> {
    const installedText = this.page.getByText('Installed').first();
    // Find the card section containing the miner name
    const section = this.page.locator(`text=${minerName}`).locator('..');
    return await section.getByText('Installed').isVisible().catch(() => false);
  }

  async clickInstallMiner(minerName: string) {
    // Find Install button near the miner name
    const installBtn = this.page.getByRole('button', { name: 'Install' });
    await installBtn.click();
  }

  async clickUninstallMiner(minerName: string) {
    const uninstallBtn = this.page.getByRole('button', { name: 'Uninstall' });
    await uninstallBtn.click();
  }

  async isInstallButtonVisible(minerName: string): Promise<boolean> {
    const installBtn = this.page.getByRole('button', { name: 'Install' });
    return await installBtn.isVisible();
  }

  async isUninstallButtonVisible(minerName: string): Promise<boolean> {
    const uninstallBtn = this.page.getByRole('button', { name: 'Uninstall' });
    return await uninstallBtn.isVisible();
  }

  async hasSystemInfo(): Promise<boolean> {
    return await this.systemInfoTitle.isVisible();
  }

  async getPlatform(): Promise<string> {
    const platformLabel = this.page.getByText('Platform');
    const platformSection = platformLabel.locator('..');
    // Get the next sibling or adjacent text
    const platformText = await platformSection.textContent() ?? '';
    // Extract value after "Platform"
    return platformText.replace('Platform', '').trim();
  }

  async getCPU(): Promise<string> {
    const cpuLabel = this.page.getByText('CPU');
    const cpuSection = cpuLabel.locator('..');
    const cpuText = await cpuSection.textContent() ?? '';
    return cpuText.replace('CPU', '').trim();
  }

  async getCores(): Promise<string> {
    const coresLabel = this.page.getByText('Cores');
    const coresSection = coresLabel.locator('..');
    const coresText = await coresSection.textContent() ?? '';
    return coresText.replace('Cores', '').trim();
  }

  async getMemory(): Promise<string> {
    const memoryLabel = this.page.getByText('Memory');
    const memorySection = memoryLabel.locator('..');
    const memoryText = await memorySection.textContent() ?? '';
    return memoryText.replace('Memory', '').trim();
  }
}
