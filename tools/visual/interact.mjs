// Exercises the interactive parts of the theme in a real browser.
//
//   node interact.mjs [--wm http://localhost:4321]
import { chromium } from 'playwright';

const args = Object.fromEntries(process.argv.slice(2).reduce((acc, a, i, all) => {
  if (a.startsWith('--')) acc.push([a.slice(2), all[i + 1]]);
  return acc;
}, []));
const base = args.wm || 'http://localhost:4321';
const b = await chromium.launch();
const ctx = await b.newContext({ viewport: { width: 1440, height: 900 }, permissions: ['clipboard-read', 'clipboard-write'] });
const p = await ctx.newPage();
const errs = [];
p.on('pageerror', e => errs.push(e.message));
p.on('console', m => { if (m.type() === 'error') errs.push(m.text()); });
let failures = 0;
const ok = (cond, msg) => { console.log((cond ? '✓ ' : '✗ ') + msg); if (!cond) failures++; };

await p.goto(base + '/reference/components', { waitUntil: 'networkidle' });
const panels = await p.$$eval('.wm-tabs > .wm-tabs__panel', els => els.map(e => e.hidden));
ok(panels[0] === false && panels.slice(1).every(Boolean), 'Tabs: only the first panel is visible');
await p.click('.wm-tabs .wm-tabs__tab:nth-child(2)');
const after = await p.$$eval('.wm-tabs > .wm-tabs__panel', els => els.map(e => e.hidden));
ok(after[1] === false && after[0] === true, 'Tabs: a click switches the panel');
await p.keyboard.press('ArrowRight');
ok(await p.evaluate(() => document.activeElement.textContent) === 'Windows', 'Tabs: the right arrow moves focus');

await p.click('.wm-codegroup .wm-tabs__tab:nth-child(2)');
await p.click('.wm-codegroup [data-wm-copy]');
ok((await p.evaluate(() => navigator.clipboard.readText())).startsWith('import Anthropic from'), 'CodeGroup: copies the active panel');
ok(await p.getAttribute('.wm-codegroup [data-wm-copy]', 'data-state') === 'copied', 'Copy: the button confirms');

await p.click('[data-wm-theme]');
ok(await p.evaluate(() => document.documentElement.dataset.theme === 'light'), 'Theme: system → light');
await p.click('[data-wm-theme]');
await p.waitForTimeout(400);
ok(await p.evaluate(() => document.documentElement.dataset.theme === 'dark'), 'Theme: light → dark');
await p.reload({ waitUntil: 'networkidle' });
ok(await p.evaluate(() => document.documentElement.dataset.theme === 'dark'), 'Theme: the choice persists');
await p.click('[data-wm-theme]');
await p.waitForTimeout(400);
ok(await p.evaluate(() => !document.documentElement.dataset.theme), 'Theme: dark → system');

await p.evaluate(() => document.getElementById('steps').scrollIntoView());
await p.waitForTimeout(500);
ok(await p.$eval('[data-wm-toc] a[aria-current="true"]', a => a.textContent).catch(() => null) === 'Steps', 'Table of contents: highlights the visible section');

await p.click('.wm-accordion summary');
ok(await p.$eval('.wm-accordion', d => d.open), 'Accordion: opens');

await p.setViewportSize({ width: 390, height: 800 });
await p.click('[data-wm-menu]');
ok(await p.getAttribute('[data-wm-sidebar]', 'data-open') === 'true', 'Mobile: the menu opens');
await p.keyboard.press('Escape');
ok(await p.getAttribute('[data-wm-sidebar]', 'data-open') === 'false', 'Mobile: Esc closes');

const nojs = await b.newContext({ javaScriptEnabled: false });
const q = await nojs.newPage();
await q.goto(base + '/reference/components');
const visible = await q.$$eval('.wm-tabs__panel', els => els.filter(e => !e.hidden && getComputedStyle(e).display !== 'none').length);
ok(visible >= 6, `No JS: every panel is visible (${visible})`);

ok(errs.length === 0, 'No JS or CSP errors' + (errs.length ? ': ' + errs.join(' | ') : ''));
await b.close();
process.exit(failures ? 1 : 0);
