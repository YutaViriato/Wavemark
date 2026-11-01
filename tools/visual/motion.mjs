// Captures frames of the entrance and scroll animations, and checks that
// reduced motion and no-JS show everything immediately.
//
//   node motion.mjs [--wm http://localhost:4321] [--out shots/motion]
import { chromium } from 'playwright';
import fs from 'node:fs';

const args = Object.fromEntries(process.argv.slice(2).reduce((acc, a, i, all) => {
  if (a.startsWith('--')) acc.push([a.slice(2), all[i + 1]]);
  return acc;
}, []));
const WM = args.wm || 'http://localhost:4321';
const out = args.out || 'shots/motion';
fs.mkdirSync(out, { recursive: true });

const browser = await chromium.launch();
let failures = 0;
const check = (ok, msg) => { console.log((ok ? '✓ ' : '✗ ') + msg); if (!ok) failures++; };

// Hero entrance frames.
{
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const p = await ctx.newPage();
  await p.goto(WM + '/', { waitUntil: 'commit' });
  for (const ms of [150, 400, 800, 1600]) {
    await p.waitForTimeout(ms === 150 ? 150 : ms - [150, 400, 800, 1600][[150, 400, 800, 1600].indexOf(ms) - 1]);
    await p.screenshot({ path: `${out}/hero-${ms}ms.png`, clip: { x: 300, y: 150, width: 840, height: 420 } });
  }
  const words = await p.$$eval('.wm-hero__title .wm-word', els => els.map(e => getComputedStyle(e).opacity));
  check(words.length > 3 && words.every(o => o === '1'), `hero: ${words.length} palavras visíveis ao fim da animação`);
  // Scroll reveal: blocks below the fold start pending and are shown after scrolling.
  const pending = await p.$$eval('[data-wm-reveal="pending"]', els => els.length);
  check(pending > 0, `reveal: ${pending} blocos abaixo da dobra aguardando rolagem`);
  await p.evaluate(() => window.scrollTo(0, 900));
  await p.waitForTimeout(120);
  await p.screenshot({ path: `${out}/reveal-120ms.png` });
  await p.waitForTimeout(1200);
  await p.screenshot({ path: `${out}/reveal-1300ms.png` });
  const shown = await p.$$eval('[data-wm-reveal="shown"]', els => els.length);
  check(shown > 0, `reveal: ${shown} blocos revelados após rolar`);
  await ctx.close();
}

// Reduced motion: nothing pending, nothing hidden.
{
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, reducedMotion: 'reduce' });
  const p = await ctx.newPage();
  await p.goto(WM + '/', { waitUntil: 'networkidle' });
  const pending = await p.$$eval('[data-wm-reveal="pending"]', els => els.length);
  const hidden = await p.$$eval('.wm-hero__title .wm-word', els => els.filter(e => getComputedStyle(e).opacity !== '1').length);
  const anim = await p.$$eval('.wm-sketch__ink path', els => els.filter(e => getComputedStyle(e).animationName !== 'none').length);
  check(pending === 0 && hidden === 0 && anim === 0, `movimento reduzido: tudo estático (pendentes=${pending}, palavras ocultas=${hidden}, traços animando=${anim})`);
  await ctx.close();
}

// No JS: everything visible.
{
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, javaScriptEnabled: false });
  const p = await ctx.newPage();
  await p.goto(WM + '/', { waitUntil: 'networkidle' });
  const invisible = await p.$$eval('.wm-prose *', els => els.filter(e => !e.matches('.wm-anchor, .wm-anchor *') && getComputedStyle(e).opacity === '0').length); // heading anchors appear on hover by design
  check(invisible === 0, `sem JS: nenhum elemento invisível (${invisible})`);
  await ctx.close();
}

// Tab indicator follows the selected tab.
{
  const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
  const p = await ctx.newPage();
  await p.goto(WM + '/referencia/componentes', { waitUntil: 'networkidle' });
  await p.click('.wm-tabs .wm-tabs__tab:nth-child(3)');
  await p.waitForTimeout(500);
  const [tab, ind] = await p.evaluate(() => {
    const t = document.querySelector('.wm-tabs .wm-tabs__tab:nth-child(3)').getBoundingClientRect();
    const i = document.querySelector('.wm-tabs .wm-tabs__indicator').getBoundingClientRect();
    return [[Math.round(t.x), Math.round(t.width)], [Math.round(i.x), Math.round(i.width)]];
  });
  check(Math.abs(tab[0] - ind[0]) <= 1 && Math.abs(tab[1] - ind[1]) <= 1, `abas: indicador sob a aba ativa (aba ${tab}, indicador ${ind})`);
  await ctx.close();
}

await browser.close();
process.exit(failures ? 1 : 0);
