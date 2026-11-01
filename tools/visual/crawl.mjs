// Crawls a site from its home page and checks every internal page, link and
// resource: status codes, console errors and loaded fonts.
//
//   node crawl.mjs --url http://localhost:4321/          a running server
//   node crawl.mjs --dist ../../dist --base /Wavemark    a `wavemark build` folder, served
//                                                        the way GitHub Pages serves it
import { chromium } from 'playwright';
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';

const args = Object.fromEntries(process.argv.slice(2).reduce((acc, a, i, all) => {
  if (a.startsWith('--')) acc.push([a.slice(2), all[i + 1]]);
  return acc;
}, []));

// GitHub Pages resolution: /x → x, x.html, or x/ (redirect) → x/index.html;
// anything else is 404.html with status 404.
function pagesServer(dist, base) {
  const root = path.resolve(dist);
  const types = { '.html': 'text/html; charset=utf-8', '.css': 'text/css', '.js': 'text/javascript', '.woff2': 'font/woff2', '.svg': 'image/svg+xml', '.png': 'image/png', '.md': 'text/markdown; charset=utf-8', '.txt': 'text/plain; charset=utf-8', '.sh': 'text/plain; charset=utf-8', '.ps1': 'text/plain; charset=utf-8' };
  const send = (res, file, status = 200) => { res.writeHead(status, { 'Content-Type': types[path.extname(file)] || 'application/octet-stream' }); res.end(fs.readFileSync(file)); };
  return http.createServer((req, res) => {
    const p = decodeURIComponent(new URL(req.url, 'http://x').pathname);
    const notFound = () => send(res, path.join(root, '404.html'), 404);
    if (base && p !== base && !p.startsWith(base + '/')) return notFound();
    const full = path.normalize(path.join(root, p.slice(base.length)));
    if (!full.startsWith(root)) return notFound();
    if (fs.existsSync(full) && fs.statSync(full).isFile()) return send(res, full);
    if (fs.existsSync(full) && fs.statSync(full).isDirectory()) {
      if (!p.endsWith('/')) { res.writeHead(301, { Location: p + '/' }); return res.end(); }
      if (fs.existsSync(path.join(full, 'index.html'))) return send(res, path.join(full, 'index.html'));
    }
    if (fs.existsSync(full + '.html')) return send(res, full + '.html');
    notFound();
  });
}

let server;
let startURL = args.url || 'http://localhost:4321/';
if (args.dist) {
  const base = args.base ? '/' + args.base.replace(/^\/+|\/+$/g, '') : '';
  server = pagesServer(args.dist, base);
  await new Promise(r => server.listen(0, '127.0.0.1', r));
  startURL = `http://127.0.0.1:${server.address().port}${base}/`;
}
const start = new URL(startURL);
const scope = start.pathname.endsWith('/') ? start.pathname : start.pathname + '/';

const browser = await chromium.launch();
const ctx = await browser.newContext({ viewport: { width: 1440, height: 900 } });
const page = await ctx.newPage();
const failures = [];
const resources = new Map();
page.on('response', (r) => {
  const u = new URL(r.url());
  if (u.origin === start.origin) resources.set(u.pathname, r.status());
});
page.on('console', (m) => { if (m.type() === 'error' && !/status of 404/.test(m.text())) failures.push(`console on ${page.url()}: ${m.text()}`); });
page.on('pageerror', (e) => failures.push(`js error on ${page.url()}: ${e.message}`));

const queue = [start.pathname];
const seen = new Set();
const links = new Map(); // link → page where it was found
let pages = 0;
while (queue.length) {
  const p = queue.shift();
  if (seen.has(p)) continue;
  seen.add(p);
  const res = await page.goto(start.origin + p, { waitUntil: 'networkidle' });
  if (!res || res.status() >= 400) { failures.push(`page ${p}: HTTP ${res && res.status()}`); continue; }
  pages++;
  const fontsOK = await page.evaluate(async () => { await document.fonts.ready; return document.fonts.check('16px "Hanken Grotesk"') && [...document.fonts].some(f => f.status === 'loaded'); });
  if (!fontsOK) failures.push(`page ${p}: theme fonts did not load`);
  const css = await page.evaluate(() => getComputedStyle(document.body).backgroundColor);
  if (css === 'rgba(0, 0, 0, 0)') failures.push(`page ${p}: stylesheet did not load`);
  const hrefs = await page.$$eval('a[href]', as => as.map(a => a.href));
  for (const h of hrefs) {
    const u = new URL(h);
    if (u.origin !== start.origin) continue;
    if (!links.has(u.pathname)) links.set(u.pathname, p);
    const isPage = !/\.(md|txt|png|svg|jpg|sh|ps1|css|js|woff2)$/.test(u.pathname);
    if (isPage && u.pathname.startsWith(scope) && !seen.has(u.pathname)) queue.push(u.pathname);
  }
}
// Every link must resolve (non-page links such as .md, llms.txt, install.sh).
for (const [l, from] of links) {
  if (seen.has(l)) continue;
  const r = await ctx.request.get(start.origin + l, { maxRedirects: 5 });
  if (r.status() >= 400) failures.push(`link ${l} (from ${from}): HTTP ${r.status()}`);
}
for (const [r, status] of resources) if (status >= 400 && !seen.has(r)) failures.push(`resource ${r}: HTTP ${status}`);
const nf = await ctx.request.get(start.origin + scope + 'this-does-not-exist');
if (nf.status() !== 404 || !(await nf.text()).includes('wm-')) failures.push(`missing route: expected a themed 404, got ${nf.status()}`);

console.log(`${pages} pages, ${links.size} internal links, ${resources.size} resources checked`);
for (const f of failures) console.log('✗ ' + f);
if (!failures.length) console.log('✓ no broken pages, links or resources');
await browser.close();
if (server) server.close();
process.exit(failures.length ? 1 : 0);
