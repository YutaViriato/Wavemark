// 12–14 s. 02 No build step: edit the H1 in the editor, save, the real docs
// page updates in place; then the view pulls out to 600 pages rebuilding in a
// half-second ripple, and one page turns dark (an error) for chapter 3.
import { scene, h, set, css, clamp, lerp, ease, spring, cue, panX, rng } from '../engine.mjs';
import { FRAG, headline, wordsIn, wordsOut, label, typeIn, codeWindow, typeRow, pop } from '../kit.mjs';

export const CH2_EDITOR = { x: 120, y: 476, w: 680, h: 262 };
const WIN = { x: 900, y: 330, w: 920, h: 620 };
const S = 0.94;
const GRID = { cols: 30, rows: 20, x0: 18, y0: 22, cw: 63.2, ch: 52.4, tw: 54, th: 42 };
export const ERR_TILE = { c: 21, r: 11 };
export const errTileCenter = () => ({ x: GRID.x0 + ERR_TILE.c * GRID.cw + GRID.tw / 2, y: GRID.y0 + ERR_TILE.r * GRID.ch + GRID.th / 2 });

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#faf9f5' } });
  cam.append(root);
  const scale = h('div', { class: 'fill', style: { transformOrigin: '0 0' } });
  root.append(scale);

  const lab = label('02 — Runtime');
  const hl = headline(['No build step.']);
  const head = h('div', { class: 'abs', style: { left: '120px', top: '196px' } }, lab, h('div', { style: { height: '30px' } }), hl.el);
  scale.append(head);

  // Editor
  const ed = codeWindow({ ...CH2_EDITOR, tab: 'first-page.md', lines: ['# Your first page', '', 'Create a `.md` file anywhere in the folder.', 'The first # heading becomes the page title.'], font: 22, lh: 42 });
  ed.body.style.padding = '22px 0';
  scale.append(ed.el);
  const r0 = ed.rows[0];
  r0.code.innerHTML = '<span class="tk-hm">#</span> <span class="tk-h">Your first </span><span class="tk-h sr-old">page</span><span class="tk-h sr-new"></span>';
  const oldWord = r0.code.querySelector('.sr-old');
  const newWord = r0.code.querySelector('.sr-new');
  const sel = h('div', { class: 'abs', style: { height: '34px', top: '4px', background: 'rgb(106 155 204 / .38)', borderRadius: '3px' } });
  r0.row.style.position = 'relative';
  r0.code.style.position = 'relative';
  r0.code.prepend(sel);
  const ecaret = h('div', { class: 'abs', style: { width: '3px', height: '32px', top: '5px', background: '#faf9f5', borderRadius: '2px' } });
  r0.code.append(ecaret);
  const log = h('div', { class: 'abs sr-mono', style: { left: '124px', top: '762px', font: '400 19px/1.4 "JetBrains Mono", monospace', color: '#87867f', whiteSpace: 'nowrap' }, html: '<span style="color:#5e5d59">wavemark 12:04:31</span> rebuilt in 2ms · 17 pages · 0 error(s) · 0 warning(s)' });
  scale.append(log);

  // Browser with the real docs page (guide/first-page)
  const win = h('div', { class: 'sr-window', style: { left: `${WIN.x}px`, top: `${WIN.y}px`, width: `${WIN.w}px`, height: `${WIN.h}px` } });
  const url = h('div', { class: 'sr-url', html: '<span>localhost:4321/guide/first-page</span>' });
  const bar = h('div', { class: 'sr-bar' }, h('span', { class: 'sr-dot' }), h('span', { class: 'sr-dot' }), h('span', { class: 'sr-dot' }), url);
  const vp = h('div', { class: 'sr-viewport' });
  const page = h('div', { class: 'sr-page', style: { font: '400 14px/1.5 var(--wm-font-sans)', color: 'var(--wm-fg)', height: '1200px', transform: `scale(${S})` } });
  page.innerHTML = FRAG.docs;
  page.querySelector('.wm-skip')?.remove();
  page.querySelector('.wm-theme-btn')?.removeAttribute('hidden');
  page.style.width = `${WIN.w / S}px`;
  page.querySelector('.wm-toc')?.remove();
  page.querySelector('.wm-shell')?.classList.add('wm-shell--no-toc');
  for (const el of page.querySelectorAll('.wm-sidebar')) { el.style.position = 'relative'; el.style.top = '0'; el.style.height = 'auto'; }
  vp.append(page);
  win.append(bar, vp);
  scale.append(win);
  const title = page.querySelector('.wm-title');
  title.innerHTML = 'Your first <span class="sr-swap" style="display:inline-block;position:relative"><span class="sr-a" style="display:inline-block">page</span><span class="sr-b" style="display:inline-block;position:absolute;left:0;top:0">site</span></span>';
  const tA = title.querySelector('.sr-a'), tB = title.querySelector('.sr-b');
  const glow = h('div', { class: 'abs', style: { left: '-14px', top: '-8px', right: '-14px', bottom: '-8px', borderRadius: '10px', background: 'rgb(217 119 87 / .16)', zIndex: -1 } });
  title.style.position = 'relative';
  title.style.isolation = 'isolate';
  title.prepend(glow);
  const ripple = h('div', { class: 'abs', style: { left: '0', top: '0', width: '40px', height: '40px', marginLeft: '-20px', marginTop: '-20px', borderRadius: '50%', border: '3px solid #d97757' } });
  win.append(ripple);

  // 600 page tiles (the landing stat: ~0.5 s to rebuild 600 pages)
  const grid = h('div', { class: 'fill' });
  root.append(grid);
  const r = rng(21);
  const tiles = [];
  const ctr = errTileCenter();
  for (let rr = 0; rr < GRID.rows; rr++) for (let c = 0; c < GRID.cols; c++) {
    const x = GRID.x0 + c * GRID.cw, y = GRID.y0 + rr * GRID.ch;
    const dark = r() < 0.09;
    const el = h('div', { class: 'abs', style: { left: `${x}px`, top: `${y}px`, width: `${GRID.tw}px`, height: `${GRID.th}px`, borderRadius: '5px', background: dark ? '#262624' : '#fff', boxShadow: '0 0 0 1px rgb(20 20 19 / .07), 0 6px 14px -8px rgb(20 20 19 / .28)', backgroundImage: dark ? 'linear-gradient(#5e5d59,#5e5d59), linear-gradient(#3d3d3a,#3d3d3a), linear-gradient(#3d3d3a,#3d3d3a)' : 'linear-gradient(#b0aea5,#b0aea5), linear-gradient(#e8e6dc,#e8e6dc), linear-gradient(#e8e6dc,#e8e6dc)', backgroundSize: `${18 + r() * 14}px 4px, ${30 + r() * 12}px 3px, ${22 + r() * 16}px 3px`, backgroundPosition: '8px 9px, 8px 20px, 8px 28px', backgroundRepeat: 'no-repeat' } });
    const d = Math.hypot((x + GRID.tw / 2 - 960) / 1.2, y + GRID.th / 2 - 540);
    grid.append(el);
    tiles.push({ el, d, c, r: rr, err: c === ERR_TILE.c && rr === ERR_TILE.r });
  }
  const maxD = Math.max(...tiles.map((x) => x.d));
  const errTile = tiles.find((x) => x.err);
  errTile.el.style.zIndex = 2;
  const errDot = h('div', { class: 'abs', style: { left: `${GRID.tw - 14}px`, top: '6px', width: '8px', height: '8px', borderRadius: '50%', background: '#e0786c' } });
  errTile.el.append(errDot);

  const counter = h('div', { style: { height: '96px', padding: '0 40px', display: 'flex', alignItems: 'center', gap: '22px', borderRadius: '999px', background: '#141413', color: '#faf9f5', boxShadow: '0 24px 60px -20px rgb(20 20 19 / .6)', whiteSpace: 'nowrap', transformOrigin: '50% 50%' } });
  const cNum = h('span', { style: { font: '400 58px/1 "Newsreader", serif', letterSpacing: '-0.02em', minWidth: '96px', textAlign: 'right', fontVariantNumeric: 'tabular-nums' }, text: '0' });
  const cTxt = h('span', { style: { font: '400 28px/1 "Hanken Grotesk", sans-serif', color: '#c2c0b6' }, text: 'pages rebuilt in' });
  const cSec = h('span', { style: { font: '400 58px/1 "Newsreader", serif', letterSpacing: '-0.02em', minWidth: '130px', fontVariantNumeric: 'tabular-nums' }, text: '0.00 s' });
  counter.append(cNum, cTxt, cSec);
  root.append(h('div', { class: 'abs', style: { left: '0', top: '868px', width: '1920px', display: 'flex', justifyContent: 'center' } }, counter));

  // ---- cues
  cue(12.0, 'type', { n: 8, pan: -0.5, vel: 0.35 });
  cue(12.04, 'whoosh', { dur: 0.45, pan: 0.7, vel: 0.45, tone: 1.2, from: 0.9, to: 0.3 });
  [12.5, 12.625, 12.75, 12.875].forEach((t) => cue(t, 'key', { pan: -0.45, vel: 0.75 }));
  cue(13.0, 'save', { pan: -0.4, vel: 0.8 });
  cue(13.0, 'pop', { note: 86, pan: 0.45, vel: 0.6 });
  cue(13.25, 'whoosh', { dur: 0.5, vel: 0.55, tone: 0.8, rev: false });
  cue(13.25, 'ripple', { dur: 0.5, vel: 0.6 });
  cue(13.75, 'error', { pan: panX(ctr.x), vel: 0.6, soft: true });
  cue(13.8, 'whoosh', { dur: 0.28, vel: 0.9, tone: 0.7, rev: true, pan: panX(ctr.x) });

  scene({
    t0: 12.0, t1: 14.06, root,
    update(t) {
      typeIn(lab, t, 12.0, 0.02);
      wordsIn(hl.words, t, 12.04, { stagger: 0.06 });
      ed.rows.forEach((row, i) => (i === 0 ? set(row.row, { o: 1 }) : typeRow(row, t, 12.02 + i * 0.04, 0.004, ed.cw)));

      // select "page", type "site" on 16ths, save on the beat
      const oldW = 4 * ed.cw;
      const selP = ease.outExpo(clamp((t - 12.3) / 0.14));
      const typed = t < 12.5 ? 0 : Math.min(4, Math.floor((t - 12.5) / 0.125) + 1);
      css(sel, 'left', `${(11 * ed.cw + oldW * (1 - selP)).toFixed(1)}px`);
      css(sel, 'width', `${(oldW * selP).toFixed(1)}px`);
      set(sel, { o: t >= 12.3 && t < 12.5 ? 1 : 0 });
      css(oldWord, 'display', t < 12.5 ? 'inline' : 'none');
      newWord.textContent = 'site'.slice(0, typed);
      const cx = t < 12.3 ? 15 * ed.cw : t < 12.5 ? (11 + 4 * (1 - selP)) * ed.cw : (11 + typed) * ed.cw;
      set(ecaret, { x: cx, o: t < 13.15 ? ((t * 2) % 1 < 0.55 || (t > 12.25 && t < 13) ? 1 : 0.1) : 0 });
      css(ed.tabDot, 'display', t >= 12.5 && t < 13.0 ? 'block' : 'none');
      set(log, { o: clamp((t - 13.02) / 0.06), x: (1 - ease.outExpo(clamp((t - 13.02) / 0.35))) * -30 });

      // browser
      const wi = spring(t - 12.06, 1.8, 0.74);
      set(win, { x: (1 - wi) * 760, r: (1 - wi) * 5, o: t < 12.06 ? 0 : 1 });
      const sw = t - 13.0;
      set(tA, { y: sw > 0 ? -ease.inCubic(clamp(sw / 0.12)) * 30 : 0, o: sw > 0 ? 1 - clamp(sw / 0.1) : 1, blur: sw > 0 ? clamp(sw / 0.12) * 3 : 0 });
      set(tB, { y: sw > 0 ? (1 - spring(sw - 0.05, 3, 0.55)) * 30 : 30, o: clamp((sw - 0.05) / 0.08) });
      set(glow, { o: sw > 0 ? Math.exp(-sw * 3) * clamp(sw / 0.05) : 0, s: 1 + 0.06 * (1 - Math.exp(-sw * 6)) });
      const tr = title.getBoundingClientRect(), wr = win.getBoundingClientRect();
      void tr; void wr;
      css(ripple, 'left', `${WIN.w - 60}px`); css(ripple, 'top', '22px');
      set(ripple, { s: sw > 0 ? 1 + 5 * ease.outExpo(clamp(sw / 0.5)) : 0, o: sw > 0 ? 1 - clamp(sw / 0.5) : 0 });

      // pull out to the grid
      const po = ease.inOutCubic(clamp((t - 13.2) / 0.32));
      // the browser window shrinks into the error tile's neighbour at the centre
      const centreTile = tiles.find((x) => x.c === 14 && x.r === 9);
      const ct = { x: GRID.x0 + 14 * GRID.cw, y: GRID.y0 + 9 * GRID.ch };
      const k = lerp(1, GRID.tw / WIN.w, po);
      css(scale, 'transform', `translate(${lerp(0, ct.x - WIN.x * k, po).toFixed(2)}px,${lerp(0, ct.y - WIN.y * k, po).toFixed(2)}px) scale(${k.toFixed(5)})`);
      set(scale, { o: 1 - clamp((t - 13.46) / 0.06) });
      wordsOut(hl.words, t, 13.2, { stagger: 0.02, dur: 0.18 });
      set(lab, { o: 1 - clamp((t - 13.2) / 0.1) });
      void centreTile;

      // ripple of 600 rebuilt pages, radiating from the centre in 0.5 s
      const zoom = ease.inExpo(clamp((t - 13.8) / 0.26));
      const zk = lerp(1, 46, zoom);
      css(grid, 'transform', `translate(${(ctr.x - ctr.x * zk).toFixed(2)}px,${(ctr.y - ctr.y * zk).toFixed(2)}px) scale(${zk.toFixed(4)})`);
      css(grid, 'transform-origin', '0 0');
      for (const tl of tiles) {
        const t0 = 13.25 + (tl.d / maxD) * 0.45;
        const lt = t - t0;
        if (lt < 0) { set(tl.el, { o: 0 }); continue; }
        const p = spring(lt, 3.2, 0.55);
        set(tl.el, { s: 0.3 + 0.7 * p, ry: (1 - clamp(lt / 0.18)) * 80, persp: 300, o: clamp(lt / 0.05) });
        if (tl.err) {
          const e = clamp((t - 13.75) / 0.06);
          css(tl.el, 'background-color', e > 0 ? '#141413' : '#fff');
          set(errDot, { o: e, s: 1 + 0.6 * Math.exp(-(t - 13.75) * 10) });
          if (e > 0) set(tl.el, { s: 1 + 0.25 * Math.exp(-(t - 13.75) * 8) * (1 - zoom) });
        }
      }

      // live counter
      const cp = clamp((t - 13.25) / 0.5);
      pop(counter, t, 13.25, { f: 3, z: 0.6, from: 0.6, dy: 40 });
      cNum.textContent = String(Math.round(600 * ease.outQuad(cp)));
      cSec.textContent = `${(0.5 * cp).toFixed(2)} s`;
      set(counter, { o: (t >= 13.25 ? 1 : 0) * (1 - clamp((t - 13.8) / 0.1)) });
    },
  });
}
