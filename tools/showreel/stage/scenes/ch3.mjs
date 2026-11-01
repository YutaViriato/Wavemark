// 14–16 s. 03 Strict on purpose: the real starter page with a typo, the real
// diagnostic (file, line, column, hint), the extra letter pops out, and the
// real success line. Exit: the check mark opens into chapter 4.
import { scene, h, set, css, draw, clamp, lerp, ease, spring, springVel, cue, panX, TAU } from '../engine.mjs';
import { ART, headline, wordsIn, label, typeIn, icon, drawIcon, scalePath, pop } from '../kit.mjs';
import { frag } from '../engine.mjs';

const SRC = '<Card tittle="Pages" icon="file" href="./getting-started.md">';
const CW = 24; // JetBrains Mono at 40px
const CODE = { x: 120, y: 492, w: 1680, h: 112 };
const TX = CODE.x + 118; // first character
const DROP = 9; // the extra "t"
export const CHECK = { x: 120, y: 658, size: 76 };
export const checkCenter = () => ({ x: CHECK.x + CHECK.size / 2, y: CHECK.y + CHECK.size / 2 });

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#141413' } });
  cam.append(root);
  const zoomer = h('div', { class: 'fill', style: { transformOrigin: `${checkCenter().x}px ${checkCenter().y}px` } });
  root.append(zoomer);

  const lab = label('03 — Check');
  const hl = headline(['Strict on purpose.'], { color: '#faf9f5' });
  zoomer.append(h('div', { class: 'abs', style: { left: '120px', top: '214px' } }, lab, h('div', { style: { height: '30px' } }), hl.el));

  // Source line
  const code = h('div', { class: 'abs', style: { left: `${CODE.x}px`, top: `${CODE.y}px`, width: `${CODE.w}px`, height: `${CODE.h}px`, borderRadius: '20px', background: '#1f1e1d', border: '1px solid #30302e' } });
  const gutter = h('div', { class: 'abs sr-mono', text: '16', style: { left: '30px', top: '33px', font: '400 40px/1 "JetBrains Mono", monospace', color: '#5e5d59' } });
  code.append(gutter);
  zoomer.append(code);
  const colorOf = (i) => {
    if (i === 0 || SRC[i] === '>' || SRC[i] === '=') return '#808080';
    if (i >= 1 && i <= 4) return '#569cd6';
    const before = SRC.slice(0, i);
    const quotes = (before.match(/"/g) || []).length;
    if (SRC[i] === '"' || quotes % 2 === 1) return '#ce9178';
    return '#9cdcfe';
  };
  const chars = [...SRC].map((c, i) => {
    const el = h('span', { class: 'abs sr-mono', text: c, style: { left: `${TX + i * CW}px`, top: `${CODE.y + 36}px`, width: `${CW}px`, textAlign: 'center', font: '400 40px/1 "JetBrains Mono", monospace', color: colorOf(i), transformOrigin: '50% 60%' } });
    zoomer.append(el);
    return el;
  });

  // Hand-drawn ring around "tittle": the product's Circle, stretched as geometry.
  const ringSvg = frag(ART.circles[4]);
  ringSvg.removeAttribute('width'); ringSvg.removeAttribute('height');
  const ringPath = ringSvg.querySelector('path');
  const ring = { w: 6 * CW + 64, h: 92 };
  ringPath.setAttribute('d', scalePath(ringPath.getAttribute('d'), ring.w / 48, ring.h / 48));
  ringSvg.setAttribute('viewBox', `0 0 ${ring.w} ${ring.h}`);
  Object.assign(ringSvg.style, { position: 'absolute', left: `${TX + 6 * CW - 32}px`, top: `${CODE.y + 12}px`, width: `${ring.w}px`, height: `${ring.h}px`, overflow: 'visible', color: '#d97757' });
  Object.assign(ringPath.style, { fill: 'none', stroke: 'currentColor', strokeWidth: 3.4, strokeLinecap: 'round', strokeDasharray: '1 2' });
  zoomer.append(ringSvg);
  const caretMark = h('div', { class: 'abs sr-mono', text: '^', style: { left: `${TX + 6 * CW}px`, top: `${CODE.y + CODE.h + 2}px`, width: `${CW}px`, textAlign: 'center', font: '500 36px/1 "JetBrains Mono", monospace', color: '#d97757', transformOrigin: '50% 0' } });
  zoomer.append(caretMark);

  // Diagnostic, with the exact text of `wavemark check`
  const card = h('div', { class: 'abs', style: { left: `${CODE.x}px`, top: '636px', width: `${CODE.w}px`, padding: '30px 38px', borderRadius: '20px', background: '#1f1e1d', border: '1px solid #5a2c27', transformOrigin: '50% 0', boxShadow: '0 30px 70px -30px rgb(0 0 0 / .8)' } });
  const rowA = h('div', { class: 'sr-mono', style: { font: '400 27px/1.3 "JetBrains Mono", monospace', color: '#c2c0b6', whiteSpace: 'nowrap' }, html: 'index.md:16:7: <span style="color:#e0786c;font-weight:600">error</span> <span style="color:#87867f">[component.attr-unknown]</span>' });
  const rowB = h('div', { style: { font: '400 42px/1.3 "Hanken Grotesk", sans-serif', color: '#faf9f5', marginTop: '14px', whiteSpace: 'nowrap' }, html: 'unknown attribute <span style="color:#e6a68d">\'tittle\'</span> on &lt;Card&gt;' });
  const rowC = h('div', { style: { font: '400 31px/1.3 "Hanken Grotesk", sans-serif', color: '#9c9a92', marginTop: '10px', whiteSpace: 'nowrap' }, html: 'hint: Did you mean <span style="color:#a3b88a">\'title\'</span>? Valid attributes: title, href, icon, cta.' });
  card.append(rowA, rowB, rowC);
  zoomer.append(card);

  // Success line
  const check = icon('check', { size: CHECK.size, color: '#a3b88a', fill: '#3d4a31', stroke: 2.4 });
  check.el.classList.add('abs');
  Object.assign(check.el.style, { left: `${CHECK.x}px`, top: `${CHECK.y}px`, transformOrigin: '50% 50%' });
  const okText = h('div', { class: 'abs sr-mono', style: { left: `${CHECK.x + CHECK.size + 30}px`, top: `${CHECK.y + 18}px`, font: '400 38px/1 "JetBrains Mono", monospace', color: '#faf9f5', whiteSpace: 'nowrap' }, html: '2 pages · <span style="color:#a3b88a">0 error(s)</span> · 0 warning(s)' });
  zoomer.append(check.el, okText);

  cue(14.02, 'type', { n: 10, pan: 0, vel: 0.35 });
  cue(14.5, 'error', { pan: panX(TX + 9 * CW), vel: 0.9 });
  cue(14.5, 'marker', { dur: 0.35, pan: panX(TX + 9 * CW), vel: 0.45 });
  cue(15.0, 'pluck-out', { pan: panX(TX + DROP * CW), vel: 0.8 });
  cue(15.25, 'success', { pan: -0.6, vel: 0.85 });

  scene({
    t0: 13.98, t1: 16.06, root,
    update(t) {
      typeIn(lab, t, 14.0, 0.02);
      wordsIn(hl.words, t, 14.04, { stagger: 0.06 });

      const ci = spring(t - 14.06, 2.2, 0.7);
      set(code, { y: (1 - ci) * 50, o: clamp((t - 14.06) / 0.1) });
      set(gutter, { o: clamp((t - 14.1) / 0.1) });

      // the fix: the extra "t" jumps out, the rest closes the gap
      const fx = t - 15.0;
      chars.forEach((el, i) => {
        const t0 = 14.1 + i * 0.004;
        if (t < t0) { set(el, { o: 0 }); return; }
        const inP = spring(t - t0, 4, 0.6);
        let x = 0, y = (1 - inP) * 24, r = 0, s = 1, o = clamp((t - t0) / 0.04);
        if (i === DROP && fx > 0) {
          const k = (1 - Math.exp(-2.4 * fx)) / 2.4;
          x = 260 * k; y = -520 * k + 0.5 * 2600 * fx * fx; r = 260 * k; s = 1 + 0.5 * clamp(fx / 0.08);
          o = 1 - clamp((fx - 0.25) / 0.2);
          el.style.color = '#e6a68d';
        } else if (i > DROP && fx > 0) {
          x = -CW * spring(fx - 0.06 - (i - DROP) * 0.004, 3.4, 0.62);
        }
        set(el, { x, y, r, s, o });
      });

      // the error
      draw(ringPath, ease.brand(clamp((t - 14.5) / 0.38)));
      const ringOut = clamp((fx - 0.08) / 0.2);
      css(ringSvg, 'color', fx > 0 ? '#a3b88a' : '#d97757');
      set(ringSvg, { o: 1 - ringOut, s: 1 + 0.08 * ringOut });
      const cp = pop(caretMark, t, 14.52, { f: 4, z: 0.45, from: 0, dy: 20 });
      void cp;
      if (fx > 0) set(caretMark, { o: 1 - clamp(fx / 0.1) });

      const open = spring(t - 14.5, 2.6, 0.62);
      const close = ease.inBack(clamp(fx / 0.18));
      set(card, { sy: t < 14.5 ? 0 : Math.max(0.001, open * (1 - close)), o: clamp((t - 14.5) / 0.05) * (1 - clamp((fx - 0.1) / 0.08)) });
      [rowA, rowB, rowC].forEach((r, i) => { const lt = t - 14.56 - i * 0.07; set(r, { y: (1 - spring(lt, 2.8, 0.7)) * 22, o: clamp(lt / 0.1) }); });

      // success
      drawIcon(check, t, 15.2, { dur: 0.42, stagger: 0.12, fillAt: 0.25 });
      const sx = 1 + 0.25 * Math.exp(-Math.max(0, t - 15.5) * 7) * (t > 15.5 ? 1 : 0);
      const pre = ease.inQuad(clamp((t - 15.62) / 0.13));
      set(check.el, { s: (t < 15.2 ? 0.6 : 1) * sx * (1 + 0.3 * pre), o: t >= 15.2 ? 1 : 0 });
      set(okText, { x: (1 - ease.outExpo(clamp((t - 15.25) / 0.4))) * -40, o: clamp((t - 15.25) / 0.08) });

      // exit: dive toward the check while chapter 4 opens from it
      const dive = ease.inCubic(clamp((t - 15.76) / 0.3));
      set(zoomer, { s: 1 + 0.6 * dive, o: 1 - clamp((t - 15.9) / 0.12) });
    },
  });
}
