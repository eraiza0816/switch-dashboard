import { test, expect } from '@playwright/test';

test.describe('API endpoints', () => {
  test('GET /api/switches returns switch data', async ({ page }) => {
    const resp = await page.request.get('/api/switches');
    expect(resp.ok()).toBeTruthy();
    const data = await resp.json();
    expect(Array.isArray(data)).toBeTruthy();
    expect(data.length).toBeGreaterThanOrEqual(1);
    expect(data[0]).toHaveProperty('name');
    expect(data[0]).toHaveProperty('ports');
    expect(data[0].ports.length).toBeGreaterThanOrEqual(1);
  });

  test('GET /api/speeds returns speed data', async ({ page }) => {
    const resp = await page.request.get('/api/speeds');
    expect(resp.ok()).toBeTruthy();
  });

  test('GET /api/topology returns graph', async ({ page }) => {
    const resp = await page.request.get('/api/topology');
    expect(resp.ok()).toBeTruthy();
    const data = await resp.json();
    expect(data).toHaveProperty('nodes');
    expect(data).toHaveProperty('links');
  });

  test('GET /api/settings returns settings', async ({ page }) => {
    const resp = await page.request.get('/api/settings');
    expect(resp.ok()).toBeTruthy();
  });

  test('POST /api/notes saves a note', async ({ page }) => {
    const resp = await page.request.post('/api/notes', {
      data: { key: '192.168.1.1:1', note: 'test note' },
    });
    expect(resp.ok()).toBeTruthy();
  });

  test('POST /api/reset returns ok', async ({ page }) => {
    const resp = await page.request.post('/api/reset');
    expect(resp.ok()).toBeTruthy();
  });

  test('GET /api/vendors returns vendor list', async ({ page }) => {
    const resp = await page.request.get('/api/vendors');
    expect(resp.ok()).toBeTruthy();
  });

  test('GET /api/openapi.json returns spec', async ({ page }) => {
    const resp = await page.request.get('/api/openapi.json');
    expect(resp.ok()).toBeTruthy();
    const spec = await resp.json();
    expect(spec).toHaveProperty('openapi', '3.0.3');
    expect(spec).toHaveProperty('info');
  });
});
