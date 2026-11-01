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

await p.goto(base + '/referencia/componentes', { waitUntil: 'networkidle' });
const panels = await p.$$eval('.wm-tabs > .wm-tabs__panel', els => els.map(e => e.hidden));
ok(panels[0] === false && panels.slice(1).every(Boolean), 'Tabs: só o primeiro painel visível');
await p.click('.wm-tabs .wm-tabs__tab:nth-child(2)');
const after = await p.$$eval('.wm-tabs > .wm-tabs__panel', els => els.map(e => e.hidden));
ok(after[1] === false && after[0] === true, 'Tabs: clique troca o painel');
await p.keyboard.press('ArrowRight');
ok(await p.evaluate(() => document.activeElement.textContent) === 'Windows', 'Tabs: seta para a direita move o foco');

await p.click('.wm-codegroup .wm-tabs__tab:nth-child(2)');
await p.click('.wm-codegroup [data-wm-copy]');
ok((await p.evaluate(() => navigator.clipboard.readText())).startsWith('import Anthropic from'), 'CodeGroup: copia o painel ativo');
ok(await p.getAttribute('.wm-codegroup [data-wm-copy]', 'data-state') === 'copied', 'Copiar: botão confirma');

await p.click('[data-wm-theme]');
ok(await p.evaluate(() => document.documentElement.dataset.theme === 'light'), 'Tema: system → light');
await p.click('[data-wm-theme]');
await p.waitForTimeout(400);
ok(await p.evaluate(() => document.documentElement.dataset.theme === 'dark'), 'Tema: light → dark');
await p.reload({ waitUntil: 'networkidle' });
ok(await p.evaluate(() => document.documentElement.dataset.theme === 'dark'), 'Tema: escolha persiste');
await p.click('[data-wm-theme]');
await p.waitForTimeout(400);
ok(await p.evaluate(() => !document.documentElement.dataset.theme), 'Tema: dark → system');

await p.evaluate(() => document.getElementById('passos').scrollIntoView());
await p.waitForTimeout(500);
ok(await p.$eval('[data-wm-toc] a[aria-current="true"]', a => a.textContent).catch(() => null) === 'Passos', 'Sumário: destaca a seção visível');

await p.click('.wm-accordion summary');
ok(await p.$eval('.wm-accordion', d => d.open), 'Accordion: abre');

await p.setViewportSize({ width: 390, height: 800 });
await p.click('[data-wm-menu]');
ok(await p.getAttribute('[data-wm-sidebar]', 'data-open') === 'true', 'Mobile: menu abre');
await p.keyboard.press('Escape');
ok(await p.getAttribute('[data-wm-sidebar]', 'data-open') === 'false', 'Mobile: Esc fecha');

const nojs = await b.newContext({ javaScriptEnabled: false });
const q = await nojs.newPage();
await q.goto(base + '/referencia/componentes');
const visible = await q.$$eval('.wm-tabs__panel', els => els.filter(e => !e.hidden && getComputedStyle(e).display !== 'none').length);
ok(visible >= 6, `Sem JS: todos os painéis visíveis (${visible})`);

ok(errs.length === 0, 'Sem erros de JS/CSP' + (errs.length ? ': ' + errs.join(' | ') : ''));
await b.close();
process.exit(failures ? 1 : 0);
