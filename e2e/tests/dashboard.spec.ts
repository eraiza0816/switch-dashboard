import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test('loads and shows navigation and switch cards', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('h1')).toHaveText('E2E Test');
    await expect(page.locator('.nav-links')).toBeVisible();
    await expect(page.locator('.switch-card')).toHaveCount(1);
    await expect(page.locator('.switch-header h2')).toContainText('Test Switch');
  });

  test('port table renders with correct data', async ({ page }) => {
    await page.goto('/');
    const rows = page.locator('.port-table tbody tr');
    await expect(rows).toHaveCount(6);
    await expect(rows.nth(0).locator('td').first()).toContainText('1');
    await expect(rows.nth(4).locator('td').first()).toContainText('5');
    await expect(rows.nth(0)).toContainText('10G');
  });

  test('reset link triggers confirm dialog', async ({ page }) => {
    await page.goto('/');
    page.on('dialog', (dialog) => {
      expect(dialog.message()).toContain('Reset');
      dialog.dismiss();
    });
    await page.locator('#reset-link').click();
  });

  test('font size buttons update body attribute', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('body')).toHaveAttribute('data-font-size', 'md');
    await page.locator('.fs-btn[data-size="lg"]').click();
    await page.waitForTimeout(100);
    await expect(page.locator('body')).toHaveAttribute('data-font-size', 'lg');
    await page.locator('.fs-btn[data-size="sm"]').click();
    await page.waitForTimeout(100);
    await expect(page.locator('body')).toHaveAttribute('data-font-size', 'sm');
  });

  test('switch card shows device info bar', async ({ page }) => {
    await page.goto('/');
    const card = page.locator('.switch-card').first();
    await expect(card.locator('.device-bar')).toBeVisible();
    await expect(card.locator('.device-bar')).toContainText('MAC:');
    await expect(card.locator('.device-bar')).toContainText('v0.2.19');
  });

  test('MAC table section is expandable', async ({ page }) => {
    await page.goto('/');
    const macHeader = page.locator('.mac-header').first();
    await expect(macHeader).toBeVisible();
    await macHeader.click();
    const macTable = page.locator('.mac-table').first();
    await expect(macTable).toBeVisible();
  });

  test('bandwidth graph modal opens on port click', async ({ page }) => {
    await page.goto('/');
    const statusBadge = page.locator('.status-badge').first();
    await statusBadge.click();
    await expect(page.locator('#graph-overlay')).toBeVisible();
    await expect(page.locator('#graph-title')).toContainText('Port ');
  });

  test('last update timestamp is shown', async ({ page }) => {
    await page.goto('/');
    await expect(page.locator('.last-update')).toContainText('Last update:');
  });

  test.describe('Extra status sections', () => {
    const sections: { key: string; label: string }[] = [
      { key: 'eee', label: 'EEE' },
      { key: 'vlan', label: 'VLAN' },
      { key: 'lag', label: 'LAG' },
      { key: 'mirror', label: 'Mirroring' },
      { key: 'bandwidth', label: 'Bandwidth' },
    ];

    for (const { key, label } of sections) {
      test(`${key} section header is visible`, async ({ page }) => {
        await page.goto('/');
        const section = page.locator(`.${key}-section`).first();
        await expect(section).toBeVisible();
        await expect(section.locator('h4')).toContainText(label);
      });

      test(`${key} section is expandable on click`, async ({ page }) => {
        await page.goto('/');
        const section = page.locator(`.${key}-section`).first();
        const header = section.locator('h4');
        await header.click();
        const content = section.locator('div[class*="-content-"]');
        await expect(content).toBeAttached();
        await expect(content).toBeVisible();
      });
    }
  });
});
