// Entry point: loads the product assets, builds every scene and exposes the
// seek API the renderer drives (window.__seek, window.__cues).
import { renderAt, scenes, cues, set, ease, clamp, pulses, noise, DURATION } from './engine.mjs';
import { loadAssets } from './kit.mjs';
import { kicks } from './score.mjs';

const sceneModules = ['open', 'drop', 'ch1', 'ch2', 'ch3', 'ch4', 'ch5', 'ch6', 'ch7', 'wall', 'end', 'hud'];

async function boot() {
  const [art, ...frags] = await Promise.all([
    fetch('./gen/art.json').then((r) => r.json()),
    ...['landing', 'docs', 'components', 'browse'].map((n) => fetch(`./gen/frag/${n}.html`).then((r) => r.text())),
  ]);
  loadAssets(art, { landing: frags[0], docs: frags[1], components: frags[2], browse: frags[3] });

  const faces = ['400 64px Newsreader', 'italic 400 64px Newsreader', '500 64px Newsreader', '400 20px "Hanken Grotesk"', '500 20px "Hanken Grotesk"', '600 20px "Hanken Grotesk"', '700 20px "Hanken Grotesk"', '400 20px "JetBrains Mono"', '500 20px "JetBrains Mono"'];
  await Promise.all(faces.map((f) => document.fonts.load(f, 'Wavemark ÁÉÍÓÚáéíóúãõç—·→✓')));
  await document.fonts.ready;

  const stage = document.getElementById('stage');
  const cam = document.getElementById('cam');
  const fx = document.getElementById('fx');
  const ctx = { stage, cam, fx };
  for (const name of sceneModules) {
    const mod = await import(`./scenes/${name}.mjs`);
    await mod.build(ctx);
  }
  for (const s of scenes) if (s.root) s.root.style.display = 'none';

  // Camera shake on impacts and a soft punch on the downbeats of the drops.
  const impacts = [[8, 16], [14.5, 5], [24, 10], [28, 12]];
  const downbeats = kicks().filter((k) => (k >= 8 && k < 24 && k % 2 === 0) || (k >= 24 && k < 27 && k % 1 === 0));
  ctx.camera = (t) => {
    let sx = 0, sy = 0, rot = 0;
    for (const [t0, amp] of impacts) {
      const d = t - t0;
      if (d < 0 || d > 0.5) continue;
      const a = amp * Math.exp(-d * 9);
      sx += noise(d * 38, t0) * a;
      sy += noise(d * 41, t0 + 9) * a;
      rot += noise(d * 30, t0 + 3) * a * 0.03;
    }
    const punch = 1 + 0.006 * pulses(t, downbeats, 0.12);
    const push = 1 + 0.035 * clamp((t - 10) / 14) - 0.035 * clamp((t - 23.6) / 0.4);
    const hand = clamp((t - 10) / 0.5) * (1 - clamp((t - 23.6) / 0.4));
    sx += noise(t * 0.55, 1) * 7 * hand;
    sy += noise(t * 0.48, 2) * 5 * hand;
    rot += noise(t * 0.4, 3) * 0.12 * hand;
    set(cam, { x: sx, y: sy, r: rot, s: punch * push });
  };

  window.__seek = (t) => {
    renderAt(t);
    ctx.camera(t);
  };
  window.__cues = () => cues;
  window.__duration = DURATION;
  window.__ready = true;
}

boot().catch((e) => {
  window.__error = String(e && e.stack ? e.stack : e);
  console.error(e);
});
