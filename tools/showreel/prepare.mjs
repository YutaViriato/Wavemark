// Collects the real product assets the showreel is cut from. Nothing is
// mocked: pages, markup and artwork all come from the wavemark binary.
//
//   node prepare.mjs
//
// Writes stage/gen/:
//   art.json        hand-drawn icons, underlines, arrows and rings (internal/sketch)
//   shots/*.png     screenshots of the docs site and of sites written by blind agents
//   frag/*.html     rendered markup of the landing page, a docs page, the
//                   component reference and `wavemark browse` on this repository
import { chromium } from 'playwright';
import { spawn, execFileSync } from 'node:child_process';
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '../..');
const gen = path.join(here, 'stage/gen');
const bin = path.join(here, 'out/wavemark');
for (const d of [gen, path.join(gen, 'shots'), path.join(gen, 'frag'), path.dirname(bin)]) fs.mkdirSync(d, { recursive: true });

execFileSync('go', ['build', '-o', bin, './cmd/wavemark'], { cwd: root, stdio: 'inherit' });
fs.writeFileSync(path.join(gen, 'art.json'), execFileSync('go', ['run', './tools/showreel/icons'], { cwd: root }));

const sites = [
  { dir: 'site', port: 4621, pages: ['/', '/guide/install', '/reference/components', '/reference/icons', '/blog/hand-drawn-by-code', '/agents'] },
  { dir: '.e2e-agents/tidepool', port: 4622, pages: ['/', '/getting-started/quickstart', '/reference/cli'] },
  { dir: '.e2e-agents/estudio', port: 4623, pages: ['/', '/pesquisa/atencao', '/produtos/mare-eval'] },
  { dir: '.e2e-agents/blind2', port: 4624, pages: ['/', '/research/sycophancy-under-pressure', '/beacon/quickstart'] },
].filter((s) => fs.existsSync(path.join(root, s.dir)));

const procs = sites.map((s) => spawn(bin, ['serve', s.dir, '--addr', `127.0.0.1:${s.port}`], { cwd: root, stdio: 'ignore' }));
procs.push(spawn(bin, ['browse', '.', '--no-open', '--addr', '127.0.0.1:4625'], { cwd: root, stdio: 'ignore' }));
await new Promise((r) => setTimeout(r, 1500));

const frags = {
  landing: 'http://127.0.0.1:4621/',
  docs: 'http://127.0.0.1:4621/guide/first-page',
  components: 'http://127.0.0.1:4621/reference/components',
  browse: 'http://127.0.0.1:4625/',
};
const browser = await chromium.launch({ args: ['--font-render-hinting=none', '--force-color-profile=srgb'] });
const manifest = [];
try {
  for (const [name, url] of Object.entries(frags)) {
    const html = await (await fetch(url)).text();
    const body = html.slice(html.indexOf('<body'), html.lastIndexOf('</body>'));
    fs.writeFileSync(path.join(gen, 'frag', `${name}.html`), body.replace(/^<body[^>]*>/, '').replace(/<script[\s\S]*?<\/script>/g, ''));
  }
  for (const scheme of ['light', 'dark']) {
    const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 }, colorScheme: scheme, deviceScaleFactor: 1 });
    const page = await ctx.newPage();
    for (const s of sites) {
      for (const p of s.pages) {
        await page.goto(`http://127.0.0.1:${s.port}${p}`, { waitUntil: 'networkidle' });
        await page.evaluate(() => document.fonts.ready);
        await page.waitForTimeout(1800);
        const name = `${path.basename(s.dir)}${p === '/' ? '' : p.replace(/\//g, '_')}-${scheme}.png`;
        await page.screenshot({ path: path.join(gen, 'shots', name) });
        manifest.push({ name, site: path.basename(s.dir), page: p, scheme });
        console.log(name);
      }
    }
    await ctx.close();
  }
} finally {
  await browser.close();
  for (const p of procs) p.kill();
}
fs.writeFileSync(path.join(gen, 'shots', 'manifest.json'), JSON.stringify(manifest, null, 2));
