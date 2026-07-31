import { test } from '@playwright/test';

const OUT = process.env.OUT_DIR || '../images';

test('dashboard full page', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto('/');
  await page.waitForSelector('.switch-card', { timeout: 15000 });
  await page.waitForTimeout(500);
  await page.screenshot({ path: `${OUT}/dashboard.png`, fullPage: true });
});

test('MAC table', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto('/');
  await page.waitForSelector('.switch-card', { timeout: 15000 });
  await page.locator('.mac-header').first().click();
  await page.waitForSelector('.mac-table', { timeout: 5000 });
  await page.waitForTimeout(300);
  await page.locator('.mac-table').first().screenshot({ path: `${OUT}/mac_table.png` });
});

test('network map', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto('/map');
  await page.waitForSelector('#map-canvas svg', { timeout: 15000 });
  await page.waitForTimeout(500);
  await page.locator('#map-canvas').screenshot({ path: `${OUT}/network-map.png` });
});

test('speed graph modal', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 });
  await page.goto('/');
  await page.waitForSelector('.switch-card', { timeout: 15000 });
  await page.locator('.status-badge').first().click();
  await page.waitForSelector('#graph-overlay', { timeout: 5000 });
  await page.waitForTimeout(300);
  await page.locator('#graph-overlay').screenshot({ path: `${OUT}/speed_graph.png` });
});
