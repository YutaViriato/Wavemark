// 17.74–20.1 s. 05 Hand-drawn, by code: the product's icon set draws itself
// in a diagonal wave on 32nds, then the camera dives into one icon to show the
// geometry under the pen: nudged corners, bowed edges, the fill out of register.
import { scene, h, set, css, draw, clamp, lerp, ease, spring, cue, panX } from '../engine.mjs';
import { headline, wordsIn, label, typeIn, icon, drawIcon, svgLayer } from '../kit.mjs';
import { WHIP } from './ch4.mjs';

export const VWHIP = { t0: 19.72, dur: 0.38, at: (t) => ease.whip(clamp((t - 19.72) / 0.38)) };
const NAMES = ['book', 'code', 'terminal', 'rocket', 'gear', 'bolt', 'shield', 'key', 'cloud', 'database', 'folder', 'file', 'globe', 'message', 'user', 'chart', 'search', 'sparkles', 'heart', 'check', 'clock', 'compass', 'flask', 'sprout', 'layers', 'lock', 'mail', 'map', 'pencil', 'play', 'star', 'wand', 'wave', 'download', 'box', 'cpu', 'lightbulb', 'target', 'flag', 'calendar', 'server', 'home'];
const COLS = 7, CELL = 116, GX = 1000, GY = 152, SIZE = 84;
const HERO = 'pencil';

export async function build({ cam }) {
  // The clay background travels with the whip pans, so it lives on `pan`.
  const root = h('div', { class: 'sc', style: { background: 'transparent' } });
  cam.append(root);
  const pan = h('div', { class: 'fill', style: { background: '#d97757' } });
  root.append(pan);

  const lab = label('05 — Sketch', { color: '#141413' });
  lab.style.opacity = 0.75;
  const hl = headline(['Hand-drawn,', 'by code.'], { color: '#141413', size: 124 });
  const sub = h('div', { class: 'sr-sub', style: { color: '#30302e', marginTop: '34px', maxWidth: '640px' }, text: 'Every icon is geometry traced by a deterministic pen. Every stroke can animate.' });
  pan.append(h('div', { class: 'abs', style: { left: '120px', top: '330px' } }, lab, h('div', { style: { height: '30px' } }), hl.el, sub));

  const field = h('div', { class: 'fill', style: { transformOrigin: '0 0' } });
  pan.append(field);
  const fill = '#efc2b1';
  const ics = NAMES.map((n, i) => {
    const c = i % COLS, r = Math.floor(i / COLS);
    const ic = icon(n, { size: SIZE, color: '#141413', fill });
    ic.el.classList.add('abs');
    Object.assign(ic.el.style, { left: `${GX + c * CELL + (CELL - SIZE) / 2}px`, top: `${GY + r * CELL + (CELL - SIZE) / 2}px`, transformOrigin: '50% 50%' });
    field.append(ic.el);
    return { ...ic, c, r, cx: GX + c * CELL + CELL / 2, cy: GY + r * CELL + CELL / 2, t0: 18.02 + (c + r) * 0.045 };
  });
  const hero = ics.find((x) => x.name === HERO);

  // Construction view of the pencil: the ideal geometry from internal/sketch
  // (G(32,7, 41,16, 17,40, 7,41, 8,31) and L(28,11, 37,20)) under the pen.
  const BIG = 520, HX = 1250, HY = 560, BX = HX - BIG / 2, BY = HY - BIG / 2;
  const cons = svgLayer(root);
  const u = BIG / 48;
  const P = (x, y) => [BX + x * u, BY + y * u];
  const poly = [[32, 7], [41, 16], [17, 40], [7, 41], [8, 31]].map(([x, y]) => P(x, y));
  const ideal = h('path', { d: `M${poly.map((p) => p.join(' ')).join(' L')} Z M${P(28, 11).join(' ')} L${P(37, 20).join(' ')}`, fill: 'none', stroke: '#141413', 'stroke-width': 1.6, 'stroke-dasharray': '7 9', opacity: 0 });
  cons.append(ideal);
  const vdots = poly.map(([x, y]) => { const d = h('circle', { cx: x, cy: y, r: 7, fill: '#faf9f5', stroke: '#141413', 'stroke-width': 2, opacity: 0 }); cons.append(d); return d; });
  const notes = [
    { text: 'nudged corners', at: poly[1], dx: 60, dy: -60 },
    { text: 'bowed edges', at: [(poly[1][0] + poly[2][0]) / 2, (poly[1][1] + poly[2][1]) / 2], dx: 90, dy: 30 },
    { text: 'fill out of register', at: P(14, 33), dx: -40, dy: 110 },
  ].map((n, i) => {
    const el = h('div', { class: 'abs sr-mono', text: n.text, style: { font: '500 22px/1 "JetBrains Mono", monospace', color: '#141413', background: '#faf9f5', padding: '9px 14px', borderRadius: '10px', whiteSpace: 'nowrap', boxShadow: '0 10px 24px -12px rgb(20 20 19 / .5)' } });
    root.append(el);
    const line = h('path', { d: `M${n.at[0]} ${n.at[1]} Q${n.at[0] + n.dx * 0.2} ${n.at[1] + n.dy * 0.8} ${n.at[0] + n.dx} ${n.at[1] + n.dy}`, fill: 'none', stroke: '#141413', 'stroke-width': 2, 'stroke-linecap': 'round', pathLength: 1, 'stroke-dasharray': '1 2' });
    cons.append(line);
    return { ...n, el, line, t0: 19.36 + i * 0.125 };
  });
  const penTip = h('div', { class: 'abs', style: { width: '18px', height: '18px', marginLeft: '-9px', marginTop: '-9px', borderRadius: '50%', background: '#141413', boxShadow: '0 0 0 6px rgb(250 249 245 / .5)' } });
  root.append(penTip);
  const heroPath = hero.ink[0];

  cue(WHIP.t0, 'whoosh', { dur: 0.3, pan: 0.6, vel: 0.3, tone: 1.6 });
  for (let d = 0; d < 12; d += 2) cue(18.02 + d * 0.045, 'scribble', { pan: panX(GX + d * 60), vel: 0.22 });
  cue(19.0, 'whoosh', { dur: 0.4, pan: 0.5, vel: 0.55, tone: 0.9, rev: true });
  notes.forEach((n) => cue(n.t0, 'pop', { note: 86, pan: 0.5, vel: 0.35 }));
  cue(19.25, 'scribble', { pan: 0.5, vel: 0.45, dur: 0.45 });
  cue(VWHIP.t0, 'whoosh', { dur: 0.42, pan: 0, vel: 0.9, tone: 0.85 });

  scene({
    t0: 17.74, t1: 20.12, root,
    update(t) {
      const w = WHIP.at(t);
      const vw = VWHIP.at(t);
      set(pan, { x: 1920 * (1 - w), y: -1080 * vw });
      set(cons, { x: 1920 * (1 - w), y: -1080 * vw });

      typeIn(lab, t, 18.0, 0.02);
      wordsIn(hl.words, t, 18.0, { stagger: 0.07, em: 124 });
      set(sub, { y: (1 - spring(t - 18.2, 2.4, 0.75)) * 24, o: clamp((t - 18.2) / 0.2) });

      // diagonal draw-on wave
      for (const ic of ics) {
        drawIcon(ic, t, ic.t0, { dur: 0.42, stagger: 0.05, fillAt: 0.26 });
        const p = spring(t - ic.t0, 3, 0.5);
        set(ic.el, { s: t < ic.t0 ? 0.6 : 0.6 + 0.4 * p, o: t < ic.t0 ? 0 : 1 });
      }

      // dive into the pencil
      const z = ease.inOutCubic(clamp((t - 19.0) / 0.34));
      const k = lerp(1, BIG / SIZE, z);
      const tx = lerp(0, HX - hero.cx * k, z), ty = lerp(0, HY - hero.cy * k, z);
      css(field, 'transform', `translate(${tx.toFixed(2)}px,${ty.toFixed(2)}px) scale(${k.toFixed(4)})`);
      for (const ic of ics) if (ic !== hero) set(ic.el, { o: t < ic.t0 ? 0 : 1 - clamp((t - 19.0) / 0.18) });
      set(hl.el, { o: 1 }); // the headline stays readable on the left

      // the pen re-traces the pencil over its ideal geometry
      const ci = clamp((t - 19.3) / 0.12);
      ideal.setAttribute('opacity', (ci * 0.6).toFixed(3));
      vdots.forEach((d, i) => { const p = spring(t - 19.3 - i * 0.03, 4, 0.5); d.setAttribute('r', (7 * p).toFixed(2)); d.setAttribute('opacity', p > 0 ? '1' : '0'); });
      const trace = clamp((t - 19.25) / 0.45);
      if (t > 19.25) {
        draw(heroPath, ease.inOutSine(trace));
        hero.ink.slice(1).forEach((p) => draw(p, ease.brand(clamp((t - 19.62) / 0.15))));
        const len = heroPath.getTotalLength();
        const pt = heroPath.getPointAtLength(len * ease.inOutSine(trace));
        const m = heroPath.getScreenCTM();
        const sp = pt.matrixTransform(m);
        set(penTip, { x: sp.x, y: sp.y, o: trace > 0 && trace < 1 ? 1 : clamp(1 - (t - 19.7) / 0.1) });
      } else set(penTip, { o: 0 });
      notes.forEach((n) => {
        draw(n.line, ease.brand(clamp((t - n.t0) / 0.2)));
        const p = spring(t - n.t0 - 0.08, 3.6, 0.55);
        set(n.el, { x: n.at[0] + n.dx + (n.dx < 0 ? -230 : 10), y: n.at[1] + n.dy - 20 + (1 - p) * 20 - 1080 * vw, s: 0.6 + 0.4 * p, o: t > n.t0 + 0.08 ? 1 : 0 });
      });
      if (vw > 0) set(penTip, { o: 0 });
    },
  });
}
