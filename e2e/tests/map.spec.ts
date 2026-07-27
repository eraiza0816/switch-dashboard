import { test, expect } from '@playwright/test';

test.describe('Network Map', () => {
  test('loads topology and renders SVG with nodes', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('h1')).toContainText('Network Map');
    await expect(page.locator('#map-canvas')).toBeVisible();
    await expect(page.locator('#map-canvas svg')).toBeVisible({ timeout: 10000 });
    const nodes = page.locator('.node-group');
    await expect(nodes).toHaveCount(4);
  });

  test('shows switch and client nodes', async ({ page }) => {
    await page.goto('/map');
    const switchNode = page.locator('.node-group').first();
    await expect(switchNode).toContainText('S');
    const nodeTexts = await page.locator('.node-group text').allTextContents();
    const hasSwitchName = nodeTexts.some(t => t.includes('Test Switch'));
    expect(hasSwitchName).toBeTruthy();
  });

  test('clicking a switch node opens sidebar', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('#sidebar')).not.toBeVisible();
    const node = page.locator('.node-group').first();
    await node.dispatchEvent('mousedown', { button: 0 });
    await expect(page.locator('#sidebar')).toBeVisible({ timeout: 5000 });
    await expect(page.locator('#sidebar h3')).toContainText('Test Switch');
  });

  test('clicking a client node opens sidebar', async ({ page }) => {
    await page.goto('/map');
    const clientNode = page.locator('.node-group').nth(1);
    await clientNode.dispatchEvent('mousedown', { button: 0 });
    await expect(page.locator('#sidebar')).toBeVisible({ timeout: 5000 });
  });

  test('drag-and-drop moves a node', async ({ page }) => {
    await page.goto('/map');
    const node = page.locator('.node-group').first();
    await expect(node).toBeVisible({ timeout: 10000 });
    const box = await node.boundingBox();
    expect(box).toBeTruthy();

    const cx = box!.x + box!.width / 2;
    const cy = box!.y + box!.height / 2;

    await page.mouse.move(cx, cy);
    await page.mouse.down();
    await page.mouse.move(cx + 200, cy + 100, { steps: 30 });
    await page.mouse.up();

    const newBox = await node.boundingBox();
    expect(newBox).toBeTruthy();
    const movedX = Math.abs(newBox!.x - box!.x);
    expect(movedX).toBeGreaterThan(50);
  });

  test('search input is visible and accepts text', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('#search-input')).toBeVisible();
    await page.locator('#search-input').fill('Test');
    await page.waitForTimeout(300);
    await expect(page.locator('#search-results')).toBeVisible();
  });

  test('toggle clients link is present', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('a[href*="toggleClients"]')).toBeVisible();
  });

  test('reset layout link is present', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('a[href*="resetLayout"]')).toBeVisible();
  });

  test('bulk rename link is present', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('a[href*="bulkRename"]')).toBeVisible();
  });

  test('language selector is present on map page', async ({ page }) => {
    await page.goto('/map');
    await expect(page.locator('#lang-select')).toBeVisible();
  });
});
