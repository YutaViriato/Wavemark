// Screenshots of Wavemark pages for visual review.
//
//   node shots.mjs [--wm http://localhost:4321] [--out shots] [--pages /,/guia] [--scheme light,dark] [--width 1440] [--scroll 0,900]
import { chromium } from 'playwright';
import fs from 'node:fs';

const args = Object.fromEntries(process.argv.slice(2).reduce((acc, a, i, all) => {
  if (a.startsWith('--')) acc.push([a.slice(2), all[i + 1]]);
  return acc;
}, []));
const WM = args.wm || 'http://localhost:4321';
const out = args.out || 'shots';
const pages = (args.pages || '/,/guia/instalacao,/referencia/componentes').split(',');
const schemes = (args.scheme || 'light').split(',');
const width = +(args.width || 1440);
const scrolls = (args.scroll || '0').split(',').map(Number);
fs.mkdirSync(out, { recursive: true });

const browser = await chromium.launch();
let failed = false;
for (const scheme of schemes) {
  const ctx = await browser.newContext({ viewport: { width, height: 900 }, colorScheme: scheme, deviceScaleFactor: 1 });
  const page = await ctx.newPage();
  page.on('pageerror', (e) => { failed = true; console.error('erro de JS:', e.message); });
  page.on('console', (m) => { if (m.type() === 'error') { failed = true; console.error('console:', m.text()); } });
  for (const path of pages) {
    await page.goto(WM + path, { waitUntil: 'networkidle' });
    await page.evaluate(() => document.fonts.ready);
    const name = path === '/' ? 'home' : path.slice(1).replace(/\//g, '_');
    for (const y of scrolls) {
      await page.evaluate((y) => window.scrollTo(0, y), y);
      await page.waitForTimeout(200);
      const file = `${out}/${name}-${scheme}-${width}-${y}.png`;
      await page.screenshot({ path: file });
      console.log(file);
    }
  }
  await ctx.close();
}
await browser.close();
process.exit(failed ? 1 : 0);
