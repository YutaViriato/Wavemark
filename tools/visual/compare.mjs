// Compares computed styles of Wavemark pages with the reference sites.
//
//   node compare.mjs [--wm http://localhost:4321] [--scheme light|dark]
//
// Each check pairs an element on a reference page with the equivalent element
// on a Wavemark page and diffs the properties that define the visual identity.
// Font families always differ (the reference fonts are proprietary), so they
// are reported but never counted as mismatches.
import { chromium } from 'playwright';

const args = Object.fromEntries(process.argv.slice(2).reduce((acc, a, i, all) => {
  if (a.startsWith('--')) acc.push([a.slice(2), all[i + 1]]);
  return acc;
}, []));
const WM = args.wm || 'http://localhost:4321';
const SCHEME = args.scheme || 'light';

const DOCS = 'https://platform.claude.com/docs/en/agents-and-tools/tool-use/overview';
const LANDING = 'https://www.claude.com/';

// [label, reference page, reference selector, wavemark page, wavemark selector, properties]
const TYPE = ['fontSize', 'fontWeight', 'lineHeight', 'color'];
const BOX = ['backgroundColor', 'borderTopColor', 'borderTopWidth', 'borderRadius'];
const checks = [
  ['body', DOCS, 'body', '/guide/install', 'body', ['backgroundColor', 'color']],
  ['header', DOCS, 'header', '/guide/install', 'header.wm-header', ['height', 'backgroundColor', 'borderBottomColor', 'borderBottomWidth']],
  ['title h1', DOCS, 'main h1', '/guide/install', 'h1.wm-title', TYPE],
  ['lead', DOCS, 'main h1 ~ p, main header p', '/guide/install', '.wm-lead', TYPE],
  ['paragraph', DOCS, 'main h2 ~ p', '/guide/install', '.wm-prose p', TYPE],
  ['h2', DOCS, 'main h2', '/guide/install', '.wm-prose h2', TYPE],
  ['link', DOCS, 'main p a', '/reference/markdown', '.wm-prose p a:not([class])', ['color', 'textDecorationColor', 'textUnderlineOffset']],
  ['inline code', DOCS, 'main p code', '/reference/markdown', '.wm-prose p code', ['fontSize', 'color', 'backgroundColor', 'borderTopColor', 'borderRadius']],
  ['code block', DOCS, 'main pre', '/reference/markdown', '.wm-code__pre', ['fontSize', 'lineHeight']],
  ['sidebar width', DOCS, 'aside', '/guide/install', '.wm-sidebar', ['width']],
  ['sidebar link', DOCS, 'aside a[href*="define-tools"]', '/guide/install', '.wm-nav__link:not([aria-current])', ['fontSize', 'height', 'color', 'borderRadius']],
  ['sidebar active', DOCS, 'aside a[href$="tool-use/overview"]', '/guide/install', '.wm-nav__link[aria-current]', ['backgroundColor', 'color', 'borderRadius']],
  ['content width', DOCS, 'main h2', '/guide/install', '.wm-prose h2', ['width', 'x']],
  ['table header', DOCS, 'main th', '/reference/markdown', '.wm-prose th', ['backgroundColor', 'fontWeight', 'color', 'padding']],
  ['landing bg', LANDING, 'body', '/', 'body', ['backgroundColor', 'color']],
  ['landing h1', LANDING, 'h1', '/', '.wm-hero__title', ['fontWeight', 'color']],
  ['landing primary button', LANDING, 'text:Try Claude', '/', '.wm-button--primary', ['backgroundColor', 'color', 'borderRadius']],
];

async function measure(page, selector, props) {
  return page.evaluate(([selector, props]) => {
    // "text:Label" finds the innermost element whose text is exactly Label.
    const el = selector.startsWith('text:')
      ? [...document.querySelectorAll('a, button')].find(e => e.textContent.trim() === selector.slice(5))
      : selector.split(',').map(s => document.querySelector(s.trim())).find(Boolean);
    if (!el) return null;
    const cs = getComputedStyle(el);
    const r = el.getBoundingClientRect();
    const out = { fontFamily: cs.fontFamily.split(',')[0].replace(/"/g, '') };
    for (const p of props) {
      if (p === 'width') out.width = Math.round(r.width) + 'px';
      else if (p === 'height') out.height = Math.round(r.height) + 'px';
      else if (p === 'x') out.x = Math.round(r.x) + 'px';
      else out[p] = cs[p];
    }
    return out;
  }, [selector, props]);
}

// Normalize colors to rgba() so "rgb(…)" and "color(srgb …)" compare equal.
function norm(v) {
  if (typeof v !== 'string') return v;
  return v
    .replace(/color\(srgb ([\d.]+) ([\d.]+) ([\d.]+)(?: \/ ([\d.]+))?\)/g, (_, r, g, b, a) => `rgba(${Math.round(r * 255)}, ${Math.round(g * 255)}, ${Math.round(b * 255)}, ${a ?? 1})`)
    .replace(/rgb\((\d+), (\d+), (\d+)\)/g, 'rgba($1, $2, $3, 1)')
    .replace(/(\d+\.\d{2})\d+px/g, '$1px');
}

function close(a, b) {
  const na = parseFloat(a), nb = parseFloat(b);
  if (!Number.isNaN(na) && !Number.isNaN(nb) && /px$/.test(a) && /px$/.test(b)) return Math.abs(na - nb) <= 1;
  return a === b;
}

const browser = await chromium.launch();
const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, colorScheme: SCHEME });
const pages = new Map();
async function open(url) {
  if (!pages.has(url)) {
    const p = await ctx.newPage();
    await p.goto(url, { waitUntil: 'networkidle', timeout: 60000 }).catch(() => {});
    await p.evaluate(() => document.fonts.ready);
    pages.set(url, p);
  }
  return pages.get(url);
}

let mismatches = 0, total = 0;
for (const [label, refURL, refSel, wmPath, wmSel, props] of checks) {
  const [ref, wm] = await Promise.all([measure(await open(refURL), refSel, props), measure(await open(WM + wmPath), wmSel, props)]);
  console.log(`\n${label}`);
  if (!ref || !wm) {
    console.log(`  ! element not found (${!ref ? 'reference: ' + refSel : 'wavemark: ' + wmSel})`);
    continue;
  }
  console.log(`  font: ${ref.fontFamily} → ${wm.fontFamily}`);
  for (const p of props) {
    const a = norm(ref[p]), b = norm(wm[p]);
    const ok = close(a, b);
    total++;
    if (!ok) mismatches++;
    console.log(`  ${ok ? '✓' : '✗'} ${p.padEnd(20)} ${String(a).padEnd(34)} ${ok ? '' : '→ ' + b}`);
  }
}
console.log(`\n${total - mismatches}/${total} properties match (${SCHEME})`);
await browser.close();
