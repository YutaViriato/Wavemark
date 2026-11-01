// Objective checks of the soundtrack, since the mix is made without ears:
// level of each stem per section, spectral balance of the mix per octave,
// stereo correlation, and onset times compared with the score.
//
//   node audio/analyze.mjs
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { events } from '../stage/score.mjs';

const here = path.dirname(fileURLToPath(import.meta.url));
const out = path.join(here, '../out');

function readWav(file) {
  const b = fs.readFileSync(file);
  let p = 12, fmt, data;
  while (p < b.length) {
    const id = b.toString('ascii', p, p + 4), len = b.readUInt32LE(p + 4);
    if (id === 'fmt ') fmt = { ch: b.readUInt16LE(p + 10), sr: b.readUInt32LE(p + 12), bits: b.readUInt16LE(p + 22) };
    if (id === 'data') data = b.subarray(p + 8, p + 8 + len);
    p += 8 + len;
  }
  const n = data.length / (fmt.bits / 8) / fmt.ch;
  const L = new Float32Array(n), R = new Float32Array(n);
  for (let i = 0; i < n; i++) {
    L[i] = data.readIntLE(i * 6, 3) / 8388608;
    R[i] = data.readIntLE(i * 6 + 3, 3) / 8388608;
  }
  return { sr: fmt.sr, L, R };
}
const SECTIONS = [['intro', 0, 4], ['build', 4, 7.75], ['dropA', 8, 16], ['dropB', 16, 24], ['climax', 24, 27.75], ['end', 28, 30]];
const rmsDb = ({ L, R, sr }, a, b) => {
  let s = 0;
  const i0 = Math.round(a * sr), i1 = Math.round(b * sr);
  for (let i = i0; i < i1; i++) s += (L[i] * L[i] + R[i] * R[i]) / 2;
  return 10 * Math.log10(s / (i1 - i0) + 1e-12);
};
const mix = readWav(path.join(out, 'soundtrack.wav'));
const rows = [['mix', mix]];
const stemDir = path.join(out, 'stems');
if (fs.existsSync(stemDir)) for (const f of fs.readdirSync(stemDir).sort()) rows.push([f.replace('.wav', ''), readWav(path.join(stemDir, f))]);
console.log('RMS dBFS'.padEnd(10) + SECTIONS.map(([n]) => n.padStart(8)).join(''));
for (const [name, w] of rows) console.log(name.padEnd(10) + SECTIONS.map(([, a, b]) => rmsDb(w, a, b).toFixed(1).padStart(8)).join(''));

// Octave bands of the mix in the drop (biquad band-passes).
function bandDb(w, fc, a, b) {
  const { sr } = w;
  const q = 1.414, w0 = (2 * Math.PI * fc) / sr, al = Math.sin(w0) / (2 * q);
  const b0 = al, b2 = -al, a0 = 1 + al, a1 = -2 * Math.cos(w0), a2 = 1 - al;
  let x1 = 0, x2 = 0, y1 = 0, y2 = 0, s = 0;
  const i0 = Math.round(a * sr), i1 = Math.round(b * sr);
  for (let i = i0; i < i1; i++) {
    const x = (w.L[i] + w.R[i]) / 2;
    const y = (b0 * x + b2 * x2 - a1 * y1 - a2 * y2) / a0;
    x2 = x1; x1 = x; y2 = y1; y1 = y;
    s += y * y;
  }
  return 10 * Math.log10(s / (i1 - i0) + 1e-12);
}
const bands = [31, 63, 125, 250, 500, 1000, 2000, 4000, 8000, 16000];
console.log('\nmix octave bands (dB), drop 8–24 s');
console.log(bands.map((f) => String(f).padStart(7)).join(''));
console.log(bands.map((f) => bandDb(mix, f, 8, 24).toFixed(1).padStart(7)).join(''));
console.log(bands.map((f) => bandDb(mix, f, 0, 4).toFixed(1).padStart(7)).join('') + '   (intro)');

// Stereo correlation of the mix.
let sLR = 0, sLL = 0, sRR = 0;
for (let i = 0; i < mix.L.length; i++) { sLR += mix.L[i] * mix.R[i]; sLL += mix.L[i] ** 2; sRR += mix.R[i] ** 2; }
console.log(`\nstereo correlation ${(sLR / Math.sqrt(sLL * sRR)).toFixed(3)}`);
let peak = 0, dc = 0;
for (let i = 0; i < mix.L.length; i++) { peak = Math.max(peak, Math.abs(mix.L[i]), Math.abs(mix.R[i])); dc += mix.L[i] + mix.R[i]; }
console.log(`peak ${(20 * Math.log10(peak)).toFixed(2)} dBFS, dc ${(dc / mix.L.length / 2).toExponential(2)}`);

// Kick onsets in the drums stem vs the score.
if (fs.existsSync(path.join(stemDir, 'kick.wav'))) {
  const d = readWav(path.join(stemDir, 'kick.wav'));
  const kicksT = events().filter((e) => e.inst === 'kick').map((e) => e.t);
  let worst = 0;
  for (const t of kicksT) {
    const i0 = Math.round((t - 0.01) * d.sr), i1 = Math.round((t + 0.03) * d.sr);
    let best = i0, m = 0;
    for (let i = i0; i < i1; i++) { const v = Math.abs(d.L[i]); if (v > m) { m = v; best = i; } }
    // onset = first sample above 20% of the local peak
    let on = i0;
    for (let i = i0; i <= best; i++) if (Math.abs(d.L[i]) > 0.2 * m) { on = i; break; }
    worst = Math.max(worst, Math.abs(on / d.sr - t));
  }
  console.log(`kick onsets vs score: worst offset ${(worst * 1000).toFixed(2)} ms over ${kicksT.length} kicks`);
}
