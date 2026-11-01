// 10–12 s. 01 Install: the real one-line installer types itself on 16ths, the
// real installer output prints, a "1 binary" sticker slaps onto the window.
// Exit: the terminal morphs into the editor of chapter 2.
import { scene, h, set, css, clamp, lerp, ease, spring, cue, panX } from '../engine.mjs';
import { headline, wordsIn, wordsOut, label, typeIn, codeWindow, pop, clipLeftOf, icon, drawIcon } from '../kit.mjs';
import { WAVE } from '../trans.mjs';
import { CH2_EDITOR } from './ch2.mjs';

const TERM = { x: 920, y: 330, w: 880, h: 430 };
const CMD = '$ curl -fsSL https://wavemark.viriato.site/install.sh | sh';
const CHUNKS = [[2, 6], [6, 12], [12, 29], [29, 42], [42, 53], [53, 58]];
const OUT = [
  'wavemark: downloading wavemark_darwin_arm64.tar.gz (latest)',
  'wavemark: installed wavemark v0.1.0 to ~/.local/bin/wavemark',
  'wavemark: next: wavemark init my-site && wavemark dev my-site',
];

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#faf9f5' } });
  cam.append(root);

  const lab = label('01 — Install');
  const hl = headline(['One line', 'to install.']);
  const head = h('div', { class: 'abs', style: { left: '120px', top: '300px' } }, lab, h('div', { style: { height: '34px' } }), hl.el);
  root.append(head);

  const pills = ['Linux', 'macOS', 'Windows'].map((p) => h('div', { class: 'sr-pill', text: p }));
  const pillRow = h('div', { class: 'abs', style: { left: '120px', top: '684px', display: 'flex', gap: '14px' } }, ...pills);
  root.append(pillRow);

  // Terminal (a dark code window without line numbers), tilted in 3D.
  const term = codeWindow({ x: 0, y: 0, w: TERM.w, h: TERM.h, tab: 'zsh — 80×24', lines: [CMD, ...OUT, '$ '], numbers: false, font: 22, lh: 44, dot: false });
  term.body.style.padding = '26px 0';
  for (const [i, r] of term.rows.entries()) if (i >= 1 && i <= 3) r.code.innerHTML = `<span class="tk-dim">wavemark:</span>${r.src.slice(9).replace(/&/g, '&amp;')}`;
  const tilt = h('div', { class: 'abs', style: { left: `${TERM.x}px`, top: `${TERM.y}px`, width: `${TERM.w}px`, height: `${TERM.h}px`, transformOrigin: '50% 50%' } }, term.el);
  term.el.style.left = '0';
  term.el.style.top = '0';
  root.append(tilt);
  const caret = h('div', { class: 'abs', style: { width: '12px', height: '26px', borderRadius: '2px', background: '#d97757' } });
  term.body.style.position = 'relative';
  term.body.append(caret);

  // Sticker: the product "box" icon and the two facts from the landing stats.
  const box = icon('box', { size: 74, color: '#141413', fill: '#e3dacc' });
  const sticker = h('div', { class: 'abs', style: { left: '1520px', top: '232px', width: '300px', padding: '18px 22px', display: 'flex', alignItems: 'center', gap: '16px', background: '#fff', borderRadius: '20px', border: '1px solid #e8e6dc', boxShadow: '0 18px 40px -16px rgb(20 20 19 / .35)', transformOrigin: '50% 50%' } },
    box.el,
    h('div', {}, h('div', { text: '1 binary', style: { font: '500 30px/1.1 "Newsreader", serif', color: '#141413' } }), h('div', { text: '0 dependencies', style: { font: '400 19px/1.4 "Hanken Grotesk", sans-serif', color: '#5e5d59', marginTop: '4px' } })));
  root.append(sticker);

  // ---- cues
  CHUNKS.forEach((c, i) => cue(10.25 + i * 0.125, 'key', { pan: panX(TERM.x + 30 + c[1] * 12.6), vel: 0.75, n: 2 }));
  cue(11.0, 'enter', { pan: 0.5, vel: 0.9 });
  cue(11.5, 'thump', { pan: 0.7, vel: 0.9 });
  [10.5, 10.625, 10.75].forEach((t, i) => cue(t, 'pop', { note: [74, 78, 81][i], pan: -0.6 + i * 0.12, vel: 0.5 }));
  cue(10.02, 'whoosh', { dur: 0.45, pan: 0.6, vel: 0.45, tone: 1.1, from: 0.9, to: 0.4 });
  cue(11.78, 'whoosh', { dur: 0.4, pan: -0.2, vel: 0.5, tone: 1.0, from: 0.6, to: -0.4 });

  const CMD_T = 10.25, STEP = 0.125;
  scene({
    t0: 9.85, t1: 12.02, root,
    update(t) {
      css(root, 'clip-path', t < 10.4 ? clipLeftOf(WAVE.trail(t), { phase: WAVE.phase(t) + 0.6 * 0, amp: 30, wl: 420, lean: -0.05 }) : 'none');

      typeIn(lab, t, 9.98, 0.02);
      wordsIn(hl.words, t, 10.0, { stagger: 0.06, em: 128 });
      pills.forEach((p, i) => pop(p, t, 10.5 + i * 0.125, { f: 3.6, z: 0.45, from: 0.3, dy: 20 }));

      // exit
      const ex = clamp((t - 11.76) / 0.26);
      if (ex > 0) {
        wordsOut(hl.words, t, 11.76, { stagger: 0.03, dur: 0.22, dy: -0.35 });
        set(lab, { o: 1 - ex });
        pills.forEach((p, i) => { const e = ease.inCubic(clamp((t - 11.76 - i * 0.03) / 0.2)); set(p, { y: e * 60, o: 1 - e, s: 1 }); });
      }

      // terminal entrance, tilt and morph into the chapter 2 editor
      const ti = spring(t - 10.0, 1.9, 0.72);
      const m = ease.inOutCubic(clamp((t - 11.76) / 0.26));
      const x = lerp(TERM.x + (1 - ti) * 700, CH2_EDITOR.x, m), y = lerp(TERM.y + (1 - ti) * 80, CH2_EDITOR.y, m);
      const w = lerp(TERM.w, CH2_EDITOR.w, m), hh = lerp(TERM.h, CH2_EDITOR.h, m);
      css(tilt, 'left', `${x.toFixed(1)}px`); css(tilt, 'top', `${y.toFixed(1)}px`);
      css(tilt, 'width', `${w.toFixed(1)}px`); css(tilt, 'height', `${hh.toFixed(1)}px`);
      css(term.el, 'width', `${w.toFixed(1)}px`); css(term.el, 'height', `${hh.toFixed(1)}px`);
      set(tilt, { persp: 2400, ry: lerp(-14 * (1 - ti) - 7, 0, m), rx: lerp(4, 0, m), r: lerp((1 - ti) * 6, 0, m), o: t < 10.0 ? 0 : 1 });
      set(term.body, { o: 1 - clamp((t - 11.76) / 0.12) });

      // command, chunk by chunk on 16ths
      let shown = 2;
      CHUNKS.forEach(([a, b], i) => { const t0 = CMD_T + i * STEP; if (t >= t0) shown = Math.max(shown, Math.min(b, a + Math.floor((t - t0) / 0.008) + 1)); });
      const r0 = term.rows[0];
      set(r0.row, { o: t >= 10.12 ? 1 : 0 });
      css(r0.code, 'clip-path', shown >= CMD.length ? 'none' : `inset(-4px calc(100% - ${shown * term.cw}px) -4px -2px)`);
      [11.0, 11.5, 11.625, 11.75].forEach((t0, k) => {
        const r = term.rows[k + 1];
        const lt = t - t0;
        set(r.row, { o: clamp(lt / 0.05), x: (1 - ease.outExpo(clamp(lt / 0.3))) * -24 });
      });
      const line = t < 11.0 ? 0 : t < 11.5 ? 1 : t < 11.625 ? 2 : t < 11.75 ? 3 : 4;
      const col = line === 0 ? shown : line === 4 ? 2 : OUT[line - 1].length;
      const blink = line === 0 && t > 10.9 ? 1 : (t * 2) % 1 < 0.6 ? 1 : 0.15;
      set(caret, { x: 28 + col * term.cw + 2, y: 26 + line * 44 + 9, o: (t >= 10.12 ? 1 : 0) * blink });

      // a small bump of the window on Enter and on each printed line
      const bump = [11.0, 11.5].reduce((acc, t0) => acc + (t >= t0 ? Math.exp(-(t - t0) * 14) : 0), 0);
      set(term.el, { s: 1 + 0.012 * bump });

      // sticker
      const sp = pop(sticker, t, 11.5, { f: 2.6, z: 0.42, from: 1.8, rot: -34, squash: 0.25 });
      if (t >= 11.5) {
        const fly = ease.inCubic(clamp((t - 11.8) / 0.22));
        set(sticker, { s: lerp(1.8, 1, sp) * (1 - 0.4 * fly), r: -8 + (1 - sp) * -26 + fly * 40, x: fly * 500, y: fly * -200, o: (1 - fly) * clamp((t - 11.5) / 0.05) });
      }
      drawIcon(box, t, 11.55, { dur: 0.4, stagger: 0.05, fillAt: 0.22 });
    },
  });
}
