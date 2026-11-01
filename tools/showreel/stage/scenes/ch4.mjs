// 16–18 s. 04 Declarative components: JSX-style tags fly in on 16ths and
// unfold into the real rendered components (markup from the product's own
// reference page). Exit: a whip pan to chapter 5.
import { scene, h, set, css, draw, clamp, lerp, ease, spring, cue, panX } from '../engine.mjs';
import { FRAG, headline, wordsIn, label, typeIn, pop } from '../kit.mjs';
import { checkCenter } from './ch3.mjs';

const Z = 1.5;
export const WHIP = { t0: 17.74, dur: 0.36, at: (t) => ease.whip(clamp((t - 17.74) / 0.36)) };

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#faf9f5' } });
  cam.append(root);
  const pan = h('div', { class: 'fill' });
  root.append(pan);

  const lab = label('04 — Components');
  const hl = headline(['Declarative', 'components.'], { size: 120 });
  const sub = h('div', { class: 'sr-sub', style: { color: '#5e5d59', marginTop: '34px', maxWidth: '640px' }, text: 'Cards, tabs, steps and callouts as JSX-style tags. No JavaScript in your content.' });
  pan.append(h('div', { class: 'abs', style: { left: '120px', top: '330px' } }, lab, h('div', { style: { height: '30px' } }), hl.el, sub));

  // Real components, rendered by Wavemark.
  const comps = h('div', { html: FRAG.components });
  const land = h('div', { html: FRAG.landing });
  const pick = (root2, sel) => root2.querySelector(sel).cloneNode(true);
  const card = pick(comps, 'a.wm-card[href="/guide/install"]');
  const code = pick(land, '.wm-codegroup');
  const steps = pick(comps, 'ol.wm-steps');
  const tip = pick(comps, '.wm-callout--tip');
  const stat = [...land.querySelectorAll('.wm-stat')].find((s) => s.textContent.includes('29')).cloneNode(true);
  const badges = h('p', { html: '<span class="wm-badge wm-badge--clay">Soon</span> <span class="wm-badge wm-badge--olive">Stable</span> <span class="wm-badge wm-badge--sky">Docs</span> <span class="wm-badge wm-badge--fig">New</span> <span class="wm-badge wm-badge--gray">Planned</span>', style: { margin: 0, fontSize: '22px' } });

  const grid = h('div', { class: 'abs wm-prose wm-prose--landing', 'data-layout': 'landing', style: { left: `${880 / Z}px`, top: `${128 / Z}px`, width: `${920 / Z}px`, zoom: Z, display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '18px', alignItems: 'start', color: 'var(--wm-fg)', fontSize: '15px' } });
  const colA = h('div', { style: { display: 'flex', flexDirection: 'column', gap: '18px' } });
  const colB = h('div', { style: { display: 'flex', flexDirection: 'column', gap: '18px' } });
  steps.style.margin = '6px 0 0';
  const stepsBox = h('div', { style: { padding: '18px 18px 0', borderRadius: '16px', border: '1px solid var(--wm-border)', background: 'var(--wm-surface)' } }, steps);
  stat.style.maxWidth = 'none';
  colA.append(card, stepsBox);
  colB.append(code, tip, stat, badges);
  grid.append(colA, colB);
  pan.append(grid);
  for (const el of [card, code, tip, stat]) el.style.margin = '0';
  // CodeGroup: show one panel; the indicator is positioned here, not by wavemark.js
  const panels = [...code.querySelectorAll('.wm-tabs__panel')];
  const tabs = [...code.querySelectorAll('.wm-tabs__tab')];
  const indicator = code.querySelector('.wm-tabs__indicator');
  code.querySelector('.wm-tabs__list').setAttribute('data-indicator', '');
  panels.forEach((p, i) => (p.hidden = i !== 0));

  // Measure (zoom is applied by layout, so rects are already on screen).
  const rectOf = (el) => { const r = el.getBoundingClientRect(); return { x: r.left, y: r.top, w: r.width, h: r.height, cx: r.left + r.width / 2, cy: r.top + r.height / 2 }; };
  const tabRects = tabs.map((tb) => ({ x: tb.offsetLeft, w: tb.offsetWidth, h: tb.offsetHeight, y: tb.offsetTop }));
  const items = [
    { el: card, tag: '<Card>' },
    { el: code, tag: '<CodeGroup>' },
    { el: stepsBox, tag: '<Steps>' },
    { el: tip, tag: '<Tip>' },
    { el: stat, tag: '<Stat>' },
    { el: badges, tag: '<Badge>' },
  ].map((it, i) => ({ ...it, r: rectOf(it.el), land: 16.0 + i * 0.125 }));

  // Tag pills that fly in and become the components.
  const pills = items.map((it, i) => {
    const pill = h('div', { class: 'sr-tag abs', html: it.tag.replace('<', '&lt;<b>').replace('>', '</b>&gt;'), style: { transformOrigin: '50% 50%' } });
    root.append(pill);
    const pr = pill.getBoundingClientRect();
    it.pw = pr.width; it.ph = pr.height;
    cue(it.land - 0.16, 'zip', { pan: panX(it.r.cx), vel: 0.3 });
    cue(it.land, 'pop', { note: [74, 76, 78, 81, 83, 86][i], pan: panX(it.r.cx), vel: 0.6 });
    return pill;
  });
  const cardInk = [...card.querySelectorAll('.wm-sketch__ink path')];
  const cardFill = card.querySelector('.wm-sketch__fill');
  const rings = [...steps.querySelectorAll('.wm-step__ring path')];
  for (const p of [...cardInk, ...rings]) p.style.strokeDasharray = '1 2';
  const stepEls = [...steps.querySelectorAll('.wm-step')];
  const statVal = stat.querySelector('.wm-stat__value');
  const badgeEls = [...badges.querySelectorAll('.wm-badge')];
  const tipIcon = tip.querySelector('.wm-callout__icon');
  const TAB_T = [16.875, 17.25, 17.5];
  TAB_T.forEach((t, i) => cue(t, 'tick', { pan: panX(items[1].r.cx), vel: 0.55, note: 81 + i * 2 }));
  cue(WHIP.t0, 'whoosh', { dur: 0.42, pan: 0, vel: 0.9, tone: 1.0, from: 0.8, to: -0.8 });

  scene({
    t0: 15.74, t1: 18.12, root,
    update(t) {
      const c = checkCenter();
      const R = t < 15.75 ? 0 : lerp(30, 2300, ease.inCubic(clamp((t - 15.75) / 0.3)));
      css(root, 'clip-path', t < 16.06 ? `circle(${R.toFixed(1)}px at ${c.x}px ${c.y}px)` : 'none');

      typeIn(lab, t, 16.0, 0.02);
      wordsIn(hl.words, t, 16.0, { stagger: 0.07, em: 120 });
      set(sub, { y: (1 - spring(t - 16.25, 2.4, 0.75)) * 24, o: clamp((t - 16.25) / 0.2) });

      items.forEach((it, i) => {
        const pill = pills[i];
        const launch = it.land - 0.2;
        const lt = t - launch;
        // flight: from the left edge along an arc, landing at the component centre
        const u = ease.inOutCubic(clamp(lt / 0.2));
        const sx = 760 + i * 6, sy = 560 + (i - 2.5) * 70;
        const x = lerp(sx, it.r.cx, u), y = lerp(sy, it.r.cy, u) - Math.sin(u * Math.PI) * 140;
        const unfold = spring(t - it.land - 0.04, 2.6, 0.7);
        const gone = clamp((t - it.land - 0.05) / 0.12);
        set(pill, { x: x - it.pw / 2, y: y - it.ph / 2, r: (1 - u) * -16, s: (0.7 + 0.3 * u) * (1 + 0.25 * gone), o: (lt > 0 ? 1 : 0) * (1 - gone) });
        // the component grows out of the pill
        const ix = (it.r.w - it.pw) / 2 * (1 - unfold), iy = (it.r.h - it.ph) / 2 * (1 - unfold);
        const k = 1 / Z;
        css(it.el, 'clip-path', t < it.land ? 'inset(50% 50% 50% 50%)' : unfold > 0.995 ? 'none' : `inset(${(iy * k).toFixed(1)}px ${(ix * k).toFixed(1)}px ${(iy * k).toFixed(1)}px ${(ix * k).toFixed(1)}px round ${lerp(10, 12, unfold).toFixed(1)}px)`);
        set(it.el, { s: 0.94 + 0.06 * unfold, o: t < it.land ? 0 : 1 });
      });

      // Card
      const cT = items[0].land + 0.1;
      cardInk.forEach((p, i) => draw(p, ease.brand(clamp((t - cT - i * 0.05) / 0.4))));
      if (cardFill) set(cardFill, { o: clamp((t - cT - 0.3) / 0.1) });

      // CodeGroup tabs
      let active = 0;
      TAB_T.forEach((tt, i) => { if (t >= tt) active = i + 1; });
      const from = active > 0 ? tabRects[active - 1] : tabRects[0], to = tabRects[active];
      const sp = active > 0 ? spring(t - TAB_T[active - 1], 3.4, 0.62) : 1;
      css(indicator, 'width', `${lerp(from.w, to.w, sp).toFixed(2)}px`);
      css(indicator, 'height', `${to.h}px`);
      css(indicator, 'transform', `translateX(${lerp(from.x, to.x, sp).toFixed(2)}px)`);
      css(indicator, 'top', `${to.y}px`);
      tabs.forEach((tb, i) => tb.setAttribute('aria-selected', i === active ? 'true' : 'false'));
      panels.forEach((p, i) => { p.hidden = i !== active; });
      const pIn = active > 0 ? clamp((t - TAB_T[active - 1]) / 0.12) : 1;
      set(panels[active], { o: pIn, y: (1 - pIn) * 6 });

      // Steps
      rings.forEach((p, i) => draw(p, ease.brand(clamp((t - items[2].land - 0.12 - i * 0.12) / 0.38))));
      stepEls.forEach((s, i) => set(s, { y: (1 - spring(t - items[2].land - 0.1 - i * 0.12, 2.6, 0.7)) * 16, o: clamp((t - items[2].land - 0.1 - i * 0.12) / 0.12) }));

      // Tip
      pop(tipIcon, t, items[3].land + 0.12, { f: 4, z: 0.4 });

      // Stat counts to 29
      const sc = clamp((t - items[4].land - 0.08) / 0.5);
      statVal.textContent = String(Math.round(29 * ease.outCubic(sc)));

      // Badges pop on 16ths
      badgeEls.forEach((b, i) => pop(b, t, items[5].land + 0.1 + i * 0.0625, { f: 4.2, z: 0.42, from: 0.2 }));

      // exit
      const w = WHIP.at(t);
      set(pan, { x: -1920 * w });
      for (const p of pills) if (w > 0) set(p, { o: 0 });
    },
  });
}
