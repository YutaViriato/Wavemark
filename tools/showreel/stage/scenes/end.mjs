// 27.45–30 s. End card: an ivory iris opens, a dot charges for a beat of
// silence, then the wordmark bursts out on the tonic chord with the sonic logo
// (A–D–F#), the underline draws, and the install line and URL settle.
import { scene, h, set, css, draw, clamp, lerp, ease, spring, springVel, splitGlyphs, cue, rng, TAU } from '../engine.mjs';
import { underline, burst, svgLayer, shock, wordsIn, pop, splitWordsSafe } from '../kit.mjs';

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#faf9f5', zIndex: 3 } });
  cam.append(root);
  const comp = h('div', { class: 'fill', style: { transformOrigin: '960px 520px' } });
  root.append(comp);

  const dot = h('div', { class: 'abs', style: { left: '946px', top: '496px', width: '28px', height: '28px', borderRadius: '50%', background: '#141413' } });
  root.append(dot);
  const svg = svgLayer(root);
  const charge = h('circle', { cx: 960, cy: 510, r: 200, fill: 'none', stroke: '#141413', 'stroke-width': 2, opacity: 0 });
  svg.append(charge);
  const rings = [
    shock(svg, { x: 960, y: 510, t0: 28.0, r0: 14, r1: 1400, w0: 70, dur: 1.0, color: '#d97757' }),
    shock(svg, { x: 960, y: 510, t0: 28.05, r0: 10, r1: 900, w0: 6, dur: 0.8, color: '#141413' }),
  ];
  const parts = h('div', { class: 'fill' });
  root.append(parts);
  const particles = burst(parts, { x: 960, y: 510, t0: 28.0, n: 46, seed: 31, colors: ['#141413', '#d97757', '#d97757', '#e3dacc', '#87867f', '#788c5d'], speed: [600, 2000], size: [18, 52], life: [0.55, 0.95], drag: 3, gravity: 300 });

  // Wordmark
  const mark = h('div', { class: 'abs', style: { left: '0', top: '300px', width: '1920px', textAlign: 'center' } });
  const word = h('span', { text: 'Wavemark', style: { font: '400 236px/1 "Newsreader", serif', letterSpacing: '-0.03em', color: '#141413', position: 'relative', display: 'inline-block' } });
  mark.append(word);
  comp.append(mark);
  const wr = word.getBoundingClientRect();
  const glyphs = splitGlyphs(word);
  const ul = underline(5, { width: 3.2 });
  ul.svg.setAttribute('width', '100%');
  ul.svg.setAttribute('height', '100%');
  const ulWrap = h('div', { class: 'abs', style: { left: `${wr.left - wr.width * 0.02}px`, top: `${wr.top + 236 * 0.9}px`, width: `${wr.width * 1.04}px`, height: `${236 * 0.24}px` } }, ul.svg);
  comp.append(ulWrap);
  const r = rng(9);
  const g = glyphs.map((el, i) => {
    const b = el._box;
    const gx = wr.left + b.x + b.w / 2, gy = wr.top + b.y + b.h / 2;
    return { el, dx: 960 - gx, dy: 510 - gy, rot: (r() - 0.5) * 70, delay: Math.abs(i - 3.5) * 0.02 };
  });

  // Glints on the second and third notes of the sonic logo
  const glints = [{ t: 28.25, x: wr.left - 30, y: wr.top + 40 }, { t: 28.5, x: wr.right + 26, y: wr.top + 28 }].map((p) => {
    const el = h('div', { class: 'abs', html: '<svg width="56" height="56" viewBox="-28 -28 56 56"><path d="M0 -26 C3 -6 6 -3 26 0 C6 3 3 6 0 26 C-3 6 -6 3 -26 0 C-6 -3 -3 -6 0 -26Z" fill="#d97757"/></svg>', style: { left: `${p.x - 28}px`, top: `${p.y - 28}px`, transformOrigin: '50% 50%' } });
    comp.append(el);
    return { ...p, el };
  });

  const tag = h('div', { class: 'abs', style: { left: '0', top: '626px', width: '1920px', textAlign: 'center', font: '400 50px/1.2 "Newsreader", serif', letterSpacing: '-0.012em', color: '#5e5d59', whiteSpace: 'nowrap' }, html: 'Markdown that becomes a site <em style="font-style:italic;color:#141413">while you write.</em>' });
  comp.append(tag);
  const tagWords = splitWordsSafe(tag);

  const pillRow = h('div', { class: 'abs', style: { left: '0', top: '732px', width: '1920px', display: 'flex', justifyContent: 'center' } });
  const pill = h('div', { style: { display: 'flex', alignItems: 'center', gap: '18px', height: '78px', padding: '0 18px 0 30px', borderRadius: '18px', background: '#fff', border: '1px solid #e8e6dc', boxShadow: '0 18px 44px -22px rgb(20 20 19 / .35)', font: '400 27px/1 "JetBrains Mono", monospace', color: '#141413', whiteSpace: 'pre', transformOrigin: '50% 50%' }, html: '<span style="color:#d97757">$</span><span>curl -fsSL https://wavemark.viriato.site/install.sh | sh</span><span style="display:grid;place-items:center;width:46px;height:46px;border-radius:12px;background:#f0eee6"><svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#5e5d59" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect width="14" height="14" x="8" y="8" rx="2"/><path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2"/></svg></span>' });
  pillRow.append(pill);
  comp.append(pillRow);
  const url = h('div', { class: 'abs', style: { left: '0', top: '862px', width: '1920px', textAlign: 'center', font: '500 34px/1 "Hanken Grotesk", sans-serif', letterSpacing: '0.005em', color: '#141413' }, html: 'wavemark.viriato.site <span style="color:#d97757">→</span>' });
  comp.append(url);

  cue(27.75, 'charge', { dur: 0.25, vel: 0.6 });
  cue(28.75, 'pop', { note: 86, pan: 0, vel: 0.45 });
  cue(29.0, 'tick', { pan: 0, vel: 0.35 });

  scene({
    t0: 27.45, t1: 30.05, root,
    update(t) {
      const R = lerp(0, 1150, ease.inCubic(clamp((t - 27.45) / 0.3)));
      css(root, 'clip-path', t < 27.76 ? `circle(${R.toFixed(1)}px at 960px 540px)` : 'none');

      const din = spring(t - 27.6, 4, 0.5);
      const ch = ease.inExpo(clamp((t - 27.75) / 0.25));
      const out = clamp((t - 28.0) / 0.08);
      set(dot, { s: (t < 27.6 ? 0 : din) * (0.9 + 0.7 * ch) * (1 + 4 * out), o: 1 - out });
      charge.setAttribute('r', lerp(300, 20, ease.inQuart(clamp((t - 27.75) / 0.25))).toFixed(1));
      charge.setAttribute('opacity', t >= 27.75 && t < 28 ? (0.1 + 0.5 * ch).toFixed(3) : '0');
      for (const rg of rings) rg.update(t);
      particles.update(t);

      g.forEach((gl) => {
        const lt = t - 28.0 - gl.delay;
        if (lt < 0) { set(gl.el, { o: 0 }); return; }
        const p = spring(lt, 1.8, 0.62);
        const v = springVel(lt, 1.8, 0.62) / (TAU * 1.8);
        set(gl.el, { x: gl.dx * (1 - p), y: gl.dy * (1 - p), r: gl.rot * (1 - p), s: 0.15 + 0.85 * p, sx: 1 + v * 0.12, sy: 1 - v * 0.12, o: clamp(lt / 0.05), blur: (1 - clamp(lt / 0.28)) * 14 });
      });
      draw(ul.path, ease.brand(clamp((t - 28.18) / 0.5)));
      glints.forEach((gp) => {
        const lt = t - gp.t;
        const p = spring(lt, 3.2, 0.45);
        set(gp.el, { s: lt < 0 ? 0 : p * (1 - ease.inQuad(clamp((lt - 0.25) / 0.3))), r: lt * 120, o: lt < 0 ? 0 : 1 });
      });
      wordsIn(tagWords, t, 28.4, { stagger: 0.03, em: 50, dy: 0.5, blur: 6 });
      pop(pill, t, 28.75, { f: 2.8, z: 0.55, from: 0.7, dy: 30, squash: 0.1 });
      set(url, { y: (1 - spring(t - 29.0, 2.4, 0.75)) * 24, o: clamp((t - 29.0) / 0.15) });
      set(comp, { s: 1 + 0.018 * ease.inOutSine(clamp((t - 28.6) / 1.4)) });
    },
  });
}
