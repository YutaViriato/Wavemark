// Motion engine for the showreel. Everything is a pure function of time:
// the renderer seeks to t, every scene sets its styles for t, and the frame is
// captured. No CSS transitions, no requestAnimationFrame, no wall clock.

export const W = 1920;
export const H = 1080;
export const BPM = 120;
export const BEAT = 60 / BPM; // 0.5 s
export const BAR = 4 * BEAT; // 2 s
export const DURATION = 30;
export const beat = (n) => n * BEAT;

// ---------------------------------------------------------------- math
export const clamp = (x, a = 0, b = 1) => (x < a ? a : x > b ? b : x);
export const lerp = (a, b, t) => a + (b - a) * t;
export const invlerp = (a, b, x) => clamp((x - a) / (b - a));
export const mix = (a, b, t) => (Array.isArray(a) ? a.map((v, i) => lerp(v, b[i], t)) : lerp(a, b, t));
export const smooth = (x) => x * x * (3 - 2 * x);
export const TAU = Math.PI * 2;

// ---------------------------------------------------------------- easing
function bezier(x1, y1, x2, y2) {
  // Same solver as browsers' cubic-bezier(): Newton steps, bisection fallback.
  const cx = 3 * x1, bx = 3 * (x2 - x1) - cx, ax = 1 - cx - bx;
  const cy = 3 * y1, by = 3 * (y2 - y1) - cy, ay = 1 - cy - by;
  const sx = (t) => ((ax * t + bx) * t + cx) * t;
  const sy = (t) => ((ay * t + by) * t + cy) * t;
  const dx = (t) => (3 * ax * t + 2 * bx) * t + cx;
  return (x) => {
    if (x <= 0) return 0;
    if (x >= 1) return 1;
    let t = x;
    for (let i = 0; i < 8; i++) {
      const e = sx(t) - x;
      if (Math.abs(e) < 1e-6) return sy(t);
      const d = dx(t);
      if (Math.abs(d) < 1e-6) break;
      t -= e / d;
    }
    let lo = 0, hi = 1;
    t = x;
    while (hi - lo > 1e-6) {
      const v = sx(t);
      if (v < x) lo = t; else hi = t;
      t = (lo + hi) / 2;
    }
    return sy(t);
  };
}
const back = (s) => (x) => 1 + (s + 1) * Math.pow(x - 1, 3) + s * Math.pow(x - 1, 2);
export const ease = {
  linear: (x) => x,
  inQuad: (x) => x * x,
  outQuad: (x) => 1 - (1 - x) * (1 - x),
  inOutQuad: (x) => (x < 0.5 ? 2 * x * x : 1 - Math.pow(-2 * x + 2, 2) / 2),
  inCubic: (x) => x * x * x,
  outCubic: (x) => 1 - Math.pow(1 - x, 3),
  inOutCubic: (x) => (x < 0.5 ? 4 * x * x * x : 1 - Math.pow(-2 * x + 2, 3) / 2),
  inQuart: (x) => x * x * x * x,
  outQuart: (x) => 1 - Math.pow(1 - x, 4),
  inOutQuart: (x) => (x < 0.5 ? 8 * x ** 4 : 1 - Math.pow(-2 * x + 2, 4) / 2),
  inQuint: (x) => x ** 5,
  outQuint: (x) => 1 - Math.pow(1 - x, 5),
  inOutQuint: (x) => (x < 0.5 ? 16 * x ** 5 : 1 - Math.pow(-2 * x + 2, 5) / 2),
  inExpo: (x) => (x <= 0 ? 0 : Math.pow(2, 10 * x - 10)),
  outExpo: (x) => (x >= 1 ? 1 : 1 - Math.pow(2, -10 * x)),
  inOutExpo: (x) => (x <= 0 ? 0 : x >= 1 ? 1 : x < 0.5 ? Math.pow(2, 20 * x - 10) / 2 : (2 - Math.pow(2, -20 * x + 10)) / 2),
  inSine: (x) => 1 - Math.cos((x * Math.PI) / 2),
  outSine: (x) => Math.sin((x * Math.PI) / 2),
  inOutSine: (x) => -(Math.cos(Math.PI * x) - 1) / 2,
  outCirc: (x) => Math.sqrt(1 - Math.pow(x - 1, 2)),
  inCirc: (x) => 1 - Math.sqrt(1 - x * x),
  outBack: back(1.7),
  outBackSoft: back(1.1),
  outBackHard: back(2.6),
  inBack: (x) => 2.7 * x * x * x - 1.7 * x * x,
  // Wavemark's own curves (internal/theme/assets/wavemark.css)
  brand: bezier(0.165, 0.84, 0.44, 1), // out-quart, anthropic.com
  brandExpo: bezier(0.19, 1, 0.22, 1), // out-expo, claude.com
  snap: bezier(0.2, 0, 0, 1), // emphasized decelerate
  whip: bezier(0.75, 0, 0.15, 1), // fast through the middle, for pans
  bezier,
};

// ---------------------------------------------------------------- springs
// Closed-form damped harmonic oscillator, unit step from 0 to 1.
// f: natural frequency in Hz, z: damping ratio (<1 overshoots).
export function spring(t, f = 2.5, z = 0.5) {
  if (t <= 0) return 0;
  const w = TAU * f;
  if (z < 1) {
    const wd = w * Math.sqrt(1 - z * z);
    return 1 - Math.exp(-z * w * t) * (Math.cos(wd * t) + ((z * w) / wd) * Math.sin(wd * t));
  }
  return 1 - Math.exp(-w * t) * (1 + w * t);
}
// Velocity of the same spring (units per second), for squash and stretch.
export function springVel(t, f = 2.5, z = 0.5) {
  if (t <= 0) return 0;
  const w = TAU * f;
  if (z < 1) {
    const wd = w * Math.sqrt(1 - z * z);
    return Math.exp(-z * w * t) * ((w * w) / wd) * Math.sin(wd * t);
  }
  return w * w * t * Math.exp(-w * t);
}

// ---------------------------------------------------------------- tweens
// Eased progress of a tween that starts at t0 and lasts dur.
export const tw = (t, t0, dur, e = ease.outExpo) => e(clamp((t - t0) / dur));
// Keyframes: [[t, value, easeIntoThisKey?], ...]. Values may be arrays.
export function kf(t, keys) {
  if (t <= keys[0][0]) return keys[0][1];
  for (let i = 1; i < keys.length; i++) {
    const [t1, v1, e = ease.inOutCubic] = keys[i];
    if (t < t1) {
      const [t0, v0] = keys[i - 1];
      return mix(v0, v1, e((t - t0) / (t1 - t0)));
    }
  }
  return keys[keys.length - 1][1];
}
// Sum of decaying pulses after each time (beat-reactive accents).
export function pulses(t, times, decay = 0.18, attack = 0.012) {
  let v = 0;
  for (const p of times) {
    const d = t - p;
    if (d < -attack || d > decay * 8) continue;
    v += d < 0 ? 1 + d / attack : Math.exp(-d / decay);
  }
  return v;
}

// ---------------------------------------------------------------- randomness
export function rng(seed = 1) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6d2b79f5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}
// Smooth 1-D value noise in [-1, 1], for drift and handheld camera.
export function noise(x, seed = 0) {
  const hash = (n) => {
    const s = Math.sin(n * 127.1 + seed * 311.7) * 43758.5453;
    return (s - Math.floor(s)) * 2 - 1;
  };
  const i = Math.floor(x), f = x - i;
  return lerp(hash(i), hash(i + 1), smooth(f));
}

// ---------------------------------------------------------------- DOM
export function h(tag, props = {}, ...kids) {
  const svgTags = /^(svg|path|g|circle|rect|line|polyline|polygon|ellipse|defs|clipPath|mask|linearGradient|radialGradient|stop|text|use)$/;
  const el = svgTags.test(tag) ? document.createElementNS('http://www.w3.org/2000/svg', tag) : document.createElement(tag);
  for (const [k, v] of Object.entries(props)) {
    if (v == null || v === false) continue;
    if (k === 'class') el.setAttribute('class', v);
    else if (k === 'style' && typeof v === 'object') Object.assign(el.style, v);
    else if (k === 'text') el.textContent = v;
    else if (k === 'html') el.innerHTML = v;
    else el.setAttribute(k, v === true ? '' : v);
  }
  for (const kid of kids.flat()) if (kid != null) el.append(kid);
  return el;
}
export function frag(html) {
  const t = document.createElement('template');
  t.innerHTML = html.trim();
  return t.content.firstElementChild;
}

const r2 = (v) => Math.round(v * 100) / 100;
// Sets transform, opacity and filter from a small property bag, skipping
// writes that would not change anything.
export function set(el, p) {
  const st = el.style;
  if (p.x !== undefined || p.y !== undefined || p.s !== undefined || p.sx !== undefined || p.sy !== undefined || p.r !== undefined || p.rx !== undefined || p.ry !== undefined || p.z !== undefined || p.skx !== undefined) {
    const s = p.s ?? 1;
    let tr = '';
    if (p.persp) tr += `perspective(${p.persp}px) `;
    tr += `translate(${r2(p.x ?? 0)}px,${r2(p.y ?? 0)}px)`;
    if (p.z) tr += ` translateZ(${r2(p.z)}px)`;
    if (p.rx) tr += ` rotateX(${r2(p.rx)}deg)`;
    if (p.ry) tr += ` rotateY(${r2(p.ry)}deg)`;
    if (p.r) tr += ` rotate(${r2(p.r)}deg)`;
    if (p.skx) tr += ` skewX(${r2(p.skx)}deg)`;
    const sx = s * (p.sx ?? 1), sy = s * (p.sy ?? 1);
    if (sx !== 1 || sy !== 1) tr += ` scale(${Math.round(sx * 10000) / 10000},${Math.round(sy * 10000) / 10000})`;
    if (el._tr !== tr) { st.transform = tr; el._tr = tr; }
  }
  if (p.o !== undefined) {
    const o = Math.round(clamp(p.o) * 1000) / 1000;
    if (el._o !== o) { st.opacity = o; el._o = o; }
    const vis = o <= 0 ? 'hidden' : '';
    if (el._vis !== vis) { st.visibility = vis; el._vis = vis; }
  }
  if (p.blur !== undefined || p.bright !== undefined) {
    let f = '';
    if (p.blur > 0.05) f += `blur(${r2(p.blur)}px) `;
    if (p.bright !== undefined && Math.abs(p.bright - 1) > 0.001) f += `brightness(${r2(p.bright)})`;
    f = f.trim() || 'none';
    if (el._f !== f) { st.filter = f; el._f = f; }
  }
  if (p.clip !== undefined && el._clip !== p.clip) { st.clipPath = p.clip; el._clip = p.clip; }
  if (p.color !== undefined && el._color !== p.color) { st.color = p.color; el._color = p.color; }
  if (p.bg !== undefined && el._bg !== p.bg) { st.background = p.bg; el._bg = p.bg; }
}
export function css(el, prop, value) {
  const key = '_c_' + prop;
  if (el[key] !== value) { el.style.setProperty(prop, value); el[key] = value; }
}
export const show = (el, on) => css(el, 'display', on ? '' : 'none');

// Draw-on for hand-drawn strokes (paths carry pathLength="1").
export function draw(path, p) {
  const v = r2(1 - clamp(p));
  if (path._d !== v) { path.style.strokeDashoffset = v; path._d = v; }
}

// Splits a text element into words (inline-block spans), keeping spaces.
export function splitWords(el) {
  const words = [];
  const walk = (node) => {
    for (const child of [...node.childNodes]) {
      if (child.nodeType === 3) {
        const parts = child.textContent.split(/(\s+)/);
        const f = document.createDocumentFragment();
        for (const part of parts) {
          if (!part) continue;
          if (/^\s+$/.test(part)) f.append(document.createTextNode(part));
          else {
            const w = h('span', { class: 'sr-word', text: part });
            words.push(w);
            f.append(w);
          }
        }
        child.replaceWith(f);
      } else if (child.nodeType === 1 && !child.classList.contains('sr-word')) walk(child);
    }
  };
  walk(el);
  return words;
}

// Splits text into absolutely positioned glyphs at their kerned positions,
// so letters can move independently without losing the font's spacing.
export function splitGlyphs(el) {
  const text = el.textContent;
  const node = el.firstChild;
  const range = document.createRange();
  const base = el.getBoundingClientRect();
  const boxes = [];
  for (let i = 0; i < text.length; i++) {
    range.setStart(node, i);
    range.setEnd(node, i + 1);
    const r = range.getBoundingClientRect();
    boxes.push({ ch: text[i], x: r.left - base.left, y: r.top - base.top, w: r.width, h: r.height });
  }
  // Keep the text in flow (invisible) so the word keeps its width and baseline.
  el.textContent = '';
  el.style.position = 'relative';
  el.append(h('span', { text, style: { visibility: 'hidden' } }));
  return boxes.map((b) => {
    const g = h('span', { class: 'sr-glyph', text: b.ch, style: { position: 'absolute', left: `${b.x}px`, top: `${b.y}px`, width: `${b.w}px`, height: `${b.h}px` } });
    el.append(g);
    g._box = b;
    return g;
  });
}

// Rect of an element relative to the stage, at the current transforms.
export function rectOf(el, stage) {
  const r = el.getBoundingClientRect();
  const s = stage.getBoundingClientRect();
  return { x: r.left - s.left, y: r.top - s.top, w: r.width, h: r.height, cx: r.left - s.left + r.width / 2, cy: r.top - s.top + r.height / 2 };
}

// ---------------------------------------------------------------- scenes and cues
export const scenes = [];
export function scene(def) {
  scenes.push(def);
  return def;
}
export function renderAt(t) {
  for (const s of scenes) {
    const on = t >= s.t0 && t < s.t1;
    if (s.root && on !== s._on) {
      s.root.style.display = on ? '' : 'none';
      s._on = on;
    }
    if (on) s.update(t);
  }
}

// Sound cues registered by the scenes. The synthesizer reads them, so every
// pop, key and whoosh lands on the frame its animation starts.
export const cues = [];
export function cue(t, type, params = {}) {
  cues.push({ t: Math.round(t * 10000) / 10000, type, ...params });
}
// Stereo position for a stage x coordinate.
export const panX = (x) => clamp((x - W / 2) / (W / 2), -1, 1) * 0.7;
