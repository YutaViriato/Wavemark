// 10–24 s. A quiet HUD over the chapters: the wordmark top left and seven
// progress ticks bottom left. Its ink flips with the background of each
// chapter (ivory, dark, clay).
import { scene, h, set, css, clamp, ease, spring } from '../engine.mjs';

const CH = [10, 12, 14, 16, 18, 20, 22, 24];
// [from, ink] — the colour the HUD should use from that time on
const INK = [[10, '#141413'], [14.02, '#faf9f5'], [15.9, '#141413'], [21.78, '#faf9f5']];

export async function build({ fx }) {
  const root = h('div', { class: 'sc', style: { pointerEvents: 'none', background: 'transparent' } });
  fx.prepend(root);
  const mark = h('div', { class: 'abs', text: 'Wavemark', style: { left: '120px', top: '62px', font: '500 30px/1 "Newsreader", serif', letterSpacing: '-0.01em' } });
  const ticks = h('div', { class: 'abs', style: { left: '120px', top: '1006px', display: 'flex', gap: '8px' } });
  const segs = CH.slice(0, 7).map(() => {
    const seg = h('div', { style: { width: '46px', height: '3px', borderRadius: '2px', position: 'relative', overflow: 'hidden' } });
    const bg = h('div', { class: 'fill', style: { background: 'currentColor', opacity: 0.22 } });
    const fill = h('div', { class: 'fill', style: { background: 'currentColor', transformOrigin: '0 50%' } });
    seg.append(bg, fill);
    ticks.append(seg);
    return fill;
  });
  const count = h('div', { class: 'abs sr-mono', style: { left: '520px', top: '994px', height: '22px', overflow: 'hidden', font: '500 18px/22px "JetBrains Mono", monospace', letterSpacing: '0.06em', display: 'flex', whiteSpace: 'pre' } });
  const roll = h('div', { style: { display: 'flex', flexDirection: 'column' } }, ...[1, 2, 3, 4, 5, 6, 7].map((n) => h('div', { text: `0${n}` })));
  count.append(h('div', { style: { height: '22px', overflow: 'hidden' } }, roll), h('span', { text: ' / 07' }));
  root.append(mark, ticks, count);

  scene({
    t0: 10.0, t1: 24.0, root,
    update(t) {
      let ink = INK[0][1];
      for (const [from, c] of INK) if (t >= from) ink = c;
      for (const el of [mark, ticks, count]) css(el, 'color', ink);
      const vis = ease.outCubic(clamp((t - 10.1) / 0.3)) * (1 - clamp((t - 23.75) / 0.2));
      set(root, { o: vis * 0.62 });
      segs.forEach((f, i) => set(f, { sx: ease.inOutSine(clamp((t - CH[i]) / 2)) }));
      let pos = 0;
      for (let i = 1; i < 7; i++) pos += spring(t - CH[i] - 0.02, 3.2, 0.6);
      set(roll, { y: -22 * pos });
    },
  });
}
