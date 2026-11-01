// 19.72–22 s. 06 wavemark browse: the real browse page of this repository
// assembles, a pointer clicks the theme button and the dark theme opens from
// it in a circle; the circle keeps growing into chapter 7.
import { scene, h, set, css, clamp, lerp, ease, spring, cue, panX } from '../engine.mjs';
import { FRAG, headline, wordsIn, label, typeIn, pointer, pop } from '../kit.mjs';
import { VWHIP } from './ch5.mjs';

const WIN = { x: 880, y: 118, w: 920, h: 846 };
const S = 0.8;
export const CLICK = { t: 21.0 };
export const btnCenter = { x: 0, y: 0 };

function browsePage(dark) {
  const page = h('div', { class: 'sr-page', style: { font: '400 14px/1.5 var(--wm-font-sans)', color: 'var(--wm-fg)', height: `${(WIN.h - 44) / S}px`, transform: `scale(${S})`, colorScheme: dark ? 'dark' : 'light' } });
  page.style.width = `${WIN.w / S}px`;
  page.innerHTML = FRAG.browse;
  page.querySelector('.wm-skip')?.remove();
  page.querySelector('.wm-theme-btn')?.removeAttribute('hidden');
  page.querySelector('.wm-browse__filter')?.removeAttribute('hidden');
  for (const el of page.querySelectorAll('.wm-sidebar')) { el.style.position = 'relative'; el.style.top = '0'; el.style.height = 'auto'; }
  for (const el of page.querySelectorAll('[data-wm-ago]')) el.textContent = 'just now';
  for (const el of page.querySelectorAll('.wm-browse__card-time')) el.textContent = 'just now';
  const btn = page.querySelector('.wm-theme-btn');
  btn.dataset.mode = dark ? 'dark' : 'system';
  return { page, btn };
}

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: 'transparent' } });
  cam.append(root);
  const pan = h('div', { class: 'fill', style: { background: '#faf9f5' } });
  root.append(pan);

  const lab = label('06 — wavemark browse');
  const hl = headline(['Read every', 'Markdown file.'], { size: 104 });
  const sub = h('div', { class: 'sr-sub', style: { color: '#5e5d59', marginTop: '30px', maxWidth: '660px' }, text: 'Agents leave a lot of Markdown behind. Read all of it, in any project.' });
  const chips = ['CLAUDE.md', 'AGENTS.md', 'plan.md', 'spec.md'].map((c) => h('div', { class: 'sr-pill sr-mono', text: c, style: { font: '500 22px/1 "JetBrains Mono", monospace', height: '48px', padding: '0 18px' } }));
  const chipRow = h('div', { style: { display: 'flex', gap: '12px', marginTop: '34px' } }, ...chips);
  pan.append(h('div', { class: 'abs', style: { left: '120px', top: '300px' } }, lab, h('div', { style: { height: '30px' } }), hl.el, sub, chipRow));

  const win = h('div', { class: 'sr-window', style: { left: `${WIN.x}px`, top: `${WIN.y}px`, width: `${WIN.w}px`, height: `${WIN.h}px` } });
  const barL = h('div', { class: 'sr-bar' }, h('span', { class: 'sr-dot' }), h('span', { class: 'sr-dot' }), h('span', { class: 'sr-dot' }), h('div', { class: 'sr-url', html: '<span>localhost:4400</span>' }));
  const vpL = h('div', { class: 'sr-viewport' });
  const L = browsePage(false);
  vpL.append(L.page);
  const darkWrap = h('div', { class: 'fill', style: { borderRadius: '18px', overflow: 'hidden' } });
  const barD = h('div', { class: 'sr-bar', style: { background: '#1f1e1d', borderBottomColor: '#30302e' } }, h('span', { class: 'sr-dot', style: { background: '#3d3d3a' } }), h('span', { class: 'sr-dot', style: { background: '#3d3d3a' } }), h('span', { class: 'sr-dot', style: { background: '#3d3d3a' } }), h('div', { class: 'sr-url', html: '<span>localhost:4400</span>', style: { background: '#262624', borderColor: '#30302e', color: '#9c9a92' } }));
  const vpD = h('div', { class: 'sr-viewport', style: { background: '#151515' } });
  const D = browsePage(true);
  vpD.append(D.page);
  darkWrap.append(barD, vpD);
  win.append(barL, vpL, darkWrap);
  pan.append(win);

  // theme button centre, in window and screen coordinates
  const br = L.btn.getBoundingClientRect(), wr = win.getBoundingClientRect();
  const bx = br.left + br.width / 2 - wr.left, by = br.top + br.height / 2 - wr.top;
  btnCenter.x = WIN.x + bx; btnCenter.y = WIN.y + by;
  const ripple = h('div', { class: 'abs', style: { left: `${bx - 20}px`, top: `${by - 20}px`, width: '40px', height: '40px', borderRadius: '50%', border: '2px solid #d97757' } });
  win.append(ripple);
  const ptr = pointer();
  root.append(ptr);

  const navL = [...L.page.querySelectorAll('.wm-sidebar .wm-nav__link')];
  const navD = [...D.page.querySelectorAll('.wm-sidebar .wm-nav__link')];
  const blocks = (p) => [p.querySelector('.wm-page-header'), p.querySelector('.wm-browse__filter'), p.querySelector('.wm-browse__heading'), ...p.querySelectorAll('.wm-browse__card'), ...p.querySelectorAll('.wm-browse__folder, .wm-browse__list')].filter(Boolean);
  const blL = blocks(L.page), blD = blocks(D.page);
  const iconSpans = (p) => ({ sys: p.querySelector('.wm-theme-btn__system'), dark: p.querySelector('.wm-theme-btn__dark') });
  const icL = iconSpans(L.page);

  chips.forEach((c, i) => cue(20.25 + i * 0.125, 'pop', { note: [78, 81, 83, 86][i], pan: -0.6, vel: 0.4 }));
  cue(20.06, 'type', { n: 10, pan: 0.4, vel: 0.25 });
  cue(20.4, 'swipe', { dur: 0.5, pan: 0.6, vel: 0.2 });
  cue(CLICK.t, 'click', { pan: panX(btnCenter.x), vel: 0.9 });
  cue(CLICK.t, 'whoosh', { dur: 0.5, pan: panX(btnCenter.x), vel: 0.45, tone: 0.7 });
  cue(21.55, 'whoosh', { dur: 0.45, pan: 0.4, vel: 0.7, tone: 0.6 });

  scene({
    t0: 19.72, t1: 22.02, root,
    update(t) {
      const vw = VWHIP.at(t);
      set(pan, { y: 1080 * (1 - vw) });

      typeIn(lab, t, 19.98, 0.02);
      wordsIn(hl.words, t, 20.0, { stagger: 0.07, em: 104 });
      set(sub, { y: (1 - spring(t - 20.2, 2.4, 0.75)) * 24, o: clamp((t - 20.2) / 0.2) });
      chips.forEach((c, i) => pop(c, t, 20.25 + i * 0.125, { f: 3.8, z: 0.45, from: 0.3, dy: 14 }));

      // page assembles
      const nav = (list) => list.forEach((n, i) => { const lt = t - 20.02 - i * 0.012; set(n, { x: (1 - ease.outExpo(clamp(lt / 0.35))) * -30, o: clamp(lt / 0.08) }); });
      nav(navL); nav(navD);
      const blk = (list) => list.forEach((b, i) => { const lt = t - 20.08 - i * 0.035; set(b, { y: (1 - spring(lt, 2.6, 0.72)) * 30, o: clamp(lt / 0.1) }); });
      blk(blL); blk(blD);

      // pointer travels to the theme button and clicks on the beat
      const pu = ease.inOutCubic(clamp((t - 20.32) / 0.6));
      const p0 = { x: 1640, y: 1040 }, p1 = { x: btnCenter.x - 4, y: btnCenter.y - 4 };
      const px = lerp(p0.x, p1.x, pu) + Math.sin(pu * Math.PI) * -120, py = lerp(p0.y, p1.y, ease.outCubic(pu));
      const press = t >= CLICK.t ? 1 - 0.18 * Math.exp(-(t - CLICK.t) * 18) * clamp((t - CLICK.t) / 0.02) : 1;
      set(ptr, { x: px, y: py + 1080 * (1 - vw), s: press, o: t > 20.32 && t < 21.7 ? 1 : 0 });
      const rp = t - CLICK.t;
      set(ripple, { s: rp > 0 ? 1 + 2.5 * ease.outExpo(clamp(rp / 0.4)) : 0, o: rp > 0 ? 1 - clamp(rp / 0.4) : 0 });
      // the button icon flips from "system" to "dark"
      css(icL.sys, 'display', rp > 0.04 ? 'none' : 'inline-grid');
      css(icL.dark, 'display', rp > 0.04 ? 'inline-grid' : 'none');
      if (rp > 0.04) set(icL.dark, { r: (1 - spring(rp - 0.04, 3.2, 0.5)) * -90, s: 0.6 + 0.4 * spring(rp - 0.04, 3.2, 0.5) });

      // dark theme opens from the button
      const R = rp > 0 ? 1500 * ease.brand(clamp(rp / 0.55)) : 0;
      css(darkWrap, 'clip-path', `circle(${R.toFixed(1)}px at ${bx.toFixed(1)}px ${by.toFixed(1)}px)`);
    },
  });
}
