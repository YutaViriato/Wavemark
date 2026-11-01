// 7.5–10.3 s. The page collapses into an iris, the screen closes to a clay
// dot, a beat of silence, then the drop: shockwave, Markdown glyphs, and the
// wordmark bursting out of the dot with the product's hand-drawn underline.
import { scene, h, set, css, draw, clamp, lerp, ease, spring, springVel, splitGlyphs, cue, rng, TAU } from '../engine.mjs';
import { underline, burst, svgLayer, shock, bandPath, wordsIn } from '../kit.mjs';
import { WAVE } from '../trans.mjs';

export async function build({ cam, fx }) {
  const root = h('div', { class: 'sc', style: { background: 'transparent' } });
  cam.append(root);

  // Iris: a hole in a dark field that closes on the collapsing page.
  const iris = h('div', { class: 'fill' });
  const bg = h('div', { class: 'fill', style: { background: '#141413' } });
  root.append(iris, bg);
  const dot = h('div', { class: 'abs', style: { left: '944px', top: '524px', width: '32px', height: '32px', borderRadius: '50%', background: '#d97757' } });
  root.append(dot);

  const svg = svgLayer(root);
  const charge = h('circle', { cx: 960, cy: 540, r: 200, fill: 'none', stroke: '#faf9f5', 'stroke-width': 2 });
  svg.append(charge);
  const rings = [
    shock(svg, { x: 960, y: 540, t0: 8.0, r0: 16, r1: 1500, w0: 90, dur: 1.0, color: '#d97757' }),
    shock(svg, { x: 960, y: 540, t0: 8.04, r0: 10, r1: 1050, w0: 10, dur: 0.85, color: '#faf9f5' }),
    shock(svg, { x: 960, y: 540, t0: 8.1, r0: 10, r1: 700, w0: 4, dur: 0.75, color: '#e3dacc' }),
  ];
  const parts = h('div', { class: 'fill' });
  root.append(parts);
  const particles = burst(parts, { x: 960, y: 540, t0: 8.0, n: 90, seed: 11, colors: ['#faf9f5', '#faf9f5', '#d97757', '#e3dacc', '#87867f', '#6a9bcc', '#788c5d'], speed: [700, 2600], size: [18, 64], life: [0.8, 1.6], drag: 2.6, gravity: 380 });

  // Wordmark: the product's logo text, in the product's serif, kerning kept.
  const mark = h('div', { class: 'abs', style: { left: '0', top: '0', width: '1920px', textAlign: 'center' } });
  const word = h('span', { text: 'Wavemark', style: { font: '400 270px/1 "Newsreader", serif', letterSpacing: '-0.03em', color: '#faf9f5', position: 'relative', display: 'inline-block' } });
  mark.append(word);
  root.append(mark);
  css(mark, 'top', '330px');
  const wr = word.getBoundingClientRect();
  const glyphs = splitGlyphs(word);
  const ul = underline(2, { width: 3.2 });
  const ulWrap = h('div', { class: 'abs', style: { left: `${wr.left - wr.width * 0.02}px`, top: `${wr.top + 270 * 0.9}px`, width: `${wr.width * 1.04}px`, height: `${270 * 0.24}px` } }, ul.svg);
  ul.svg.setAttribute('width', '100%');
  ul.svg.setAttribute('height', '100%');
  root.append(ulWrap);
  const cx = 960, cy = 540;
  const r = rng(5);
  const g = glyphs.map((el, i) => {
    const b = el._box;
    const gx = wr.left + b.x + b.w / 2, gy = wr.top + b.y + b.h / 2;
    return { el, dx: cx - gx, dy: cy - gy, rot: (r() - 0.5) * 70, delay: Math.abs(i - 3.5) * 0.022, push: 0.02 * i };
  });

  const tag = h('div', { class: 'abs', style: { left: '0', top: '700px', width: '1920px', textAlign: 'center', font: '400 64px/1 "Newsreader", serif', letterSpacing: '-0.015em' } });
  const tagA = h('span', { text: 'Markdown in.', style: { color: '#9c9a92', display: 'inline-block' } });
  const tagB = h('span', { text: 'Website out.', style: { color: '#faf9f5', display: 'inline-block', fontStyle: 'italic' } });
  tag.append(tagA, document.createTextNode('\u2002'), tagB);
  root.append(tag);

  // The wave wipe lives above every scene.
  const wipe = h('svg', { class: 'abs', width: 1920, height: 1080, viewBox: '0 0 1920 1080', style: { overflow: 'visible' } });
  const bandOat = h('path', { fill: '#e3dacc' });
  const bandClay = h('path', { fill: '#d97757' });
  wipe.append(bandOat, bandClay);
  fx.append(wipe);

  cue(7.5, 'whoosh', { dur: 0.25, vel: 0.5, tone: 1.6, rev: true });
  cue(7.75, 'charge', { dur: 0.25, vel: 0.7 });
  cue(8.75, 'whoosh', { dur: 0.35, vel: 0.22, tone: 1.3, pan: -0.2 });
  cue(9.0, 'whoosh', { dur: 0.35, vel: 0.22, tone: 1.5, pan: 0.2 });
  cue(9.68, 'whoosh', { dur: 0.6, vel: 0.85, tone: 0.9, pan: -0.8, to: 0.8 });

  scene({
    t0: 7.5, t1: 10.34, root,
    update(t) {
      // iris closes on the collapsing page
      const ir = t < 7.75 ? lerp(1150, 0, ease.inCubic(clamp((t - 7.5) / 0.25))) : 0;
      css(iris, 'background', `radial-gradient(circle at 960px 540px, transparent ${ir.toFixed(1)}px, #141413 ${(ir + 1.5).toFixed(1)}px)`);
      set(iris, { o: t < 7.76 ? 1 : 0 });
      set(bg, { o: t >= 7.75 ? 1 : 0 });
      const dotIn = spring(t - 7.7, 4, 0.5);
      const ch = ease.inExpo(clamp((t - 7.75) / 0.25));
      const out = clamp((t - 8.0) / 0.08);
      set(dot, { s: (t < 7.7 ? 0 : dotIn) * (0.9 + 0.6 * ch) * (1 + 4 * out), o: 1 - out });
      charge.setAttribute('r', lerp(300, 22, ease.inQuart(clamp((t - 7.75) / 0.25))).toFixed(1));
      charge.setAttribute('opacity', t >= 7.75 && t < 8 ? (0.15 + 0.55 * ch).toFixed(3) : '0');

      for (const rg of rings) rg.update(t);
      particles.update(t);

      // wordmark bursts out of the dot, then is pushed off by the wipe
      g.forEach((gl) => {
        const lt = t - 8.0 - gl.delay;
        if (lt < 0) { set(gl.el, { o: 0 }); return; }
        const p = spring(lt, 1.85, 0.6);
        const v = springVel(lt, 1.85, 0.6) / (TAU * 1.85);
        const blow = ease.inCubic(clamp((t - 9.74 - gl.push) / 0.3));
        set(gl.el, {
          x: gl.dx * (1 - p) + blow * 1400, y: gl.dy * (1 - p) - blow * 60,
          r: gl.rot * (1 - p) + blow * 18, s: (0.15 + 0.85 * p) * (1 + blow * 0.3),
          sx: 1 + v * 0.12, sy: 1 - v * 0.12,
          o: clamp(lt / 0.05) * (1 - blow), blur: (1 - clamp(lt / 0.28)) * 14 + blow * 18,
        });
      });
      draw(ul.path, ease.brand(clamp((t - 8.22) / 0.5)));
      set(ulWrap, { x: ease.inCubic(clamp((t - 9.72) / 0.3)) * 1500 });

      wordsIn([tagA], t, 8.75, { em: 64, dy: 0.5, blur: 8 });
      wordsIn([tagB], t, 9.0, { em: 64, dy: 0.5, blur: 8 });
      const tagOut = ease.inCubic(clamp((t - 9.76) / 0.3));
      set(tag, { x: tagOut * 1300, o: 1 - tagOut });
    },
  });

  // The wave wipe sits above every scene, so it is driven for the whole reel.
  scene({
    t0: 0, t1: 31,
    update(t) {
      const ph = WAVE.phase(t);
      const L2 = WAVE.lead(t, 0.05), L = WAVE.lead(t), T = WAVE.trail(t);
      const on = t >= WAVE.t0 - 0.05 && T < 2300;
      css(wipe, 'display', on ? '' : 'none');
      if (on) {
        bandOat.setAttribute('d', bandPath(Math.max(T, -330), Math.max(L2, -320), { phase: ph, amp: 30, wl: 420, lean: -0.05 }));
        bandClay.setAttribute('d', bandPath(Math.max(T, -330), Math.max(L, -320), { phase: ph + 0.6, amp: 30, wl: 420, lean: -0.05 }));
      }
    },
  });
}
