// The score: 120 BPM, 4/4, D major, 15 bars = exactly 30 seconds.
// Pure data, imported by the stage (beat-reactive visuals) and by the
// synthesizer (audio/synth.mjs), so picture and sound share one grid.
//
//  bar  time      section   harmony
//   1   0–2       intro     Bm9        typing "# Markdown"
//   2   2–4       intro     Gmaj9      the tagline, one pluck per word
//   3   4–6       build     Em9        kick enters, camera pulls back
//   4   6–8       build     F#m7       page assembles, roll, cut at 7.75
//   5   8–10      drop A    Gmaj7      brand: the sonic logo A–D–F#
//   6–8 10–16     drop A    A6 F#m7 Bm7 …
//   9–12 16–24    drop B    royal road again, arp added
//  13–14 24–28    climax    Gmaj7 A7sus4→A7, cut at 27.75
//  15   28–30     end       D(add9)    sonic logo resolves on the tonic

export const BPM = 120;
export const BEAT = 0.5;
export const BAR = 2;
export const BARS = 15;
export const DURATION = 30;

export const SECTIONS = {
  intro: [0, 4],
  build: [4, 7.75],
  gap1: [7.75, 8],
  dropA: [8, 16],
  dropB: [16, 24],
  climax: [24, 27.75],
  gap2: [27.75, 28],
  end: [28, 30],
};

// MIDI note numbers.
const N = { D1: 26, E1: 28, Fs1: 30, G1: 31, A1: 33, B1: 35, D2: 38, E2: 40, Fs2: 42, G2: 43, A2: 45, B2: 47, Cs3: 49, D3: 50, E3: 52, Fs3: 54, G3: 55, A3: 57, B3: 59, Cs4: 61, D4: 62, E4: 64, Fs4: 66, G4: 67, A4: 69, B4: 71, Cs5: 73, D5: 74, E5: 76, Fs5: 78, G5: 79, A5: 81, B5: 83, Cs6: 85, D6: 86, E6: 88, Fs6: 90, G6: 91 };
export const NOTES = N;

// One chord per bar (bar 14 splits sus4 → dominant). Pad voicings move by step.
export const CHORDS = [
  { name: 'Bm9', root: N.B1, pad: [N.B2, N.D3, N.A3, N.Cs4, N.Fs4] },
  { name: 'Gmaj9', root: N.G1, pad: [N.G2, N.D3, N.A3, N.B3, N.Fs4] },
  { name: 'Em9', root: N.E1, pad: [N.E2, N.B2, N.D3, N.G3, N.Fs4] },
  { name: 'F#m7', root: N.Fs1, pad: [N.Fs2, N.Cs3, N.E3, N.A3, N.Cs4] },
  { name: 'Gmaj7', root: N.G1, pad: [N.G2, N.D3, N.Fs3, N.B3, N.D4] },
  { name: 'A6', root: N.A1, pad: [N.A2, N.E3, N.Fs3, N.Cs4, N.E4] },
  { name: 'F#m7', root: N.Fs1, pad: [N.Fs2, N.Cs3, N.E3, N.A3, N.Cs4] },
  { name: 'Bm7', root: N.B1, pad: [N.B2, N.D3, N.Fs3, N.A3, N.D4] },
  { name: 'Gmaj7', root: N.G1, pad: [N.G2, N.D3, N.Fs3, N.B3, N.D4] },
  { name: 'A6', root: N.A1, pad: [N.A2, N.E3, N.Fs3, N.Cs4, N.E4] },
  { name: 'F#m7', root: N.Fs1, pad: [N.Fs2, N.Cs3, N.E3, N.A3, N.Cs4] },
  { name: 'Bm7', root: N.B1, pad: [N.B2, N.D3, N.Fs3, N.A3, N.D4] },
  { name: 'Gmaj7', root: N.G1, pad: [N.G2, N.D3, N.Fs3, N.B3, N.D4] },
  { name: 'A7sus4', root: N.A1, pad: [N.A2, N.E3, N.G3, N.A3, N.D4], then: { at: 2, name: 'A7', pad: [N.A2, N.E3, N.G3, N.A3, N.Cs4] } },
  { name: 'Dadd9', root: N.D1, pad: [N.D2, N.A2, N.Fs3, N.A3, N.E4] },
];

// Arpeggio cells (16ths), one per chord.
const ARP = {
  Em9: [N.E4, N.G4, N.B4, N.D5, N.Fs5, N.D5, N.B4, N.G4],
  'F#m7': [N.Fs4, N.A4, N.Cs5, N.E5, N.A5, N.E5, N.Cs5, N.A4],
  Gmaj7: [N.G4, N.B4, N.D5, N.Fs5, N.A5, N.Fs5, N.D5, N.B4],
  A6: [N.A4, N.Cs5, N.E5, N.Fs5, N.A5, N.Fs5, N.E5, N.Cs5],
  Bm7: [N.B4, N.D5, N.Fs5, N.A5, N.B5, N.A5, N.Fs5, N.D5],
  A7sus4: [N.A4, N.D5, N.E5, N.G5, N.A5, N.G5, N.E5, N.D5],
};

// The hook. Its first three notes, A–D–F#, are the sonic logo: over Gmaj7 at
// the brand reveal they sound open (9, 5, 7); over D at the end card they land
// home (5, 1, 3). [beat in bar, note, length in beats]
const HOOK = [
  [[0, N.A5, 0.5], [0.5, N.D6, 0.5], [1, N.Fs6, 1.5], [2.5, N.E6, 0.5], [3, N.D6, 1]],
  [[0, N.Cs6, 1.5], [1.5, N.B5, 0.5], [2, N.A5, 2]],
  [[0, N.A5, 0.5], [0.5, N.Cs6, 0.5], [1, N.E6, 1.5], [2.5, N.D6, 0.5], [3, N.Cs6, 1]],
  [[0, N.B5, 1.5], [1.5, N.A5, 0.5], [2, N.Fs5, 2]],
];
const CLIMAX = [
  [[0, N.A5, 0.5], [0.5, N.D6, 0.5], [1, N.Fs6, 1.5], [2.5, N.G6, 0.5], [3, N.Fs6, 1]],
  [[0, N.E6, 1.5], [1.5, N.D6, 0.5], [2, N.Cs6, 1.5]],
];
export const LOGO = [[0, N.A5, 0.5], [0.5, N.D6, 0.5], [1, N.Fs6, 3]];

// The words of the tagline get one pluck each (bar 2), rising through Gmaj9.
export const TAGLINE_NOTES = [
  { t: 2.0, note: N.D5 }, // Markdown (morph)
  { t: 2.25, note: N.Fs5 }, // that
  { t: 2.5, note: N.A5 }, // becomes
  { t: 3.0, note: N.B5 }, // a site
  { t: 3.25, note: N.D6 }, // while
  { t: 3.375, note: N.E6 }, // you
  { t: 3.5, note: N.Fs6 }, // write
];

const barT = (b) => b * BAR;
const inGap = (t) => (t >= 7.75 && t < 8) || (t >= 27.75 && t < 28);

// ---------------------------------------------------------------- events
export function events() {
  const ev = [];
  const add = (inst, t, props = {}) => { if (!inGap(t) || inst === 'reverse' || inst === 'riser') ev.push({ inst, t, ...props }); };

  // Kick: build and drops, four on the floor. Air before the climax.
  for (let b = 2; b < 14; b++) for (let k = 0; k < 4; k++) {
    const t = barT(b) + k * BEAT;
    if (t >= 27 && t < 28) continue; // the wall collapses with no drums
    add('kick', t, { vel: b < 4 ? 0.82 + 0.06 * (b - 2) + k * 0.01 : 1, lp: b < 4 ? 260 * Math.pow(70, (t - 4) / 3.75) : 20000 });
  }
  add('kick', 28, { vel: 1.05, lp: 20000 });

  // Clap on 2 and 4 from the last build bar.
  for (let b = 3; b < 14; b++) for (const k of [1, 3]) {
    const t = barT(b) + k * BEAT;
    if (t >= 27) continue;
    add('clap', t, { vel: b === 3 ? 0.7 : 0.9 });
  }

  // Hats.
  for (let b = 1; b < 14; b++) for (let s = 0; s < 16; s++) {
    const t = barT(b) + s * 0.125;
    if (t >= 27) continue;
    const off8 = s % 4 === 2;
    if (b === 1) { if (s % 2 === 0) add('hat', t, { vel: 0.18 + (off8 ? 0.1 : 0) }); continue; }
    if (b === 2) { if (s % 2 === 0) add('hat', t, { vel: off8 ? 0.5 : 0.25 }); continue; }
    if (b === 3) { add('hat', t, { vel: 0.22 + 0.25 * (s / 16) + (off8 ? 0.2 : 0) }); continue; }
    if (off8) add('ohat', t, { vel: b >= 8 ? 0.62 : 0.55 });
    else add('hat', t, { vel: s % 4 === 0 ? 0.42 : 0.3 + (s % 2) * 0.04 });
    if (b >= 8 && s % 2 === 1) add('shaker', t, { vel: 0.34 });
  }

  // Snare rolls into the first drop and into the climax; a short fill at 15.5.
  const roll = (from, to, startVel) => {
    let t = from;
    while (t < to - 1e-6) {
      const p = (t - from) / (to - from);
      const step = p < 0.5 ? 0.25 : p < 0.8 ? 0.125 : 0.0625;
      add('snare', t, { vel: startVel + (1 - startVel) * p, pitch: 1 + 0.35 * p });
      t += step;
    }
  };
  roll(6, 7.75, 0.25);
  roll(22, 23.875, 0.2);
  for (const [i, t] of [15.5, 15.625, 15.75, 15.875].entries()) add('snare', t, { vel: 0.5 + i * 0.15, pitch: 1 + i * 0.05 });

  // Cymbals, impacts and transitions.
  add('reverse', 3, { dur: 1, vel: 0.5 });
  add('riser', 4, { dur: 3.75, vel: 0.85 });
  add('reverse', 7.5, { dur: 0.5, vel: 0.8 });
  add('impact', 8, { vel: 1 });
  add('crash', 8, { vel: 1 });
  add('downlifter', 8, { dur: 1.8, vel: 0.5 });
  add('crash', 16, { vel: 0.85 });
  add('riser', 22, { dur: 1.875, vel: 0.7 });
  add('impact', 24, { vel: 0.75 });
  add('crash', 24, { vel: 0.95 });
  add('crash', 26, { vel: 0.6 });
  add('riser', 26.5, { dur: 1.25, vel: 0.8, suck: true });
  add('reverse', 27.25, { dur: 0.5, vel: 0.9 });
  add('impact', 28, { vel: 1.1 });
  add('crash', 28, { vel: 1 });

  // Bass.
  for (let b = 2; b < 14; b++) {
    const ch = CHORDS[b];
    if (b < 4) {
      for (let e = 0; e < 8; e++) add('bass', barT(b) + e * BEAT / 2, { note: ch.root + 12, dur: 0.2, vel: 0.55 + 0.05 * e / 8 + (b - 2) * 0.1, cutoff: 300 + (b - 2) * 400 + e * 60 });
      continue;
    }
    add('sub', barT(b), { note: ch.root, dur: BAR, vel: 0.9 });
    for (let k = 0; k < 4; k++) {
      const t = barT(b) + k * BEAT + BEAT / 2;
      if (t >= 27 || (t > 23.8 && t < 24)) continue;
      const jump = b % 2 === 1 && k === 3;
      add('bass', t, { note: ch.root + (jump ? 24 : 12), dur: 0.2, vel: 0.85, cutoff: 900 });
    }
  }
  add('sub', 28, { note: N.D1, dur: 2, vel: 1, final: true });
  add('bass', 28, { note: N.D2, dur: 1.2, vel: 0.9, cutoff: 600 });

  // Pads, one chord per bar.
  for (let b = 0; b < 15; b++) {
    const ch = CHORDS[b];
    if (b === 13) {
      add('pad', barT(b), { notes: ch.pad, dur: 1, vel: 0.9, cutoff: 5200 });
      add('pad', barT(b) + 1, { notes: ch.then.pad, dur: 0.75, vel: 0.9, cutoff: 5200, sweepDown: true });
      continue;
    }
    if (b === 14) { add('pad', 28, { notes: ch.pad, dur: 0.9, vel: 1, cutoff: 6000, final: true }); continue; }
    const cutoff = b < 2 ? 900 + b * 300 : b < 4 ? 1400 + (b - 2) * 1400 : 5000;
    add('pad', barT(b), { notes: ch.pad, dur: BAR, vel: b < 2 ? 0.7 : 0.85, cutoff });
  }

  // Arpeggio: build bars and drop B and climax.
  for (let b = 2; b < 14; b++) {
    if (b >= 4 && b < 8) continue;
    const ch = CHORDS[b];
    const cell = ARP[ch.name];
    for (let s = 0; s < 16; s++) {
      const t = barT(b) + s * 0.125;
      if (t >= 27) continue;
      const vel = (b < 4 ? 0.35 + 0.25 * ((b - 2) * 16 + s) / 32 : 0.5) * (s % 4 === 0 ? 1 : 0.8);
      add('arp', t, { note: cell[s % 8] + (b >= 8 ? 12 : 0), dur: 0.12, vel, cutoff: b < 4 ? 1200 + 3800 * ((b - 2) * 16 + s) / 32 : 4200 });
    }
  }

  // Tagline plucks.
  for (const n of TAGLINE_NOTES) add('pluck', n.t, { note: n.note, dur: 1.2, vel: 0.75 });

  // Hook (bell) in both drops and the climax, the logo at the end.
  for (let b = 4; b < 12; b++) for (const [p, note, d] of HOOK[(b - 4) % 4]) {
    const t = barT(b) + p * BEAT;
    add('bell', t, { note, dur: d * BEAT, vel: b === 4 && p < 1.5 ? 1 : 0.8, double: b >= 8 });
  }
  for (let b = 12; b < 14; b++) for (const [p, note, d] of CLIMAX[b - 12]) {
    const t = barT(b) + p * BEAT;
    if (t >= 27.75) continue;
    add('bell', t, { note, dur: Math.min(d * BEAT, 27.75 - t), vel: 0.9, double: true });
  }
  for (const [p, note, d] of LOGO) add('bell', 28 + p * BEAT, { note, dur: d * BEAT, vel: 1, double: true, final: true });

  return ev.sort((a, b) => a.t - b.t);
}

// Kick times, for sidechain ducking and beat-reactive visuals.
export const kicks = () => events().filter((e) => e.inst === 'kick').map((e) => e.t);
export const claps = () => events().filter((e) => e.inst === 'clap').map((e) => e.t);
