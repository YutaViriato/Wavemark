// 23.75–27.75 s. Climax: a tilted wall of real Wavemark sites (the docs and
// the sites blind agents wrote), each page cut into header, hero and body
// slices that build on 16ths. Kinetic words land on the beats, then the
// whole wall is pulled into the centre for the end card.
import { scene, h, set, css, clamp, lerp, ease, spring, cue, rng } from '../engine.mjs';

const COLS = 6, ROWS = 4, TW = 480, TH = 300, GAP = 36;
const PW = COLS * TW + (COLS - 1) * GAP, PH = ROWS * TH + (ROWS - 1) * GAP;
const WORDS = [
  { t: 24.5, text: 'Docs.' },
  { t: 25.0, text: 'Articles.' },
  { t: 25.5, text: 'Landing pages.' },
  { t: 26.0, text: 'All of it Markdown.', hold: true },
];

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#141413', zIndex: 1, perspective: '1700px', perspectiveOrigin: '50% 45%' } });
  cam.append(root);
  const plane = h('div', { class: 'abs', style: { left: `${960 - PW / 2}px`, top: `${540 - PH / 2}px`, width: `${PW}px`, height: `${PH}px`, transformStyle: 'preserve-3d', transformOrigin: '50% 50%' } });
  root.append(plane);

  const manifest = await fetch('./gen/shots/manifest.json').then((r) => r.json());
  const order = ['site', 'tidepool', 'blind2', 'estudio'];
  const shots = manifest.slice().sort((a, b) => order.indexOf(a.site) - order.indexOf(b.site));
  const r = rng(77);
  // Interleave light and dark, sites and pages.
  const pickList = [];
  const light = shots.filter((s) => s.scheme === 'light'), dark = shots.filter((s) => s.scheme === 'dark');
  for (let i = 0; pickList.length < COLS * ROWS; i++) pickList.push((i % 3 === 2 ? dark : light)[(i * 7) % (i % 3 === 2 ? dark.length : light.length)]);
  await Promise.all(pickList.map((s) => { const im = new Image(); im.src = `./gen/shots/${s.name}`; return im.decode(); }));

  const k = TW / 1440;
  const cuts = [[0, 84], [84, 560], [560, 900]];
  const tiles = [];
  for (let row = 0; row < ROWS; row++) for (let c = 0; c < COLS; c++) {
    const s = pickList[row * COLS + c];
    const x = c * (TW + GAP), y = row * (TH + GAP);
    const tile = h('div', { class: 'abs', style: { left: `${x}px`, top: `${y}px`, width: `${TW}px`, height: `${TH}px`, borderRadius: '12px', overflow: 'hidden', background: s.scheme === 'dark' ? '#141413' : '#faf9f5', boxShadow: '0 30px 60px -24px rgb(0 0 0 / .8), 0 0 0 1px rgb(255 255 255 / .07)', transformOrigin: '50% 50%' } });
    const bands = cuts.map(([a, b]) => {
      const el = h('div', { class: 'abs', style: { left: '0', top: `${a * k}px`, width: `${TW}px`, height: `${(b - a) * k + 0.5}px`, backgroundImage: `url(./gen/shots/${s.name})`, backgroundSize: `${TW}px ${TH}px`, backgroundPosition: `0 ${-a * k}px` } });
      tile.append(el);
      return el;
    });
    plane.append(tile);
    const dc = c - (COLS - 1) / 2, dr = row - (ROWS - 1) / 2;
    tiles.push({ tile, bands, cx: x + TW / 2 - PW / 2, cy: y + TH / 2 - PH / 2, d: Math.hypot(dc, dr * 1.3), spin: (r() - 0.5) * 50, delay: r() * 0.05 });
  }
  const maxD = Math.max(...tiles.map((t) => t.d));

  // Scrim and words
  const scrim = h('div', { class: 'fill', style: { background: 'radial-gradient(ellipse 58% 40% at 50% 50%, rgb(20 20 19 / .9), rgb(20 20 19 / .45) 60%, rgb(20 20 19 / .1) 100%)' } });
  root.append(scrim);
  const words = WORDS.map((w) => {
    const el = h('div', { class: 'abs', text: w.text, style: { left: '0', top: '450px', width: '1920px', textAlign: 'center', font: '400 168px/1 "Newsreader", serif', letterSpacing: '-0.03em', color: '#faf9f5', whiteSpace: 'nowrap', textShadow: '0 10px 60px rgb(0 0 0 / .5)' } });
    root.append(el);
    return { ...w, el };
  });
  const italic = words[3].el;
  italic.innerHTML = 'All of it <em style="font-style:italic;color:#e6a68d">Markdown.</em>';

  for (const w of WORDS) cue(w.t, 'slam', { vel: w.hold ? 0.9 : 0.7 });
  cue(27.0, 'whoosh', { dur: 0.75, vel: 0.9, tone: 0.6, rev: true });

  scene({
    t0: 23.75, t1: 27.8, root,
    update(t) {
      // camera: dolly across the tilted wall, drifting closer
      const u = clamp((t - 23.75) / 3.25);
      const col = ease.inExpo(clamp((t - 27.0) / 0.62));
      const tx = lerp(lerp(420, -380, ease.inOutSine(u)), 0, col), ty = lerp(lerp(120, -90, ease.inOutSine(u)), 0, col);
      const tz = lerp(lerp(-520, -80, ease.outCubic(u)), 0, col);
      const dim = clamp((t - 24.1) / 0.3) * (1 - clamp((t - 26.9) / 0.2));
      set(plane, { bright: 1 - 0.42 * dim });
      css(plane, 'transform', `translate3d(${tx.toFixed(1)}px,${ty.toFixed(1)}px,${tz.toFixed(1)}px) rotateX(${lerp(30, 22, u).toFixed(2)}deg) rotateZ(${(lerp(-13, -9, u) + col * 120).toFixed(2)}deg) scale(${(1 - 0.92 * col).toFixed(4)})`);
      set(root, { o: clamp((t - 23.76) / 0.2) });

      tiles.forEach((tl) => {
        const t0 = 23.8 + (tl.d / maxD) * 0.55 + tl.delay;
        tl.bands.forEach((b, i) => {
          const lt = t - t0 - i * 0.0625;
          const p = spring(lt, 2.8, 0.62);
          set(b, { y: (1 - p) * (i === 0 ? -50 : 70 + i * 20), o: clamp(lt / 0.06) });
        });
        // pulled into the centre
        set(tl.tile, { x: -tl.cx * col, y: -tl.cy * col, r: tl.spin * col, s: 1 - 0.6 * col, o: 1 - clamp((col - 0.85) / 0.15) });
      });

      // words land on the beats with a slot-machine roll
      words.forEach((w, i) => {
        const next = words[i + 1];
        const lt = t - w.t;
        const inP = spring(lt, 2.6, 0.62);
        const outP = next ? ease.inCubic(clamp((t - next.t + 0.06) / 0.14)) : ease.inCubic(clamp((t - 26.9) / 0.18));
        set(w.el, { y: (1 - inP) * 120 - outP * 140, s: 0.9 + 0.1 * inP + (w.hold ? 0.04 * clamp(lt / 1.2) : 0), o: clamp(lt / 0.06) * (1 - outP), blur: (1 - clamp(lt / 0.16)) * 16 + outP * 14 });
      });
      set(scrim, { o: clamp((t - 24.1) / 0.3) * (1 - col) });
    },
  });
}
