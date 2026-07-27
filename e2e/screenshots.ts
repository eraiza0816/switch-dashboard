import { chromium } from '@playwright/test';

const BASE = process.env.BASE_URL || 'http://localhost:8081';
const OUT = process.env.OUT_DIR || '../images';

async function main() {
  const browser = await chromium.launch({ headless: true });
  const context = await browser.newContext({
    viewport: { width: 1440, height: 900 },
    deviceScaleFactor: 2,
  });

  const shots: { path: string; url: string; selector?: string; fullPage?: boolean }[] = [
    { path: `${OUT}/dashboard.png`, url: '/', fullPage: true },
    { path: `${OUT}/mac_table.png`, url: '/', selector: '.mac-table' },
    { path: `${OUT}/network-map.png`, url: '/map', selector: '#map-canvas' },
    { path: `${OUT}/speed_graph.png`, url: '/', selector: '#graph-overlay' },
    { path: `${OUT}/sfp_details.png`, url: '/', selector: '#transceiver-overlay' },
  ];

  for (const shot of shots) {
    const page = await context.newPage();
    await page.goto(`${BASE}${shot.url}`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(500);

    if (shot.selector) {
      const el = page.locator(shot.selector);
      if (await el.isVisible()) {
        await el.screenshot({ path: shot.path });
      } else {
        // Fallback: full page screenshot if element not visible
        await page.screenshot({ path: shot.path, fullPage: true });
      }
    } else {
      await page.screenshot({ path: shot.path, fullPage: shot.fullPage });
    }
    console.log(`  ✓ ${shot.path}`);
    await page.close();
  }

  await browser.close();
  console.log('Done');
}

main().catch(e => {
  console.error(e);
  process.exit(1);
});
