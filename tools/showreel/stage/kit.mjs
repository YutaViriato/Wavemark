// Shared building blocks for the scenes: product artwork, type, code windows,
// particles and transitions. Everything renders from time t.
import { h, frag, set, draw, clamp, lerp, ease, tw, spring, springVel, rng, splitWords, css, TAU } from './engine.mjs';

export const ART = { icons: {}, underlines: [], arrows: [], circles: [] };
export const FRAG = {};
export function loadAssets(art, frags) {
  Object.assign(ART, art);
  Object.assign(FRAG, frags);
}

// ---------------------------------------------------------------- artwork
// A product icon from internal/sketch. Returns the element, its ink paths
// and its fill group (the oat print, offset out of register).
export function icon(name, { size = 48, color = 'currentColor', fill, stroke } = {}) {
  const svg = frag(ART.icons[name]);
  svg.removeAttribute('width');
  svg.removeAttribute('height');
  const el = h('div', { class: 'sr-icon', style: { width: `${size}px`, height: `${size}px`, color } }, svg);
  if (fill) el.style.setProperty('--wm-sketch-fill', fill);
  if (stroke) el.style.setProperty('--wm-sketch-stroke', stroke);
  const ink = [...svg.querySelectorAll('.wm-sketch__ink path')];
  const fillG = svg.querySelector('.wm-sketch__fill');
  return { el, svg, ink, fill: fillG, name };
}
// Animates a product icon: the pen draws each stroke in turn, then the fill
// slides in from further out of register and settles at the product offset.
export function drawIcon(ic, t, t0, { dur = 0.55, stagger = 0.06, fillAt = 0.3, e = ease.brand } = {}) {
  ic.ink.forEach((p, i) => draw(p, e(clamp((t - t0 - i * stagger) / dur))));
  if (ic.fill) {
    const f = spring(t - t0 - fillAt, 3.2, 0.62);
    const o = clamp((t - t0 - fillAt) / 0.12);
    ic.fill.setAttribute('transform', `translate(${(1.3 + (1 - f) * 5).toFixed(2)} ${(1.1 + (1 - f) * 4).toFixed(2)})`);
    set(ic.fill, { o });
  }
}
export function underline(i = 0, { color = 'var(--clay)', width = 3 } = {}) {
  const svg = frag(ART.underlines[i % ART.underlines.length]);
  svg.style.color = color;
  svg.style.overflow = 'visible';
  const path = svg.querySelector('path');
  path.style.strokeWidth = width;
  path.style.fill = 'none';
  path.style.stroke = 'currentColor';
  path.style.strokeLinecap = 'round';
  path.style.strokeDasharray = '1 2';
  return { svg, path };
}
export function ring(i = 0, { color = 'currentColor', width = 2 } = {}) {
  const svg = frag(ART.circles[i % ART.circles.length]);
  svg.removeAttribute('width');
  svg.removeAttribute('height');
  svg.style.overflow = 'visible';
  svg.style.color = color;
  const path = svg.querySelector('path');
  Object.assign(path.style, { fill: 'none', stroke: 'currentColor', strokeWidth: width, strokeLinecap: 'round', strokeDasharray: '1 2' });
  return { svg, path };
}
export function arrow(i = 0, { color = 'var(--clay)', width = 2.2 } = {}) {
  const svg = frag(ART.arrows[i % ART.arrows.length]);
  svg.style.overflow = 'visible';
  svg.style.color = color;
  const paths = [...svg.querySelectorAll('path')];
  for (const p of paths) Object.assign(p.style, { fill: 'none', stroke: 'currentColor', strokeWidth: width, strokeLinecap: 'round', strokeLinejoin: 'round', strokeDasharray: '1 2' });
  return { svg, paths };
}

// ---------------------------------------------------------------- type
// A display headline split into words. lines: array of strings (may contain <em>).
export function headline(lines, { cls = 'sr-display', color, size, align = 'left' } = {}) {
  const el = h('div', { class: cls, style: { color, textAlign: align, ...(size ? { fontSize: `${size}px` } : {}) } });
  el.innerHTML = lines.join('<br>');
  const words = splitWords(el);
  return { el, words };
}
// The product's word-rise entrance (translateY .42em, blur 4px → 0), scaled up
// for the screen and given a little overshoot.
export function wordsIn(words, t, t0, { stagger = 0.05, dur = 0.7, dy = 0.42, blur = 10, em = 128, overshoot = true, order } = {}) {
  words.forEach((w, i) => {
    const k = order ? order[i] : i;
    const lt = t - t0 - k * stagger;
    const p = overshoot ? spring(lt, 2.4, 0.72) : ease.brandExpo(clamp(lt / dur));
    const o = clamp(lt / (dur * 0.35));
    const b = (1 - ease.outCubic(clamp(lt / (dur * 0.6)))) * blur;
    set(w, { y: (1 - p) * dy * em, o, blur: b });
  });
}
export function wordsOut(words, t, t0, { stagger = 0.025, dur = 0.3, dy = -0.3, em = 128, blur = 10, dx = 0 } = {}) {
  words.forEach((w, i) => {
    const lt = t - t0 - i * stagger;
    if (lt <= 0) return;
    const p = ease.inCubic(clamp(lt / dur));
    set(w, { x: p * dx, y: p * dy * em, o: 1 - p, blur: p * blur });
  });
}
export function label(text, { color = 'var(--clay)' } = {}) {
  return h('div', { class: 'sr-label', style: { color }, text });
}
// Reveals a label letter by letter (typewriter, on a fixed step).
export function typeIn(el, t, t0, step = 0.018) {
  if (!el._chars) {
    const text = el.textContent;
    el.textContent = '';
    el._chars = [...text].map((c) => { const s = h('span', { text: c }); el.append(s); return s; });
  }
  el._chars.forEach((c, i) => css(c, 'visibility', t >= t0 + i * step ? 'visible' : 'hidden'));
}

// ---------------------------------------------------------------- syntax
const esc = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
// A tiny highlighter for the Markdown + component source shown on screen.
export function highlight(src) {
  if (/^\$ /.test(src)) return `<span class="tk-prompt">$</span> ${esc(src.slice(2))}`;
  if (/^---\s*$/.test(src)) return `<span class="tk-pun">---</span>`;
  if (/^[a-z]+: /.test(src)) { const i = src.indexOf(':'); return `<span class="tk-key">${esc(src.slice(0, i))}</span><span class="tk-pun">:</span>${esc(src.slice(i + 1))}`; }
  if (/^#{1,3} /.test(src)) { const m = src.match(/^(#{1,3}) (.*)$/); return `<span class="tk-hm">${m[1]}</span> <span class="tk-h">${esc(m[2])}</span>`; }
  let out = '';
  let i = 0;
  while (i < src.length) {
    const rest = src.slice(i);
    let m;
    if ((m = rest.match(/^<\/?([A-Z][A-Za-z]*)/))) { out += `<span class="tk-pun">${esc(m[0].slice(0, m[0].length - m[1].length))}</span><span class="tk-tag">${m[1]}</span>`; i += m[0].length; continue; }
    if ((m = rest.match(/^\/?>/))) { out += `<span class="tk-pun">${esc(m[0])}</span>`; i += m[0].length; continue; }
    if ((m = rest.match(/^([a-z]+)(=)/)) && /[\s<]$|^$/.test(src.slice(0, i).slice(-1) || ' ')) { out += `<span class="tk-attr">${m[1]}</span><span class="tk-pun">=</span>`; i += m[0].length; continue; }
    if ((m = rest.match(/^"[^"]*"/))) { out += `<span class="tk-str">${esc(m[0])}</span>`; i += m[0].length; continue; }
    if ((m = rest.match(/^\{[^}]*\}/))) { out += `<span class="tk-code">${esc(m[0])}</span>`; i += m[0].length; continue; }
    if ((m = rest.match(/^`[^`]*`/))) { out += `<span class="tk-code">${esc(m[0])}</span>`; i += m[0].length; continue; }
    out += esc(src[i]);
    i++;
  }
  return out;
}

// A dark code window (editor or terminal) with numbered lines.
export function codeWindow({ x, y, w, h: hh, tab = 'index.md', lines = [], numbers = true, font = 17, lh = 28, dot = true }) {
  const el = h('div', { class: 'sr-code', style: { left: `${x}px`, top: `${y}px`, width: `${w}px`, height: `${hh}px` } });
  const bar = h('div', { class: 'sr-code__bar' });
  const tabEl = h('div', { class: 'sr-code__tab' });
  const tabDot = h('span', { style: { width: '8px', height: '8px', borderRadius: '50%', background: 'var(--clay)', display: dot ? 'block' : 'none' } });
  tabEl.append(h('span', { text: tab }), tabDot);
  bar.append(h('span', { class: 'sr-dot', style: { background: '#3d3d3a' } }), h('span', { class: 'sr-dot', style: { background: '#3d3d3a' } }), h('span', { class: 'sr-dot', style: { background: '#3d3d3a', marginRight: '10px' } }), tabEl);
  const body = h('div', { class: 'sr-code__body', style: { font: `400 ${font}px/${lh}px "JetBrains Mono", monospace` } });
  const rows = lines.map((src, i) => {
    const row = h('div', { class: 'sr-line', style: { height: `${lh}px` } });
    const hl = h('div', { class: 'sr-hl', style: { opacity: 0 } });
    const ln = h('span', { class: 'sr-ln', text: numbers ? String(i + 1) : '', style: numbers ? {} : { width: '28px' } });
    const code = h('span', { class: 'sr-src', html: highlight(src) });
    row.append(hl, ln, code);
    body.append(row);
    return { row, hl, ln, code, src, chars: src.length };
  });
  el.append(bar, body);
  return { el, bar, tab: tabEl, tabDot, body, rows, font, cw: font * 0.6, lh };
}
// Reveals a code row character by character between t0 and t0 + n * step.
export function typeRow(r, t, t0, step = 0.012, cw = 10.2) {
  const n = r.chars;
  const shown = t < t0 ? 0 : Math.min(n, Math.floor((t - t0) / step) + 1);
  const o = t < t0 ? 0 : 1;
  set(r.row, { o });
  css(r.code, 'clip-path', shown >= n ? 'none' : `inset(-4px calc(100% - ${shown * cw}px) -4px -2px)`);
  return shown;
}

// ---------------------------------------------------------------- pointer
export function pointer() {
  const el = h('div', { class: 'sr-pointer' });
  el.innerHTML = '<svg viewBox="0 0 34 44" width="34" height="44"><path d="M3 2 L3 34 L11 26.5 L16.5 39 L22 36.6 L16.6 24.3 L27.5 24 Z" fill="#141413" stroke="#faf9f5" stroke-width="2.2" stroke-linejoin="round"/></svg>';
  return el;
}

// ---------------------------------------------------------------- particles
// Glyph particles with drag, gravity and spin. Positions are analytic.
export function burst(parent, { x, y, t0, n = 60, seed = 7, chars = '#*>-`[]_~|{}</!', colors = ['#faf9f5'], speed = [500, 1600], size = [18, 54], life = [0.7, 1.4], drag = 3.2, gravity = 260, spread = TAU, angle = 0, font = '"JetBrains Mono", monospace' }) {
  const r = rng(seed);
  const ps = [];
  for (let i = 0; i < n; i++) {
    const a = angle + (r() - 0.5) * spread;
    const v = lerp(speed[0], speed[1], Math.pow(r(), 0.7));
    const el = h('div', { class: 'abs', text: chars[Math.floor(r() * chars.length)], style: { font: `500 ${Math.round(lerp(size[0], size[1], r()))}px/1 ${font}`, color: colors[Math.floor(r() * colors.length)], left: '0', top: '0' } });
    parent.append(el);
    ps.push({ el, a, v, life: lerp(life[0], life[1], r()), spin: (r() - 0.5) * 720, rot: (r() - 0.5) * 60, delay: r() * 0.04, depth: r() });
  }
  return {
    update(t) {
      for (const p of ps) {
        const d = t - t0 - p.delay;
        if (d < 0 || d > p.life) { set(p.el, { o: 0 }); continue; }
        const k = (1 - Math.exp(-drag * d)) / drag;
        const px = x + Math.cos(p.a) * p.v * k;
        const py = y + Math.sin(p.a) * p.v * k + 0.5 * gravity * d * d;
        const fade = 1 - ease.inQuad(clamp(d / p.life));
        const sc = (0.6 + 0.6 * p.depth) * (1 - 0.3 * clamp(d / p.life));
        set(p.el, { x: px, y: py, r: p.rot + p.spin * k, s: sc, o: fade * clamp(d / 0.03), blur: p.depth < 0.25 ? 3 : 0 });
      }
    },
  };
}

// ---------------------------------------------------------------- shapes
export function svgLayer(parent, { w = 1920, hh = 1080 } = {}) {
  const svg = h('svg', { width: w, height: hh, viewBox: `0 0 ${w} ${hh}`, class: 'abs', style: { overflow: 'visible' } });
  parent.append(svg);
  return svg;
}
// Expanding ring (shockwave).
export function shock(svg, { x, y, t0, r0 = 8, r1 = 1200, w0 = 40, dur = 0.9, color = '#d97757', e = ease.outExpo }) {
  const c = h('circle', { cx: x, cy: y, r: r0, fill: 'none', stroke: color });
  svg.append(c);
  return {
    update(t) {
      const p = clamp((t - t0) / dur);
      if (t < t0 || p >= 1) { c.setAttribute('visibility', 'hidden'); return; }
      c.setAttribute('visibility', 'visible');
      const q = e(p);
      c.setAttribute('r', (r0 + (r1 - r0) * q).toFixed(1));
      c.setAttribute('stroke-width', Math.max(0.1, w0 * (1 - q)).toFixed(2));
      c.setAttribute('opacity', (1 - ease.inQuad(p)).toFixed(3));
    },
  };
}
// Points along a vertical edge at x that waves like a hand-drawn line.
export function wavyEdge(x, { amp = 26, wl = 360, phase = 0, top = -60, bottom = 1140, n = 36, lean = 0 } = {}) {
  const pts = [];
  for (let i = 0; i <= n; i++) {
    const y = lerp(top, bottom, i / n);
    pts.push([x + Math.sin((y / wl) * TAU + phase) * amp + (y - 540) * lean, y]);
  }
  return pts;
}
// clip-path polygon for everything left of a wavy edge.
export function clipLeftOf(x, opts) {
  const e = wavyEdge(x, opts);
  return `polygon(-200px -200px, ${e.map(([px, py]) => `${px.toFixed(1)}px ${py.toFixed(1)}px`).join(', ')}, -200px 1300px)`;
}
// SVG path for the band between two wavy edges.
export function bandPath(xl, xr, opts) {
  const a = wavyEdge(xl, opts);
  const b = wavyEdge(xr, { ...opts, phase: (opts?.phase ?? 0) + 0.9 }).reverse();
  return `M${a.map(([x, y]) => `${x.toFixed(1)} ${y.toFixed(1)}`).join(' L')} L${b.map(([x, y]) => `${x.toFixed(1)} ${y.toFixed(1)}`).join(' L')} Z`;
}

// Scales the absolute coordinates of a path made of M/L/Q commands, so a
// hand-drawn ring can become an ellipse without distorting the stroke.
export function scalePath(d, sx, sy, tx = 0, ty = 0) {
  let i = 0;
  return d.replace(/-?\d*\.?\d+(?:e-?\d+)?/g, (n) => {
    const v = parseFloat(n);
    const out = i % 2 === 0 ? v * sx + tx : v * sy + ty;
    i++;
    return out.toFixed(2);
  });
}

// Splits words in an element that may contain inline markup (em, strong).
export const splitWordsSafe = (el) => splitWords(el);

// Pops an element in with a springy scale and a little squash and stretch.
export function pop(el, t, t0, { f = 3.4, z = 0.48, from = 0, rot = 0, squash = 0.18, dy = 0, o = true, extra = {} } = {}) {
  const lt = t - t0;
  if (lt < 0) { set(el, { o: 0, s: from, ...extra }); return 0; }
  const p = spring(lt, f, z);
  const v = springVel(lt, f, z) / (TAU * f);
  const s = from + (1 - from) * p;
  set(el, { s, sx: 1 + v * squash, sy: 1 - v * squash, r: rot * (1 - p), y: dy * (1 - p), o: o ? clamp(lt / 0.06) : 1, ...extra });
  return p;
}
