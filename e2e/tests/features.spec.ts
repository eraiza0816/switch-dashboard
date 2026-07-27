import { test, expect } from '@playwright/test';

test.describe('Config page', () => {
  test('loads with form fields', async ({ page }) => {
    await page.goto('/config');
    await expect(page.locator('h1')).toContainText('Settings');
    await expect(page.locator('#config-form')).toBeVisible();
    await expect(page.locator('input[name="title"]')).toHaveValue('E2E Test');
    await expect(page.locator('input[name="refresh_interval"]')).toHaveValue('30');
  });

  test('shows switch configuration group', async ({ page }) => {
    await page.goto('/config');
    await expect(page.locator('.switch-group')).toHaveCount(1);
    await expect(page.locator('.switch-group h3')).toContainText('Test Switch');
    await expect(page.locator('input[name="ip[]"]')).toHaveValue('192.168.10.247');
  });

  test('language selector is present', async ({ page }) => {
    await page.goto('/config');
    await expect(page.locator('#lang-select')).toBeVisible();
    await expect(page.locator('#lang-select')).toContainText('English');
    await expect(page.locator('#lang-select')).toContainText('日本語');
  });

  test('save button is present', async ({ page }) => {
    await page.goto('/config');
    await expect(page.locator('button[type="submit"]')).toContainText('Save Configuration');
  });
});

test.describe('Backups page', () => {
  test('loads and shows empty state', async ({ page }) => {
    await page.goto('/backups');
    await expect(page.locator('h1')).toContainText('Backups');
    await expect(page.locator('#table-wrapper')).toBeVisible();
  });

  test('fetches backup list from API', async ({ page }) => {
    const resp = await page.request.get('/api/backups');
    expect(resp.ok()).toBeTruthy();
    const data = await resp.json();
    expect(Array.isArray(data)).toBeTruthy();
  });
});

test.describe('Logs page', () => {
  test('loads with log viewer elements', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.locator('h1')).toContainText('Logs');
    await expect(page.locator('#log-level')).toBeVisible();
  });

  test('shows log level selector options', async ({ page }) => {
    await page.goto('/logs');
    const select = page.locator('#log-level');
    await expect(select).toBeVisible();
    await expect(select).toContainText('INFO');
    await expect(select).toContainText('DEBUG');
    await expect(select).toContainText('ERROR');
  });

  test('download and clear buttons are present', async ({ page }) => {
    await page.goto('/logs');
    await expect(page.locator('button', { hasText: 'Download' })).toBeVisible();
    await expect(page.locator('button', { hasText: 'Clear' })).toBeVisible();
  });

  test('fetches logs from API', async ({ page }) => {
    const resp = await page.request.get('/api/logs');
    expect(resp.ok()).toBeTruthy();
  });

  test('changes log level via API', async ({ page }) => {
    const resp = await page.request.post('/api/logs/level', {
      data: { level: 'DEBUG' },
    });
    expect(resp.ok()).toBeTruthy();
    const data = await resp.json();
    expect(data).toHaveProperty('status', 'ok');
  });

  test('clears logs via API', async ({ page }) => {
    const resp = await page.request.post('/api/logs/clear');
    expect(resp.ok()).toBeTruthy();
    const data = await resp.json();
    expect(data).toHaveProperty('status', 'ok');
  });

  test('downloads logs via API', async ({ page }) => {
    const resp = await page.request.get('/api/logs/download');
    expect(resp.ok()).toBeTruthy();
  });
});

test.describe('API Docs page', () => {
  test('loads with API documentation', async ({ page }) => {
    await page.goto('/api-docs');
    await expect(page.locator('h1')).toContainText('API');
    await expect(page.locator('.endpoint').first()).toBeVisible({ timeout: 15000 });
  });

  test('shows endpoint details', async ({ page }) => {
    await page.goto('/api-docs');
    await expect(page.locator('.route').first()).toBeVisible({ timeout: 15000 });
    await expect(page.locator('.method').first()).toBeVisible();
  });
});

test.describe('Static assets', () => {
  test('serves logo.png', async ({ page }) => {
    const resp = await page.request.get('/static/logo.png');
    expect(resp.ok()).toBeTruthy();
  });

  test('serves style.css', async ({ page }) => {
    const resp = await page.request.get('/static/style.css');
    expect(resp.ok()).toBeTruthy();
    const text = await resp.text();
    expect(text).toContain('body');
  });

  test('serves frontend JS bundles', async ({ page }) => {
    for (const file of ['dashboard.js', 'logs.js', 'map.js', 'backups.js']) {
      const resp = await page.request.get(`/static/dist/${file}`);
      expect(resp.ok()).toBeTruthy();
    }
  });
});

test.describe('Navigation', () => {
  test('all nav links work between pages', async ({ page }) => {
    await page.goto('/');
    const links = [
      { href: '/map', title: 'Map' },
      { href: '/config', title: 'Settings' },
      { href: '/backups', title: 'Backups' },
      { href: '/logs', title: 'Logs' },
    ];
    for (const link of links) {
      await page.locator(`a[href="${link.href}"]`).click();
      await expect(page.locator('h1')).toContainText(link.title);
    }
  });
});
