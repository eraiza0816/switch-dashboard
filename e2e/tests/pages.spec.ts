import { test, expect } from '@playwright/test';

test.describe('Page navigation and rendering', () => {
  test('config page loads with form', async ({ page }) => {
    await page.goto('/config');
    await expect(page.locator('h1')).toContainText('Settings');
    await expect(page.locator('#config-form')).toBeVisible();
  });

  test('backups page loads', async ({ page }) => {
    await page.goto('/backups');
    await expect(page.locator('h1')).toContainText('Backups');
  });

  test('logs page loads', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.locator('h1')).toContainText('Logs');
  });

  test('map page loads with canvas', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('h1')).toContainText('Network Map');
    await expect(page.locator('#map-canvas')).toBeVisible();
  });

  test('api docs page loads', async ({ page }) => {
    await page.goto('/api-docs');
    await expect(page.locator('h1')).toContainText('API');
  });

  test('navigation links work between pages', async ({ page }) => {
    await page.goto('/');
    await page.locator('.nav-links a[href="/config"]').click();
    await expect(page.locator('h1')).toContainText('Settings');
    await page.locator('.nav-links a[href="/logs"]').click();
    await expect(page.locator('h1')).toContainText('Logs');
    await page.locator('.nav-links a[href="/"]').click();
    await expect(page.locator('h1')).toHaveText('E2E Test');
  });
});
