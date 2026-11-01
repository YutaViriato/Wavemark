// Synthesizes the showreel soundtrack from the score (stage/score.mjs) and the
// sound cues the scenes register (out/cues.json, from `node render.mjs --cues`).
// Deterministic: same inputs, same samples.
//
//   node audio/synth.mjs            → out/soundtrack.wav (48 kHz, 24-bit, stereo)
//   node audio/synth.mjs --stems    → also out/stems/*.wav
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { events, kicks, DURATION } from '../stage/score.mjs';

const here = path.dirname(fileURLToPath(import.meta.url));
const out = path.join(here, '../out');
const SR = 48000;
const LEN = Math.round(DURATION * SR);
const PAD = SR * 3; // room for tails, trimmed at the end
const TAU = Math.PI * 2;
const mtof = (m) => 440 * Math.pow(2, (m - 69) / 12);
const clamp = (x, a, b) => (x < a ? a : x > b ? b : x);
const db = (d) => Math.pow(10, d / 20);

// ---------------------------------------------------------------- noise
let rs = 0x9e3779b9;
const rnd = () => { rs ^= rs << 13; rs ^= rs >>> 17; rs ^= rs << 5; return ((rs >>> 0) / 4294967296) * 2 - 1; };
const seedNoise = (s) => { rs = (s * 2654435761) >>> 0 || 1; };

// ---------------------------------------------------------------- buses
const bus = () => [new Float32Array(LEN + PAD), new Float32Array(LEN + PAD)];
const B = { kick: bus(), perc: bus(), bass: bus(), pads: bus(), keys: bus(), arp: bus(), fx: bus(), sfx: bus(), verb: bus(), delay: bus() };
function add(dst, t, buf, gain = 1, pan = 0) {
  const i0 = Math.round(t * SR);
  const a = ((clamp(pan, -1, 1) + 1) * Math.PI) / 4;
  const stereo = Array.isArray(buf);
  const l = stereo ? buf[0] : buf, r = stereo ? buf[1] : buf;
  const gl = Math.cos(a) * Math.SQRT2 * gain, gr = Math.sin(a) * Math.SQRT2 * gain;
  const [L, R] = dst;
  for (let i = 0; i < l.length; i++) {
    const j = i0 + i;
    if (j < 0) continue;
    if (j >= L.length) break;
    L[j] += l[i] * gl;
    R[j] += r[i] * gr;
  }
}
const send = (busName, t, buf, gain, pan, verb = 0, dly = 0) => {
  add(B[busName], t, buf, gain, pan);
  if (verb) add(B.verb, t, buf, gain * verb, pan);
  if (dly) add(B.delay, t, buf, gain * dly, pan);
};

// ---------------------------------------------------------------- DSP blocks
// Topology-preserving state variable filter (Cytomic), modulation-safe.
function svf() {
  let ic1 = 0, ic2 = 0;
  return (x, fc, q = 0.707, mode = 0) => {
    const g = Math.tan((Math.PI * clamp(fc, 10, SR * 0.45)) / SR);
    const k = 1 / q;
    const a1 = 1 / (1 + g * (g + k)), a2 = g * a1, a3 = g * a2;
    const v3 = x - ic2;
    const v1 = a1 * ic1 + a2 * v3;
    const v2 = ic2 + a2 * ic1 + a3 * v3;
    ic1 = 2 * v1 - ic1;
    ic2 = 2 * v2 - ic2;
    return mode === 0 ? v2 : mode === 1 ? v1 : x - k * v1 - v2; // low, band, high
  };
}
function onePole(fc) {
  const a = Math.exp((-TAU * fc) / SR);
  let z = 0;
  return (x) => (z = x * (1 - a) + z * a);
}
// PolyBLEP saw.
function saw() {
  let ph = 0;
  const blep = (t, dt) => { if (t < dt) { t /= dt; return t + t - t * t - 1; } if (t > 1 - dt) { t = (t - 1) / dt; return t * t + t + t + 1; } return 0; };
  return { set(p) { ph = p; }, next(f) { const dt = f / SR; let v = 2 * ph - 1 - blep(ph, dt); ph += dt; if (ph >= 1) ph -= 1; return v; } };
}
function square() {
  let ph = 0;
  const blep = (t, dt) => { if (t < dt) { t /= dt; return t + t - t * t - 1; } if (t > 1 - dt) { t = (t - 1) / dt; return t * t + t + t + 1; } return 0; };
  return { set(p) { ph = p; }, next(f) { const dt = f / SR; let v = ph < 0.5 ? 1 : -1; v += blep(ph, dt); v -= blep((ph + 0.5) % 1, dt); ph += dt; if (ph >= 1) ph -= 1; return v; } };
}
const env = (t, a, d, s, r, dur) => {
  // ADSR with exponential decay and release, evaluated at time t
  let v;
  if (t < a) v = t / a;
  else if (t < dur) v = s + (1 - s) * Math.exp(-(t - a) / d);
  else v = (s + (1 - s) * Math.exp(-(dur - a) / d)) * Math.exp(-(t - dur) / r);
  return v;
};
const buf = (sec) => new Float32Array(Math.max(1, Math.round(sec * SR)));
const sat = (x, k = 1.5) => Math.tanh(k * x) / Math.tanh(k);

// ---------------------------------------------------------------- drums
function kick(vel, lp) {
  const b = buf(0.5);
  let ph = 0;
  seedNoise(3);
  const f1 = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const f = 47 + 145 * Math.exp(-t / 0.032) + 1100 * Math.exp(-t / 0.0028);
    ph += (TAU * f) / SR;
    const amp = Math.min(1, t / 0.0015) * Math.exp(-t / 0.22) * (t > 0.32 ? Math.exp(-(t - 0.32) / 0.03) : 1);
    let s = Math.sin(ph) * amp + rnd() * Math.exp(-t / 0.0022) * 0.35;
    s = sat(s * 1.25, 1.7);
    b[i] = lp < 19000 ? f1(s, lp, 0.6) : s;
  }
  return scale(b, vel);
}
function clap(vel) {
  const b = buf(0.45);
  seedNoise(11);
  const bp = svf(), hp = svf();
  const hits = [0, 0.0095, 0.019, 0.029];
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    let e = 0;
    for (const h of hits) if (t >= h) e += Math.exp(-(t - h) / (h === 0.029 ? 0.12 : 0.004)) * (h === 0.029 ? 0.9 : 1);
    const n = rnd();
    b[i] = hp(bp(n, 1250, 0.85, 1), 500, 0.7, 2) * e * 1.6;
  }
  return scale(b, vel);
}
const HAT_F = [205.3, 304.4, 369.6, 522.7, 540, 800].map((f) => f * 1.62);
function hat(vel, open) {
  const b = buf(open ? 0.45 : 0.08);
  const osc = HAT_F.map(() => square());
  osc.forEach((o, i) => o.set((i * 0.37) % 1));
  const bp = svf(), hp = svf();
  seedNoise(open ? 21 : 19);
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    let m = 0;
    for (let k = 0; k < 6; k++) m += osc[k].next(HAT_F[k]);
    const s = m / 6 + rnd() * 0.5;
    const e = Math.min(1, t / 0.0008) * Math.exp(-t / (open ? 0.16 : 0.022));
    b[i] = hp(bp(s, 10500, 0.9, 1), 7500, 0.7, 2) * e * 1.4;
  }
  return scale(b, vel);
}
function shaker(vel) {
  const b = buf(0.07);
  seedNoise(31 + Math.floor(vel * 100));
  const hp = svf();
  for (let i = 0; i < b.length; i++) { const t = i / SR; b[i] = hp(rnd(), 6500, 0.8, 2) * Math.min(1, t / 0.006) * Math.exp(-t / 0.028); }
  return scale(b, vel);
}
function snare(vel, pitch = 1) {
  const b = buf(0.3);
  let ph = 0;
  seedNoise(41);
  const hp = svf(), lp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const f = 185 * pitch * (1 + 0.6 * Math.exp(-t / 0.012));
    ph += (TAU * f) / SR;
    const tone = Math.sin(ph) * Math.exp(-t / 0.07) * 0.55;
    const n = lp(hp(rnd(), 1700, 0.7, 2), 9000, 0.7, 0) * Math.exp(-t / 0.12) * 0.9;
    b[i] = sat((tone + n) * Math.min(1, t / 0.001), 1.3);
  }
  return scale(b, vel);
}
function cymbal(vel, len = 2.2, seed = 51) {
  const L = buf(len), R = buf(len);
  const osc = HAT_F.map(() => square());
  const hpL = svf(), hpR = svf(), lpL = svf(), lpR = svf();
  seedNoise(seed);
  for (let i = 0; i < L.length; i++) {
    const t = i / SR;
    let m = 0;
    for (let k = 0; k < 6; k++) m += osc[k].next(HAT_F[k] * 0.77);
    const e = Math.min(1, t / 0.002) * Math.exp(-t / 0.55) * (1 - 0.3 * Math.exp(-t / 0.05));
    const nl = rnd(), nr = rnd();
    L[i] = lpL(hpL(nl * 0.8 + (m / 6) * 0.4, 3800, 0.6, 2), 13000, 0.6, 0) * e;
    R[i] = lpR(hpR(nr * 0.8 + (m / 6) * 0.4, 3800, 0.6, 2), 13000, 0.6, 0) * e;
  }
  return [scale(L, vel), scale(R, vel)];
}
function scale(b, g) { for (let i = 0; i < b.length; i++) b[i] *= g; return b; }
function reverse([L, R], len) {
  const n = Math.round(len * SR);
  const l = buf(len), r = buf(len);
  for (let i = 0; i < n; i++) {
    const j = n - 1 - i;
    const fade = Math.min(1, i / (0.12 * SR));
    l[i] = (L[j] || 0) * fade;
    r[i] = (R[j] || 0) * fade;
  }
  return [l, r];
}

// ---------------------------------------------------------------- FX
function riser(dur, vel, seed = 61) {
  const L = buf(dur), R = buf(dur);
  const bl = svf(), br = svf();
  const s1 = saw(), s2 = saw();
  seedNoise(seed);
  for (let i = 0; i < L.length; i++) {
    const t = i / SR, p = t / dur;
    const fc = 300 * Math.pow(9000 / 300, p * p * 0.6 + p * 0.4);
    const a = Math.pow(p, 2.2);
    const tone = (s1.next(110 * Math.pow(4, p)) + s2.next(110.6 * Math.pow(4, p))) * 0.06 * a;
    L[i] = (bl(rnd(), fc, 3.2, 1) * 0.9 + tone) * a;
    R[i] = (br(rnd(), fc * 1.04, 3.2, 1) * 0.9 + tone) * a;
  }
  const tail = Math.round(0.004 * SR);
  for (let i = 0; i < tail; i++) { const k = i / tail; L[L.length - 1 - i] *= k; R[R.length - 1 - i] *= k; }
  return [scale(L, vel), scale(R, vel)];
}
function downlifter(dur, vel) {
  const L = buf(dur), R = buf(dur);
  const bl = svf(), br = svf();
  seedNoise(71);
  for (let i = 0; i < L.length; i++) {
    const t = i / SR, p = t / dur;
    const fc = 7000 * Math.pow(250 / 7000, p);
    const a = Math.exp(-p * 3.2) * Math.min(1, t / 0.01);
    L[i] = bl(rnd(), fc, 1.6, 1) * a;
    R[i] = br(rnd(), fc * 0.97, 1.6, 1) * a;
  }
  return [scale(L, vel), scale(R, vel)];
}
function impact(vel) {
  const b = buf(2.4);
  let ph = 0;
  seedNoise(81);
  const lp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const f = 30 + 52 * Math.exp(-t / 0.35);
    ph += (TAU * f) / SR;
    const sub = Math.sin(ph) * Math.exp(-t / 0.75) * Math.min(1, t / 0.004);
    const n = lp(rnd(), 1400 * Math.exp(-t / 0.5) + 200, 0.7, 0) * Math.exp(-t / 0.22) * 0.9;
    b[i] = sat(sub * 1.1 + n, 1.4);
  }
  return scale(b, vel);
}

// ---------------------------------------------------------------- tonal
function bass(note, dur, vel, cutoff) {
  const len = dur + 0.08;
  const b = buf(len);
  const f = mtof(note);
  const a1 = saw(), a2 = saw();
  a2.set(0.31);
  const flt = svf();
  let ph = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const e = env(t, 0.003, 0.09, 0.62, 0.035, dur);
    const fc = cutoff * (1 + 3.5 * Math.exp(-t / 0.05));
    let s = (a1.next(f * 0.996) + a2.next(f * 1.004)) * 0.5;
    s = flt(s, fc, 1.1, 0);
    ph += (TAU * f * 0.5) / SR;
    s = s * 0.8 + Math.sin(ph) * 0.55;
    b[i] = sat(s * e * 1.2, 1.6);
  }
  return scale(b, vel);
}
function subNote(note, dur, vel, final) {
  const rel = final ? 0.9 : 0.06;
  const b = buf(dur + rel);
  const f = mtof(note);
  let ph = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * f) / SR;
    const e = final ? Math.min(1, t / 0.005) * Math.exp(-t / 0.7) : env(t, 0.012, 0.6, 0.85, rel, dur);
    b[i] = Math.sin(ph) * e;
  }
  return scale(b, vel);
}
function pad(notes, dur, vel, cutoff, { sweepDown = false, final = false } = {}) {
  const rel = final ? 1.9 : 0.5;
  const len = dur + rel;
  const L = buf(len), R = buf(len);
  const voices = [];
  const detune = [-0.11, -0.05, 0, 0.05, 0.11];
  notes.forEach((n, ni) => detune.forEach((d, k) => {
    const o = saw();
    o.set(((ni * 7 + k * 3) % 11) / 11);
    voices.push({ o, f: mtof(n + d), pan: ((k - 2) / 2) * 0.8 * (ni % 2 ? -1 : 1) });
  }));
  const fl = svf(), fr = svf(), hl = svf(), hr = svf();
  const attack = final ? 0.01 : 0.2;
  for (let i = 0; i < L.length; i++) {
    const t = i / SR;
    let l = 0, r = 0;
    for (const v of voices) {
      const s = v.o.next(v.f);
      const a = ((v.pan + 1) * Math.PI) / 4;
      l += s * Math.cos(a);
      r += s * Math.sin(a);
    }
    const lfo = 1 + 0.12 * Math.sin(TAU * 0.35 * t);
    let fc = cutoff * lfo;
    if (sweepDown) fc = Math.max(240, cutoff * Math.pow(0.05, clamp(t / dur, 0, 1)));
    if (final) fc = cutoff * (0.35 + 0.65 * Math.exp(-t / 0.5));
    const e = env(t, attack, final ? 0.9 : 1.5, final ? 0.0 : 0.8, rel, dur);
    L[i] = hl(fl(l, fc, 0.8, 0), 190, 0.7, 2) * e;
    R[i] = hr(fr(r, fc, 0.8, 0), 190, 0.7, 2) * e;
  }
  const g = vel / Math.sqrt(voices.length);
  return [scale(L, g), scale(R, g)];
}
function arpNote(note, dur, vel, cutoff) {
  const b = buf(dur + 0.25);
  const f = mtof(note);
  const a = saw(), q = square();
  const flt = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const e = Math.min(1, t / 0.002) * Math.exp(-t / 0.13);
    const fc = cutoff * (0.5 + 2.2 * Math.exp(-t / 0.045));
    b[i] = flt(a.next(f) * 0.6 + q.next(f * 1.002) * 0.4, fc, 1.4, 0) * e;
  }
  return scale(b, vel);
}
// Karplus-Strong pluck with a soft sine body.
function pluck(note, dur, vel) {
  const f = mtof(note);
  const len = dur + 0.8;
  const b = buf(len);
  const n = Math.round(SR / f);
  const line = new Float32Array(n);
  seedNoise(note);
  const lp0 = onePole(3200);
  for (let i = 0; i < n; i++) line[i] = lp0(rnd());
  let idx = 0, ph = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const cur = line[idx], nxt = line[(idx + 1) % n];
    const v = 0.4985 * (cur + nxt);
    line[idx] = v;
    idx = (idx + 1) % n;
    ph += (TAU * f) / SR;
    b[i] = cur * 0.9 + Math.sin(ph) * 0.35 * Math.exp(-t / 0.5);
    b[i] *= Math.min(1, t / 0.001) * (t > dur ? Math.exp(-(t - dur) / 0.3) : 1);
  }
  return scale(b, vel);
}
// FM electric piano / bell: ratio 1:2 for the body, 1:3.5 for a glassy tine.
function bell(note, dur, vel, { final = false } = {}) {
  const f = mtof(note);
  const tail = final ? 2.4 : 0.9;
  const b = buf(dur + tail);
  let pc = 0, pm = 0, pt = 0, pc2 = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const idx = 0.35 + 2.4 * Math.exp(-t / 0.22);
    pm += (TAU * f * 2) / SR;
    pc += (TAU * f) / SR;
    const body = Math.sin(pc + idx * Math.sin(pm));
    pt += (TAU * f * 3.5) / SR;
    const tine = Math.sin(pt) * Math.exp(-t / 0.06) * 0.35;
    pc2 += (TAU * f * 0.5) / SR;
    const low = Math.sin(pc2) * 0.18 * Math.exp(-t / 0.6);
    const amp = Math.min(1, t / 0.0025) * Math.exp(-t / (final ? 1.4 : 0.75)) * (t > dur ? Math.exp(-(t - dur) / (final ? 1.2 : 0.22)) : 1);
    b[i] = (body * 0.7 + tine + low) * amp;
  }
  return scale(b, vel);
}

// ---------------------------------------------------------------- sound design (cues)
function keyClick(vel, seed) {
  const b = buf(0.06);
  seedNoise(seed);
  const bp = svf(), bp2 = svf();
  const thock = 150 + ((seed * 37) % 70);
  let ph = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * thock * (1 + 0.5 * Math.exp(-t / 0.004))) / SR;
    const click = bp(rnd(), 3600 + ((seed * 131) % 900), 1.4, 1) * Math.exp(-t / 0.0016) * 2.2;
    const tick2 = t > 0.011 ? bp2(rnd(), 5200, 2, 1) * Math.exp(-(t - 0.011) / 0.0012) * 0.8 : 0;
    const body = Math.sin(ph) * Math.exp(-t / 0.016) * 0.7;
    b[i] = (click + tick2 + body) * Math.min(1, t / 0.0004);
  }
  return scale(b, vel);
}
function popSound(note, vel) {
  const f = mtof(note);
  const b = buf(0.16);
  let ph = 0;
  seedNoise(note * 3);
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * f * (1 + 0.7 * Math.exp(-t / 0.01))) / SR;
    b[i] = (Math.sin(ph) * Math.exp(-t / 0.055) + rnd() * Math.exp(-t / 0.0015) * 0.3) * Math.min(1, t / 0.0005);
  }
  return scale(b, vel);
}
function tickSound(vel, note) {
  const b = buf(0.05);
  const f = note ? mtof(note) : 2600;
  let ph = 0;
  seedNoise(99);
  const hp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * f) / SR;
    b[i] = hp(rnd(), 4000, 0.7, 2) * Math.exp(-t / 0.0012) * 0.8 + Math.sin(ph) * Math.exp(-t / 0.012) * 0.45;
  }
  return scale(b, vel);
}
function zip(vel) {
  const b = buf(0.14);
  let ph = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR, p = t / 0.14;
    ph += (TAU * (700 * Math.pow(4, p))) / SR;
    b[i] = Math.sin(ph) * Math.sin(Math.PI * p) * 0.5;
  }
  return scale(b, vel);
}
function whoosh(dur, vel, { tone = 1, rev = false, from, to, pan = 0 } = {}) {
  const L = buf(dur + 0.05), R = buf(dur + 0.05);
  const bl = svf(), br = svf();
  seedNoise(Math.round(dur * 1000 + tone * 77));
  const p0 = from ?? pan, p1 = to ?? pan;
  for (let i = 0; i < L.length; i++) {
    const t = i / SR, p = clamp(t / dur, 0, 1);
    const shape = rev ? Math.pow(p, 2.4) * (p < 0.98 ? 1 : (1 - p) / 0.02) : Math.sin(Math.PI * Math.pow(p, 0.6)) * Math.exp(-p * 0.6);
    const fc = (rev ? 300 + 3200 * Math.pow(p, 1.5) : 300 + 2600 * Math.sin(Math.PI * Math.pow(p, 0.7))) * tone;
    const pp = p0 + (p1 - p0) * p;
    const a = ((clamp(pp, -1, 1) + 1) * Math.PI) / 4;
    const nl = bl(rnd(), fc, 1.3, 1), nr = br(rnd(), fc * 1.03, 1.3, 1);
    L[i] = nl * shape * Math.cos(a) * 1.6;
    R[i] = nr * shape * Math.sin(a) * 1.6;
  }
  return [scale(L, vel), scale(R, vel)];
}
function marker(dur, vel) {
  const b = buf(dur + 0.03);
  seedNoise(123);
  const bp = svf(), lfo = onePole(30);
  for (let i = 0; i < b.length; i++) {
    const t = i / SR, p = clamp(t / dur, 0, 1);
    const speed = Math.sin(Math.PI * Math.pow(p, 0.8));
    const grain = 0.6 + 0.4 * Math.abs(lfo(rnd() * 3));
    b[i] = bp(rnd(), 2600 + 2400 * speed, 1.6, 1) * speed * grain * 1.4;
  }
  return scale(b, vel);
}
function scribble(dur, vel) {
  const b = buf(dur);
  seedNoise(321);
  const bp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const strokes = Math.pow(Math.abs(Math.sin(TAU * 7.5 * t)), 1.5);
    b[i] = bp(rnd(), 3200 + 1600 * strokes, 1.8, 1) * strokes * Math.sin(Math.PI * t / dur) * 1.2;
  }
  return scale(b, vel);
}
function rippleSound(dur, vel) {
  const L = buf(dur + 0.1), R = buf(dur + 0.1);
  seedNoise(555);
  for (let g = 0; g < 70; g++) {
    const t0 = (g / 70) * dur * (0.9 + 0.1 * rnd());
    const f = 1800 + 2600 * (g / 70) + rnd() * 300;
    const pan = Math.sin(g * 2.3) * 0.8;
    const a = ((pan + 1) * Math.PI) / 4;
    const i0 = Math.round(t0 * SR);
    for (let i = 0; i < 0.03 * SR; i++) {
      const t = i / SR;
      const s = Math.sin(TAU * f * t) * Math.exp(-t / 0.006) * 0.25;
      if (i0 + i < L.length) { L[i0 + i] += s * Math.cos(a); R[i0 + i] += s * Math.sin(a); }
    }
  }
  return [scale(L, vel), scale(R, vel)];
}
function errorSound(vel, soft) {
  const b = buf(0.4);
  let p1 = 0, p2 = 0, pm = 0;
  const lp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    pm += (TAU * 146.8 * 1.5) / SR;
    p1 += (TAU * 146.8) / SR;
    p2 += (TAU * 155.6) / SR;
    const s = Math.sin(p1 + 1.8 * Math.exp(-t / 0.05) * Math.sin(pm)) * 0.6 + Math.sin(p2) * 0.4 * (t > 0.09 ? 1 : 0);
    b[i] = lp(s, 1100, 0.7, 0) * Math.min(1, t / 0.002) * Math.exp(-t / (soft ? 0.08 : 0.14));
  }
  return scale(b, vel);
}
function successSound(vel, soft) {
  const out = buf(1.6);
  const notes = soft ? [[0, 86]] : [[0, 81], [0.085, 86], [0.17, 90]];
  for (const [dt, n] of notes) {
    const b = bell(n, 0.25, 0.5, {});
    const i0 = Math.round(dt * SR);
    for (let i = 0; i < b.length && i0 + i < out.length; i++) out[i0 + i] += b[i];
  }
  return scale(out, vel);
}
function boing(vel) {
  const b = buf(0.25);
  let ph = 0;
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * (320 + 900 * (1 - Math.exp(-t / 0.03)))) / SR;
    b[i] = Math.sin(ph) * Math.exp(-t / 0.07) * Math.min(1, t / 0.002);
  }
  return scale(b, vel);
}
function thump(vel) {
  const b = buf(0.25);
  let ph = 0;
  seedNoise(777);
  const lp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * (60 + 50 * Math.exp(-t / 0.02))) / SR;
    b[i] = Math.sin(ph) * Math.exp(-t / 0.07) * 0.9 + lp(rnd(), 2600, 0.7, 0) * Math.exp(-t / 0.025) * 0.9;
  }
  return scale(b, vel);
}
function mouseClick(vel) {
  const b = buf(0.12);
  seedNoise(888);
  const bp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    const e = Math.exp(-t / 0.0012) + (t > 0.07 ? Math.exp(-(t - 0.07) / 0.0009) * 0.6 : 0);
    b[i] = bp(rnd(), 4200, 2.2, 1) * e * 2.2;
  }
  return scale(b, vel);
}
function charge(dur, vel) {
  const b = buf(dur);
  let ph = 0;
  seedNoise(999);
  const bp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR, p = t / dur;
    ph += (TAU * (180 * Math.pow(5, p))) / SR;
    b[i] = (Math.sin(ph) * 0.35 + bp(rnd(), 800 + 6000 * p, 2, 1) * 0.6) * Math.pow(p, 2) * (p > 0.97 ? (1 - p) / 0.03 : 1);
  }
  return scale(b, vel);
}
function slam(vel) {
  const b = buf(0.4);
  let ph = 0;
  seedNoise(4242);
  const lp = svf();
  for (let i = 0; i < b.length; i++) {
    const t = i / SR;
    ph += (TAU * (70 + 90 * Math.exp(-t / 0.015))) / SR;
    b[i] = Math.sin(ph) * Math.exp(-t / 0.11) * 0.8 + lp(rnd(), 3800, 0.7, 0) * Math.exp(-t / 0.035) * 0.8;
  }
  return scale(b, vel);
}

// ---------------------------------------------------------------- render the score
const ev = events();
for (const e of ev) {
  const introBoost = e.t < 4 ? db(4) : 1;
  switch (e.inst) {
    case 'kick': send('kick', e.t, kick(e.vel, e.lp), 1, 0); break;
    case 'clap': send('perc', e.t, clap(e.vel), 0.55, 0, 0.3); break;
    case 'hat': send('perc', e.t, hat(e.vel, false), 0.42, 0.22); break;
    case 'ohat': send('perc', e.t, hat(e.vel, true), 0.3, -0.2, 0.06); break;
    case 'shaker': send('perc', e.t, shaker(e.vel), 0.3, 0.4); break;
    case 'snare': send('perc', e.t, snare(e.vel, e.pitch), 0.5, 0, 0.28); break;
    case 'crash': send('fx', e.t, cymbal(e.vel), 0.34, 0, 0.25); break;
    case 'reverse': { const c = cymbal(1, 2.2, 57); send('fx', e.t, reverse(c, e.dur), 0.5 * e.vel, 0, 0.2); break; }
    case 'riser': send('fx', e.t, riser(e.dur, e.vel), 0.26, 0, 0.3); break;
    case 'downlifter': send('fx', e.t, downlifter(e.dur, e.vel), 0.3, 0, 0.4); break;
    case 'impact': send('fx', e.t, impact(e.vel), 0.8, 0, 0.45); break;
    case 'bass': send('bass', e.t, bass(e.note, e.dur, e.vel, e.cutoff), 0.5, 0); break;
    case 'sub': send('bass', e.t, subNote(e.note, e.dur, e.vel, e.final), 0.32, 0); break;
    case 'pad': send('pads', e.t, pad(e.notes, e.dur, e.vel, e.cutoff, e), introBoost, 0, 0.3, 0.05); break;
    case 'arp': send('arp', e.t, arpNote(e.note, e.dur, e.vel, e.cutoff), 1, Math.sin(e.t * 9) * 0.45, 0.15, 0.4); break;
    case 'pluck': send('keys', e.t, pluck(e.note, e.dur, e.vel), 1.5, 0.15, 0.45, 0.35); break;
    case 'bell': {
      send('keys', e.t, bell(e.note, e.dur, e.vel, e), 1, 0.16, 0.38, 0.3);
      if (e.double) send('keys', e.t, bell(e.note - 12, e.dur, e.vel * 0.5, e), 0.8, -0.2, 0.3, 0.12);
      break;
    }
  }
}

// ---------------------------------------------------------------- sound cues
const cuesFile = path.join(out, 'cues.json');
const cues = fs.existsSync(cuesFile) ? JSON.parse(fs.readFileSync(cuesFile, 'utf8')) : [];
if (!cues.length) console.warn('no out/cues.json: run `node render.mjs --cues` first');
let keySeed = 1;
for (const c of cues) {
  const v = c.vel ?? 0.6, p = c.pan ?? 0;
  switch (c.type) {
    case 'key': for (let k = 0; k < (c.n || 1); k++) send('sfx', c.t + k * 0.018, keyClick(v * (k ? 0.6 : 1), keySeed++), 0.5, p, 0.06); break;
    case 'type': for (let k = 0; k < (c.n || 5); k++) send('sfx', c.t + k * 0.022 + ((k * 7) % 5) * 0.003, keyClick(v * (0.7 + ((k * 13) % 5) * 0.06), keySeed++), 0.4, p, 0.05); break;
    case 'enter': send('sfx', c.t, keyClick(v, 7), 0.6, p, 0.08); send('sfx', c.t, thump(v * 0.5), 0.35, p); break;
    case 'save': send('sfx', c.t - 0.03, keyClick(v * 0.7, 5), 0.45, p); send('sfx', c.t, keyClick(v, 6), 0.5, p, 0.06); break;
    case 'pop': send('sfx', c.t, popSound(c.note ?? 81, v), 0.36, p, 0.15); break;
    case 'tick': send('sfx', c.t, tickSound(v, c.note), 0.32, p, 0.08); break;
    case 'zip': send('sfx', c.t, zip(v), 0.22, p, 0.12); break;
    case 'whoosh': send('sfx', c.t, whoosh(c.dur ?? 0.5, v, c), 0.45, 0, 0.2); break;
    case 'swipe': send('sfx', c.t, whoosh(c.dur ?? 0.4, v, { ...c, tone: 2.2 }), 0.3, 0); break;
    case 'marker': send('sfx', c.t, marker(c.dur ?? 0.4, v), 0.3, p, 0.15); break;
    case 'scribble': send('sfx', c.t, scribble(c.dur ?? 0.3, v), 0.3, p, 0.12); break;
    case 'ripple': send('sfx', c.t, rippleSound(c.dur ?? 0.5, v), 0.45, 0, 0.2); break;
    case 'error': send('sfx', c.t, errorSound(v, c.soft), 0.55, p, 0.15); break;
    case 'success': send('sfx', c.t, successSound(v, c.soft), 0.36, p, 0.35, 0.2); break;
    case 'pluck-out': send('sfx', c.t, boing(v), 0.32, p, 0.15); break;
    case 'thump': send('sfx', c.t, thump(v), 0.55, p, 0.1); break;
    case 'click': send('sfx', c.t, mouseClick(v), 0.5, p, 0.05); break;
    case 'charge': send('sfx', c.t, charge(c.dur ?? 0.25, v), 0.3, 0, 0.1); break;
    case 'slam': send('sfx', c.t, slam(v), 0.55, 0, 0.3); break;
    default: console.warn('unknown cue', c.type);
  }
}

// ---------------------------------------------------------------- effects
// Dattorro plate reverb (1997), delays scaled from 29761 Hz.
function plate([inL, inR], { decay = 0.72, damping = 0.35, predelay = 0.012 } = {}) {
  const k = SR / 29761;
  const D = (n) => Math.round(n * k);
  const mk = (n) => ({ b: new Float32Array(n + 64), i: 0, n });
  const tap = (d, back) => d.b[(d.i - back + d.b.length * 4) % d.b.length];
  const push = (d, v) => { d.b[d.i] = v; d.i = (d.i + 1) % d.b.length; };
  const ap = (d, x, g) => { const z = d.b[(d.i - d.n + d.b.length) % d.b.length]; const w = x + g * z; push(d, w); return z - g * w; };
  const apMod = (d, x, g, len) => {
    const r = d.i - len;
    const r0 = Math.floor(r), fr = r - r0;
    const z = d.b[(r0 + d.b.length * 2) % d.b.length] * (1 - fr) + d.b[(r0 + 1 + d.b.length * 2) % d.b.length] * fr;
    const w = x + g * z;
    push(d, w);
    return z - g * w;
  };
  const pre = mk(Math.round(predelay * SR) + 1);
  const id = [mk(D(142)), mk(D(107)), mk(D(379)), mk(D(277))];
  const l1 = mk(D(672) + D(16) + 4), l2 = mk(D(4453)), l3 = mk(D(1800)), l4 = mk(D(3720));
  const r1 = mk(D(908) + D(16) + 4), r2 = mk(D(4217)), r3 = mk(D(2656)), r4 = mk(D(3163));
  let bw = 0, dl = 0, dr = 0, fbL = 0, fbR = 0;
  const outL = new Float32Array(inL.length), outR = new Float32Array(inR.length);
  for (let i = 0; i < inL.length; i++) {
    push(pre, (inL[i] + inR[i]) * 0.5);
    let x = tap(pre, pre.n);
    bw += 0.9995 * (x - bw) * 0.9;
    x = ap(id[0], bw, 0.75); x = ap(id[1], x, 0.75); x = ap(id[2], x, 0.625); x = ap(id[3], x, 0.625);
    const t = i / SR;
    const exc = D(16);
    // left tank
    let a = apMod(l1, x + fbR, -0.7, D(672) + exc * Math.sin(TAU * 0.5 * t));
    push(l2, a);
    a = tap(l2, l2.n);
    dl += (1 - damping) * (a - dl);
    a = ap(l3, dl * decay, 0.5);
    push(l4, a);
    fbL = tap(l4, l4.n) * decay;
    // right tank
    let b = apMod(r1, x + fbL, -0.7, D(908) + exc * Math.sin(TAU * 0.33 * t + 1));
    push(r2, b);
    b = tap(r2, r2.n);
    dr += (1 - damping) * (b - dr);
    b = ap(r3, dr * decay, 0.5);
    push(r4, b);
    fbR = tap(r4, r4.n) * decay;
    outL[i] = 0.6 * (tap(r2, D(266)) + tap(r2, D(2974)) - tap(r3, D(1913)) + tap(r4, D(1996)) - tap(l2, D(1990)) - tap(l3, D(187)) - tap(l4, D(1066)));
    outR[i] = 0.6 * (tap(l2, D(353)) + tap(l2, D(3627)) - tap(l3, D(1228)) + tap(l4, D(2673)) - tap(r2, D(2111)) - tap(r3, D(335)) - tap(r4, D(121)));
  }
  // keep the low end out of the tail
  const hl = svf(), hr = svf();
  for (let i = 0; i < outL.length; i++) { outL[i] = hl(outL[i], 220, 0.7, 2); outR[i] = hr(outR[i], 220, 0.7, 2); }
  return [outL, outR];
}
function pingPong([inL, inR], time = 0.375, fb = 0.38) {
  const n = Math.round(time * SR);
  const bl = new Float32Array(n), br = new Float32Array(n);
  const lpL = svf(), lpR = svf(), hpL = svf(), hpR = svf();
  const oL = new Float32Array(inL.length), oR = new Float32Array(inR.length);
  let j = 0;
  for (let i = 0; i < inL.length; i++) {
    const yl = bl[j], yr = br[j];
    const fl = hpL(lpL(yr * fb + (inL[i] + inR[i]) * 0.5, 4200, 0.7, 0), 350, 0.7, 2);
    const fr = hpR(lpR(yl * fb, 4200, 0.7, 0), 350, 0.7, 2);
    bl[j] = fl; br[j] = fr;
    oL[i] = yl; oR[i] = yr;
    j = (j + 1) % n;
  }
  return [oL, oR];
}

// Sidechain from the kick: a quick dip and a musical release.
const KT = kicks();
function duckGain(depth, release = 0.11) {
  const g = new Float32Array(LEN + PAD).fill(1);
  for (const k of KT) {
    const i0 = Math.round(k * SR);
    for (let i = 0; i < release * 6 * SR; i++) {
      const t = i / SR;
      const e = t < 0.004 ? t / 0.004 : Math.exp(-(t - 0.004) / release);
      const v = 1 - depth * e;
      if (i0 + i < g.length) g[i0 + i] = Math.min(g[i0 + i], v);
    }
  }
  return g;
}
const apply = ([L, R], g) => { for (let i = 0; i < L.length; i++) { L[i] *= g[i]; R[i] *= g[i]; } };
apply(B.bass, duckGain(0.8, 0.09));
const duckMusic = duckGain(0.5, 0.12);
for (const k of ['pads', 'arp']) apply(B[k], duckMusic);
apply(B.keys, duckGain(0.22, 0.1));
const verbOut = plate(B.verb, { decay: 0.74, damping: 0.4 });
const dlyOut = pingPong(B.delay);
const verbDuck = duckGain(0.4, 0.15);
apply([verbOut[0], verbOut[1]], verbDuck);

// ---------------------------------------------------------------- mix and master
// Each element is levelled to a target RMS over the drops (8–24 s), so the
// balance is measured rather than guessed. FX and SFX are sparse: they are
// levelled on their loud moments instead (RMS above a gate).
const stems = { kick: B.kick, perc: B.perc, bass: B.bass, pads: B.pads, keys: B.keys, arp: B.arp, fx: B.fx, sfx: B.sfx, verb: verbOut, delay: dlyOut };
const TARGET_RMS = { kick: -17, perc: -24.5, bass: -19.5, pads: -22.5, keys: -21.5, arp: -27.5, verb: -27, delay: -31 };
const GATED = { fx: -21, sfx: -26.5 };
function sectionRms([L, R], a, b) {
  let s2 = 0;
  const i0 = Math.round(a * SR), i1 = Math.round(b * SR);
  for (let i = i0; i < i1; i++) s2 += (L[i] * L[i] + R[i] * R[i]) / 2;
  return 10 * Math.log10(s2 / (i1 - i0) + 1e-12);
}
function gatedRms([L, R], gateDb = -45) {
  const win = Math.round(0.05 * SR);
  let tot = 0, n = 0;
  for (let s0 = 0; s0 + win < LEN; s0 += win) {
    let e = 0;
    for (let i = s0; i < s0 + win; i++) e += (L[i] * L[i] + R[i] * R[i]) / 2;
    e /= win;
    if (10 * Math.log10(e + 1e-12) > gateDb) { tot += e; n++; }
  }
  return 10 * Math.log10(tot / Math.max(1, n) + 1e-12);
}
const gains = {};
for (const [name, st] of Object.entries(stems)) {
  if (TARGET_RMS[name] !== undefined) gains[name] = db(TARGET_RMS[name] - sectionRms(st, 8, 24));
  else gains[name] = db(GATED[name] - gatedRms(st));
}
const mix = bus();
for (const [name, [L, R]] of Object.entries(stems)) {
  const g = gains[name];
  for (let i = 0; i < mix[0].length; i++) { mix[0][i] += L[i] * g; mix[1][i] += R[i] * g; }
}
// high-pass at 30 Hz
{ const hl = svf(), hr = svf(); for (let i = 0; i < mix[0].length; i++) { mix[0][i] = hl(mix[0][i], 30, 0.7, 2); mix[1][i] = hr(mix[1][i], 30, 0.7, 2); } }
// glue compressor
function compress([L, R], { thresh = -14, ratio = 2, attack = 0.01, release = 0.15, knee = 6 } = {}) {
  const aa = Math.exp(-1 / (attack * SR)), ar = Math.exp(-1 / (release * SR));
  let envDb = -120;
  for (let i = 0; i < L.length; i++) {
    const x = Math.max(Math.abs(L[i]), Math.abs(R[i]));
    const xd = 20 * Math.log10(x + 1e-9);
    envDb = xd > envDb ? aa * envDb + (1 - aa) * xd : ar * envDb + (1 - ar) * xd;
    const over = envDb - thresh;
    let gr = 0;
    if (over > knee / 2) gr = over * (1 - 1 / ratio);
    else if (over > -knee / 2) gr = ((over + knee / 2) ** 2 / (2 * knee)) * (1 - 1 / ratio);
    const g = db(-gr);
    L[i] *= g; R[i] *= g;
  }
}
// BS.1770 integrated loudness (K-weighting, 400 ms blocks, gating).
function lufs([L, R], n) {
  const kw = (x) => {
    const y = new Float64Array(n);
    let x1 = 0, x2 = 0, y1 = 0, y2 = 0;
    const b1 = [1.53512485958697, -2.69169618940638, 1.19839281085285], a1 = [-1.69065929318241, 0.73248077421585];
    const z = new Float64Array(n);
    for (let i = 0; i < n; i++) { const v = b1[0] * x[i] + b1[1] * x1 + b1[2] * x2 - a1[0] * y1 - a1[1] * y2; x2 = x1; x1 = x[i]; y2 = y1; y1 = v; z[i] = v; }
    x1 = x2 = y1 = y2 = 0;
    const a2 = [-1.99004745483398, 0.99007225036621];
    for (let i = 0; i < n; i++) { const v = z[i] - 2 * x1 + x2 - a2[0] * y1 - a2[1] * y2; x2 = x1; x1 = z[i]; y2 = y1; y1 = v; y[i] = v; }
    return y;
  };
  const kl = kw(L), kr = kw(R);
  const blk = Math.round(0.4 * SR), hop = Math.round(0.1 * SR);
  const z = [];
  for (let s = 0; s + blk <= n; s += hop) { let acc = 0; for (let i = s; i < s + blk; i++) acc += kl[i] * kl[i] + kr[i] * kr[i]; z.push(acc / blk); }
  const L_ = (m) => -0.691 + 10 * Math.log10(m + 1e-12);
  let gated = z.filter((m) => L_(m) > -70);
  const rel = L_(gated.reduce((a, b) => a + b, 0) / gated.length) - 10;
  gated = gated.filter((m) => L_(m) > rel);
  return L_(gated.reduce((a, b) => a + b, 0) / gated.length);
}
function limit([L, R], ceiling = db(-1.2), lookahead = 0.003, release = 0.08) {
  const la = Math.round(lookahead * SR), n = L.length;
  const req = new Float32Array(n);
  for (let i = 0; i < n; i++) { const p = Math.max(Math.abs(L[i]), Math.abs(R[i])); req[i] = p > ceiling ? ceiling / p : 1; }
  // sliding minimum over the lookahead window, then smooth release
  const g = new Float32Array(n);
  const ar = Math.exp(-1 / (release * SR));
  let cur = 1;
  const dq = [];
  for (let i = 0; i < n + la; i++) {
    if (i < n) { while (dq.length && req[dq[dq.length - 1]] >= req[i]) dq.pop(); dq.push(i); }
    const j = i - la;
    if (j < 0) continue;
    while (dq.length && dq[0] < j) dq.shift();
    const target = dq.length ? req[dq[0]] : 1;
    cur = target < cur ? target : ar * cur + (1 - ar) * target;
    g[j] = cur;
  }
  for (let i = 0; i < n; i++) { L[i] *= g[i]; R[i] *= g[i]; }
}

compress(mix, { thresh: -13, ratio: 2, attack: 0.012, release: 0.16 });
for (let i = 0; i < mix[0].length; i++) { mix[0][i] = sat(mix[0][i] * 0.9, 1.15) / 0.9; mix[1][i] = sat(mix[1][i] * 0.9, 1.15) / 0.9; }
const TARGET = -14;
let gain = 1;
for (let pass = 0; pass < 3; pass++) {
  const m = [mix[0].slice(0, LEN).map((v) => v * gain), mix[1].slice(0, LEN).map((v) => v * gain)];
  limit(m);
  const l = lufs(m, LEN);
  gain *= db(TARGET - l);
  if (Math.abs(TARGET - l) < 0.1) break;
}
const final = [mix[0].slice(0, LEN), mix[1].slice(0, LEN)];
for (const ch of final) for (let i = 0; i < ch.length; i++) ch[i] *= gain;
limit(final);
// the last half second fades so the file ends in silence
const f0 = Math.round(29.5 * SR);
for (const ch of final) for (let i = f0; i < LEN; i++) { const p = (i - f0) / (LEN - f0); ch[i] *= Math.cos((p * Math.PI) / 2) ** 1.5; }
const loud = lufs(final, LEN);
let peak = 0;
for (const ch of final) for (const v of ch) peak = Math.max(peak, Math.abs(v));

function writeWav(file, [L, R], n = L.length) {
  const bytes = 3;
  const data = Buffer.alloc(n * 2 * bytes);
  let o = 0;
  for (let i = 0; i < n; i++) for (const ch of [L, R]) {
    const v = Math.round(clamp(ch[i], -1, 1) * 8388607);
    data.writeIntLE(v, o, 3);
    o += 3;
  }
  const h = Buffer.alloc(44);
  h.write('RIFF', 0); h.writeUInt32LE(36 + data.length, 4); h.write('WAVE', 8); h.write('fmt ', 12);
  h.writeUInt32LE(16, 16); h.writeUInt16LE(1, 20); h.writeUInt16LE(2, 22); h.writeUInt32LE(SR, 24);
  h.writeUInt32LE(SR * 2 * bytes, 28); h.writeUInt16LE(2 * bytes, 32); h.writeUInt16LE(8 * bytes, 34);
  h.write('data', 36); h.writeUInt32LE(data.length, 40);
  fs.writeFileSync(file, Buffer.concat([h, data]));
}
fs.mkdirSync(out, { recursive: true });
writeWav(path.join(out, 'soundtrack.wav'), final, LEN);
if (process.argv.includes('--stems')) {
  fs.mkdirSync(path.join(out, 'stems'), { recursive: true });
  for (const [name, s] of Object.entries(stems)) writeWav(path.join(out, 'stems', `${name}.wav`), [s[0].map((v) => v * gains[name] * gain), s[1].map((v) => v * gains[name] * gain)], LEN);
}
console.log(Object.entries(gains).map(([k, g]) => `${k} ${(20 * Math.log10(g)).toFixed(1)} dB`).join(' · '));
console.log(`out/soundtrack.wav  ${DURATION}s  ${loud.toFixed(1)} LUFS  peak ${(20 * Math.log10(peak)).toFixed(2)} dBFS  ${ev.length} notes  ${cues.length} cues`);
