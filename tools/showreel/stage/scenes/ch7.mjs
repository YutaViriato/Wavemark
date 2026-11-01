// 21.55–24.05 s. 07 For agents: the one sentence from the README, the agent's
// run (install, spec, check --json), and the site a blind agent actually
// built, assembled from slices of its real screenshot.
import { scene, h, set, css, clamp, lerp, ease, spring, cue, panX } from '../engine.mjs';
import { headline, wordsIn, label, typeIn, pop } from '../kit.mjs';
import { btnCenter } from './ch6.mjs';

const CARD = { x: 1200, y: 330, w: 600, h: 375 };
const PHRASES = ['Use Wavemark from ', 'wavemark.viriato.site/install-wavemark ', 'and create a docs site ', 'for Tidepool.'];
const TRACE = [
  '$ curl -fsSL https://wavemark.viriato.site/install.sh | sh',
  '$ wavemark spec',
  '$ wavemark check --json',
];

export async function build({ cam }) {
  const root = h('div', { class: 'sc', style: { background: '#141413', zIndex: 2 } });
  cam.append(root);
  const zoomer = h('div', { class: 'fill', style: { transformOrigin: '960px 540px' } });
  root.append(zoomer);

  const lab = label('07 — For agents');
  const hl = headline(['Written by people', 'and by agents.'], { color: '#faf9f5', size: 108 });
  zoomer.append(h('div', { class: 'abs', style: { left: '120px', top: '176px' } }, lab, h('div', { style: { height: '30px' } }), hl.el));

  const ask = h('div', { class: 'abs', text: 'Give your agent one sentence:', style: { left: '124px', top: '500px', font: '400 25px/1 "Hanken Grotesk", sans-serif', color: '#9c9a92' } });
  const bubble = h('div', { class: 'abs', style: { left: '120px', top: '542px', width: '1040px', minHeight: '170px', padding: '30px 110px 30px 36px', borderRadius: '28px', background: '#1f1e1d', border: '1px solid #3d3d3a', font: '400 34px/1.42 "Hanken Grotesk", sans-serif', color: '#faf9f5', boxShadow: '0 30px 70px -30px rgb(0 0 0 / .9)' } });
  const spans = PHRASES.map((p, i) => h('span', { text: p, style: { color: i === 1 ? '#e6a68d' : '#faf9f5', whiteSpace: i === 1 ? 'nowrap' : 'normal' } }));
  bubble.append(...spans);
  const send = h('div', { class: 'abs', style: { left: 'auto', top: 'auto', right: '24px', bottom: '24px', width: '60px', height: '60px', borderRadius: '50%', background: '#d97757', display: 'grid', placeItems: 'center' }, html: '<svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="#141413" stroke-width="2.6" stroke-linecap="round" stroke-linejoin="round"><path d="M12 19V5"/><path d="m5 12 7-7 7 7"/></svg>' });
  bubble.append(send);
  zoomer.append(ask, bubble);

  const trace = h('div', { class: 'abs sr-mono', style: { left: '124px', top: '800px', font: '400 23px/40px "JetBrains Mono", monospace', color: '#c2c0b6', whiteSpace: 'pre' } });
  const tLines = TRACE.map((l) => h('div', { html: `<span class="tk-prompt">$</span> ${l.slice(2)}  <span class="tk-ok">✓</span>` }));
  const json = h('div', { html: '<span class="tk-pun">{</span> <span class="tk-attr">"ok"</span>: <span class="tk-ok">true</span>, <span class="tk-attr">"pages"</span>: <span class="tk-code">12</span>, <span class="tk-attr">"errors"</span>: <span class="tk-ok">0</span>, <span class="tk-attr">"warnings"</span>: <span class="tk-code">0</span> <span class="tk-pun">}</span>' });
  trace.append(...tLines, json);
  zoomer.append(trace);

  // The Tidepool site from .e2e-agents, cut from its real screenshot into
  // header, hero and feature bands that assemble separately.
  const shot = './gen/shots/tidepool-light.png';
  const img = new Image();
  img.src = shot;
  await img.decode();
  const k = CARD.w / 1440;
  const cuts = [[0, 84], [84, 640], [640, 900]];
  const frame = h('div', { class: 'abs', style: { left: `${CARD.x}px`, top: `${CARD.y}px`, width: `${CARD.w}px`, height: `${CARD.h}px`, borderRadius: '16px', overflow: 'hidden', background: '#faf9f5', boxShadow: '0 40px 90px -30px rgb(0 0 0 / .9), 0 0 0 1px rgb(255 255 255 / .08)', transformOrigin: '50% 50%' } });
  const bands = cuts.map(([a, b]) => {
    const el = h('div', { class: 'abs', style: { left: '0', top: `${a * k}px`, width: `${CARD.w}px`, height: `${(b - a) * k}px`, backgroundImage: `url(${shot})`, backgroundSize: `${CARD.w}px ${CARD.h}px`, backgroundPosition: `0 ${-a * k}px`, transformOrigin: '50% 0' } });
    frame.append(el);
    return el;
  });
  const caption = h('div', { class: 'abs sr-mono', html: '<span style="color:#faf9f5">tidepool/</span>  12 pages · written by an agent', style: { left: `${CARD.x}px`, top: `${CARD.y + CARD.h + 26}px`, font: '400 21px/1 "JetBrains Mono", monospace', color: '#87867f', whiteSpace: 'pre' } });
  zoomer.append(frame, caption);

  PHRASES.forEach((_, i) => cue(22.125 + i * 0.125, 'type', { n: 5, pan: -0.3, vel: 0.35 }));
  cue(22.625, 'click', { pan: 0.1, vel: 0.7 });
  cue(22.625, 'zip', { pan: 0.1, vel: 0.4 });
  [22.75, 22.875, 23.0].forEach((t, i) => cue(t, 'tick', { pan: -0.4, vel: 0.5, note: 86 + i * 2 }));
  cue(23.125, 'success', { pan: -0.3, vel: 0.6, soft: true });
  [23.25, 23.32, 23.39].forEach((t, i) => cue(t, 'pop', { note: [81, 85, 88][i], pan: 0.6, vel: 0.5 }));
  cue(23.78, 'whoosh', { dur: 0.4, vel: 0.7, tone: 1.2, rev: true });

  scene({
    t0: 21.55, t1: 24.06, root,
    update(t) {
      const R = lerp(0, 2300, ease.inCubic(clamp((t - 21.55) / 0.42)));
      css(root, 'clip-path', t < 21.98 ? `circle(${R.toFixed(1)}px at ${btnCenter.x}px ${btnCenter.y}px)` : 'none');

      typeIn(lab, t, 21.96, 0.02);
      wordsIn(hl.words, t, 21.98, { stagger: 0.06, em: 108 });
      set(ask, { o: clamp((t - 22.05) / 0.1) });
      const bi = spring(t - 22.02, 2.4, 0.72);
      set(bubble, { y: (1 - bi) * 60, o: clamp((t - 22.02) / 0.1) });
      spans.forEach((s, i) => { const lt = t - 22.125 - i * 0.125; set(s, { o: clamp(lt / 0.05) }); });
      const sp = t - 22.625;
      pop(send, t, 22.5, { f: 4, z: 0.45, from: 0 });
      if (sp > 0) set(send, { s: 1 - 0.2 * Math.exp(-sp * 16) });
      tLines.forEach((l, i) => { const lt = t - (22.75 + i * 0.125); set(l, { x: (1 - ease.outExpo(clamp(lt / 0.3))) * -26, o: clamp(lt / 0.04) }); });
      { const lt = t - 23.125; set(json, { x: (1 - ease.outExpo(clamp(lt / 0.3))) * -26, o: clamp(lt / 0.04) }); }

      // the agent's site assembles from its slices
      const fi = spring(t - 23.2, 2.4, 0.68);
      set(frame, { s: 0.85 + 0.15 * fi, r: (1 - fi) * 6, o: clamp((t - 23.2) / 0.06) });
      bands.forEach((b, i) => {
        const lt = t - 23.25 - i * 0.07;
        const p = spring(lt, 3, 0.6);
        set(b, { y: (1 - p) * (i === 0 ? -40 : 60), rx: (1 - p) * -50, persp: 900, o: clamp(lt / 0.06) });
      });
      set(caption, { o: clamp((t - 23.45) / 0.1) });

      // exit: rush toward the camera, the wall is behind
      const ex = ease.inCubic(clamp((t - 23.78) / 0.27));
      set(zoomer, { s: 1 + 0.5 * ex, o: 1 - ex, blur: ex * 10 });
      set(root, { bg: `rgb(20 20 19 / ${1 - ex})` });
    },
  });
}
