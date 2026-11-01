// 0–7.75 s. A caret types "# Markdown"; the mono letters morph into the serif
// hero of the real landing page; the camera pulls back to reveal that the
// headline lives inside a browser window, next to its Markdown source, and the
// rest of the page assembles line by line on the beat. Then it all collapses.
import { scene, h, set, css, draw, clamp, lerp, ease, tw, spring, springVel, splitGlyphs, splitWords, cue, panX, rng, TAU } from '../engine.mjs';
import { FRAG, codeWindow, typeRow, pop, burst, wordsIn } from '../kit.mjs';

const WIN = { x: 860, y: 150, w: 960, h: 800 };
const S1 = WIN.w / 1440; // page scale inside the window
const EDITOR = { x: 100, y: 196, w: 700, h: 640 };
const SOURCE = [
  '---',
  'title: Wavemark',
  'layout: landing',
  '---',
  '',
  '<Hero eyebrow="Wavemark"',
  '      title="Markdown that becomes a site while you write"',
  '      highlight="while you write">',
  'One binary, a folder of `.md` files and no build step.',
  '',
  '<Button href="./install.md" label="Install" />',
  '</Hero>',
  '',
  '## Write Markdown. The rest is runtime.',
  '',
  '<Columns cols={3}>',
  '<Card title="No build step" icon="bolt" />',
  '<Card title="Strict on purpose" icon="shield" />',
  '<Card title="Declarative components" icon="layers" />',
  '</Columns>',
];

export async function build({ cam, stage }) {
  const root = h('div', { class: 'sc', style: { background: '#faf9f5' } });
  cam.append(root);
  const world = h('div', { class: 'abs', style: { width: '1920px', height: '1080px', transformOrigin: '0 0' } });
  root.append(world);

  // ---------------------------------------------------------- the real page
  const win = h('div', { class: 'sr-window', style: { left: `${WIN.x}px`, top: `${WIN.y}px`, width: `${WIN.w}px`, height: `${WIN.h}px` } });
  const dots = [h('span', { class: 'sr-dot' }), h('span', { class: 'sr-dot' }), h('span', { class: 'sr-dot' })];
  const url = h('div', { class: 'sr-url', html: '<svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round"><rect x="4" y="10" width="16" height="11" rx="2"/><path d="M8 10V7a4 4 0 0 1 8 0v3"/></svg><span>localhost:4321</span>' });
  const bar = h('div', { class: 'sr-bar' }, ...dots, url);
  const vp = h('div', { class: 'sr-viewport' });
  const page = h('div', { class: 'sr-page', 'data-layout': 'landing', style: { font: '400 14px/1.5 var(--wm-font-sans)', color: 'var(--wm-fg)' } });
  const src = h('div', { html: FRAG.landing });
  const header = src.querySelector('.wm-header');
  const main = src.querySelector('.wm-landing');
  header.querySelector('.wm-theme-btn')?.removeAttribute('hidden');
  const prose = main.querySelector('.wm-prose--landing');
  const stats = prose.querySelector('.wm-columns--4');
  while (stats.nextElementSibling) stats.nextElementSibling.remove();
  page.append(header, main);
  vp.append(page);
  win.append(bar, vp);
  world.append(win);

  // Parts of the page that animate on their own.
  const q = (s) => page.querySelector(s);
  const qa = (s) => [...page.querySelectorAll(s)];
  const logo = q('.wm-logo');
  const navs = qa('.wm-topnav__link');
  const themeBtn = q('.wm-theme-btn');
  const eyebrow = q('.wm-hero__eyebrow');
  const heroWords = qa('.wm-hero__title .wm-word');
  const heroBody = q('.wm-hero__body > p');
  const buttons = qa('.wm-hero .wm-button');
  const markPath = q('.wm-mark__line path');
  const h2 = q('h2.wm-heading');
  h2.querySelector('.wm-anchor')?.remove();
  const h2Words = splitWords(h2);
  const cards = qa('.wm-columns--3 > .wm-card').map((el) => ({
    el,
    ink: [...el.querySelectorAll('.wm-sketch__ink path')],
    fill: el.querySelector('.wm-sketch__fill'),
    parts: [el.querySelector('.wm-card__title'), el.querySelector('.wm-card__body'), el.querySelector('.wm-card__cta')],
  }));
  const statEls = qa('.wm-columns--4 > .wm-stat').map((el) => ({ el, value: el.querySelector('.wm-stat__value'), label: el.querySelector('.wm-stat__label') }));
  for (const s of statEls) { s.el.style.borderTopColor = 'transparent'; s.rule = h('div', { style: { position: 'absolute', left: 0, right: 0, top: '-1px', height: '1px', background: 'var(--wm-border-strong)', transformOrigin: '0 50%' } }); s.el.style.position = 'relative'; s.el.prepend(s.rule); }
  markPath.style.strokeDasharray = '1 2';
  for (const c of cards) for (const p of c.ink) p.style.strokeDasharray = '1 2';

  // Kerning-true glyphs of "Markdown", for the morph from the mono line.
  const mdWord = heroWords[0];
  const glyphs = splitGlyphs(mdWord);

  // Local page coordinates (page at scale 1) → B-space (final layout).
  const pr = page.getBoundingClientRect();
  const local = (el) => { const r = el.getBoundingClientRect(); return { x: r.left - pr.left, y: r.top - pr.top, w: r.width, h: r.height }; };
  const toB = (lx, ly, scroll = 0) => ({ x: WIN.x + lx * S1, y: WIN.y + 44 + (ly - scroll) * S1 });
  const titleR = local(q('.wm-hero__title'));
  const mdR = local(mdWord);
  const glyphLocal = glyphs.map((g) => ({ x: mdR.x + g._box.x + g._box.w / 2, y: mdR.y + g._box.y + g._box.h / 2 }));
  const targetOf = (el, scroll) => { const r = local(el); return toB(r.x + r.w / 2, r.y + r.h / 2, scroll); };
  const P = toB(titleR.x + titleR.w / 2, titleR.y + titleR.h / 2);
  css(page, 'transform', `scale(${S1})`);

  // ---------------------------------------------------------- the source
  const ed = codeWindow({ ...EDITOR, tab: 'index.md', lines: SOURCE });
  world.append(ed.el);
  const tabCheck = h('span', { html: '<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="#a3b88a" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6 9 17l-5-5"/></svg>', style: { display: 'inline-grid', marginLeft: '-2px' } });
  ed.tab.append(tabCheck);

  // Packets: a clay dot with a short comet tail that carries a line of source
  // to the part of the page it becomes.
  const packets = h('div', { class: 'abs' });
  world.append(packets);
  const lineEnd = (row) => ({ x: EDITOR.x + 52 + SOURCE[row].length * ed.cw + 14, y: EDITOR.y + 46 + 18 + row * ed.lh + ed.lh / 2 });
  const mkPacket = (row, target, t0, dur = 0.26) => {
    const dots = [0, 1, 2, 3, 4].map((i) => {
      const d = h('div', { class: 'abs', style: { width: `${14 - i * 2}px`, height: `${14 - i * 2}px`, marginLeft: `${-(7 - i)}px`, marginTop: `${-(7 - i)}px`, borderRadius: '50%', background: '#d97757' } });
      packets.append(d);
      return d;
    });
    return { row, target, t0, dur, dots, from: lineEnd(row) };
  };

  // Mono line (screen space) and caret.
  const overlay = h('div', { class: 'abs', style: { width: '1920px', height: '1080px' } });
  root.append(overlay);
  const MONO = { size: 150, x0: 510, top: 465 };
  const monoText = '# Markdown';
  const mono = [...monoText].map((ch, i) => {
    const el = h('span', { class: 'abs sr-mono', text: ch, style: { font: `400 ${MONO.size}px/1 "JetBrains Mono", monospace`, left: `${MONO.x0 + i * MONO.size * 0.6}px`, top: `${MONO.top}px`, width: `${MONO.size * 0.6}px`, textAlign: 'center', color: i === 0 ? '#d97757' : '#141413', transformOrigin: '50% 60%' } });
    overlay.append(el);
    return el;
  });
  const caret = h('div', { class: 'abs', style: { width: '9px', height: '150px', borderRadius: '4px', background: '#d97757', top: `${MONO.top + 6}px`, left: '0', transformOrigin: '50% 50%' } });
  overlay.append(caret);
  const hashBurst = burst(overlay, { x: MONO.x0 + 45, y: MONO.top + 75, t0: 2.0, n: 10, seed: 3, chars: '#', colors: ['#d97757', '#e3dacc'], speed: [300, 700], size: [20, 40], life: [0.4, 0.8], gravity: 600 });

  // ---------------------------------------------------------- choreography
  const TYPE0 = 0.5, STEP = 0.125;
  mono.forEach((el, i) => { if (monoText[i] !== ' ') cue(TYPE0 + i * STEP, 'key', { pan: panX(MONO.x0 + i * 90), vel: i === 0 ? 0.9 : 0.7 + (i % 3) * 0.08 }); });
  cue(2.0, 'whoosh', { dur: 0.5, pan: -0.2, vel: 0.35, tone: 1.4 });
  cue(3.5, 'marker', { dur: 0.45, pan: 0.3, vel: 0.6 });
  cue(3.86, 'whoosh', { dur: 0.95, pan: 0.15, vel: 0.85, tone: 0.6, rev: true });
  cue(4.0, 'whoosh', { dur: 1.0, pan: 0.2, vel: 0.75, tone: 0.55 });
  cue(4.06, 'whoosh', { dur: 0.6, pan: -0.6, vel: 0.4, tone: 1.1 });

  const rowsAt = [];
  const rowTimes = (rows, t0, gap = 0.035) => rows.forEach((r, i) => (rowsAt[r] = t0 + i * gap));
  rowTimes([0, 1, 2, 3, 4], 4.22, 0.04);
  rowTimes([5, 6, 7, 8, 9], 4.48, 0.035);
  rowTimes([10, 11, 12], 4.98, 0.04);
  rowTimes([13, 14], 5.48, 0.04);
  rowTimes([15], 5.74);
  rowTimes([16], 5.86);
  rowTimes([17], 6.11);
  rowTimes([18, 19], 6.36, 0.06);
  for (const t of [4.22, 4.48, 4.98, 5.48, 5.74, 5.86, 6.11, 6.36]) cue(t, 'type', { n: 6, pan: -0.5, vel: 0.35 });
  const hlAt = { 5: 4.5, 6: 4.5, 7: 4.5, 8: 4.5, 10: 5.0, 13: 5.5, 15: 5.75, 16: 5.875, 17: 6.125, 18: 6.375 };

  const packetList = [
    mkPacket(8, targetOf(heroBody), 4.58, 0.28),
    mkPacket(10, targetOf(buttons[0]), 5.04, 0.24),
    mkPacket(13, targetOf(h2), 5.53, 0.22),
    mkPacket(16, targetOf(cards[0].el), 5.9, 0.2),
    mkPacket(17, targetOf(cards[1].el), 6.15, 0.2),
    mkPacket(18, targetOf(cards[2].el), 6.4, 0.2),
  ];
  for (const p of packetList) cue(p.t0, 'zip', { pan: -0.3, vel: 0.35 });
  const ARRIVE = { body: 4.86, buttons: [5.25, 5.36], h2: 5.75, cards: [6.0, 6.25, 6.5], stats: [6.75, 6.875, 7.0, 7.125] };
  cue(ARRIVE.body, 'pop', { note: 74, pan: 0.4, vel: 0.35 });
  ARRIVE.buttons.forEach((t, i) => cue(t, 'pop', { note: [78, 81][i], pan: 0.35, vel: 0.55 }));
  ARRIVE.cards.forEach((t, i) => cue(t, 'pop', { note: [74, 78, 81][i], pan: 0.2 + i * 0.25, vel: 0.65 }));
  ARRIVE.stats.forEach((t, i) => cue(t, 'tick', { pan: 0.1 + i * 0.2, vel: 0.5 }));
  cue(6.5, 'tick', { pan: -0.5, vel: 0.4 });
  cue(6.6, 'swipe', { dur: 0.5, pan: 0.4, vel: 0.3 });
  cue(7.5, 'whoosh', { dur: 0.26, pan: 0, vel: 0.9, tone: 0.8, rev: true });

  // Camera: screen = (p - P) * k + Q, applied to the whole B-space.
  const k0 = 150 / (64 * S1);
  const camAt = (t) => {
    let k, Q, rot = 0;
    if (t < 4) {
      const push = 1 + 0.035 * ease.inOutSine(clamp(t / 3.85)) + 0.025 * ease.inQuad(clamp((t - 3.82) / 0.18));
      k = k0 * push; Q = { x: 960, y: 540 };
    } else if (t < 7.5) {
      const u = ease.bezier(0.12, 0.8, 0.12, 1)(clamp((t - 4) / 1.05));
      const kStart = k0 * 1.06;
      k = Math.pow(kStart, 1 - u);
      Q = { x: lerp(960, P.x, u), y: lerp(540, P.y, u) };
      const push = 1 + 0.05 * ease.inCubic(clamp((t - 7) / 0.5));
      k *= push;
      Q = { x: lerp(Q.x, 960, ease.inCubic(clamp((t - 7) / 0.5)) * 0.3), y: Q.y };
      rot = -1.2 * ease.inCubic(clamp((t - 7) / 0.5));
    } else {
      const u = ease.inExpo(clamp((t - 7.5) / 0.25));
      k = 1.05 * lerp(1, 0.015, u);
      Q = { x: lerp(P.x * 0.7 + 960 * 0.3, 960, u), y: lerp(P.y, 540, u) };
      rot = -1.2 + 30 * u;
    }
    return { k, Q, rot };
  };
  const applyCam = (c) => {
    const tx = c.Q.x - P.x * c.k, ty = c.Q.y - P.y * c.k;
    css(world, 'transform', `translate(${tx.toFixed(2)}px,${ty.toFixed(2)}px) rotate(${c.rot.toFixed(3)}deg) scale(${c.k.toFixed(5)})`);
    css(world, 'transform-origin', `${P.x}px ${P.y}px`);
  };
  // rotate() above pivots around P because transform-origin is P; the
  // translate/scale math assumes origin 0 0, so convert explicitly instead.
  const applyCam2 = (c) => {
    // screen = R·k·(p - P) + Q
    css(world, 'transform-origin', '0 0');
    const a = (c.rot * Math.PI) / 180, cs = Math.cos(a) * c.k, sn = Math.sin(a) * c.k;
    const tx = c.Q.x - (cs * P.x - sn * P.y), ty = c.Q.y - (sn * P.x + cs * P.y);
    css(world, 'transform', `matrix(${cs.toFixed(6)},${sn.toFixed(6)},${(-sn).toFixed(6)},${cs.toFixed(6)},${tx.toFixed(2)},${ty.toFixed(2)})`);
  };
  const toScreen = (c, p) => {
    const a = (c.rot * Math.PI) / 180, cs = Math.cos(a) * c.k, sn = Math.sin(a) * c.k;
    const dx = p.x - P.x, dy = p.y - P.y;
    return { x: c.Q.x + cs * dx - sn * dy, y: c.Q.y + sn * dx + cs * dy };
  };
  void applyCam;

  scene({
    t0: 0, t1: 7.76, root,
    update(t) {
      const c = camAt(t);
      applyCam2(c);
      const scroll = 230 * ease.inOutCubic(clamp((t - 6.6) / 0.55));
      css(page, 'transform', `translateY(${(-scroll * S1).toFixed(2)}px) scale(${S1})`);
      set(root, { o: 1 - clamp((t - 7.7) / 0.05) });

      // ---- caret and typing
      const typed = t < TYPE0 ? 0 : Math.min(monoText.length, Math.floor((t - TYPE0) / STEP) + 1);
      const caretX = MONO.x0 + Math.min(typed, monoText.length) * 90 - 4;
      const prevX = MONO.x0 + Math.max(0, typed - 1) * 90 - 4;
      const since = t - (TYPE0 + Math.max(0, typed - 1) * STEP);
      const cx = typed === 0 ? MONO.x0 - 4 : lerp(prevX, caretX, spring(since, 9, 0.7));
      const blinkOn = t < TYPE0 || t < 1.75 || (t % 0.5) < 0.25;
      const caretGrow = ease.outExpo(clamp(t / 0.35));
      const caretGone = ease.inQuad(clamp((t - 1.95) / 0.12));
      set(caret, { x: cx, sy: caretGrow * (1 - caretGone), o: (blinkOn ? 1 : 0.12) * (t < 2.1 ? 1 : 0) });
      mono.forEach((el, i) => {
        const ti = TYPE0 + i * STEP;
        if (t < ti) { set(el, { o: 0 }); return; }
        const lt = t - ti;
        if (i === 0) {
          // the "#": types in, then flies off when the line becomes a heading
          const p = pop(el, t, ti, { f: 4, z: 0.45, from: 0.3, dy: 30 });
          void p;
          const g = clamp((t - 2) / 0.45);
          if (g > 0) set(el, { x: -260 * ease.outQuart(g), y: -230 * ease.outQuart(g) + 400 * g * g, r: -140 * ease.outQuart(g), s: 1 - 0.6 * g, o: 1 - ease.inQuad(g) });
          return;
        }
        if (i === 1) { set(el, { o: 0 }); return; }
        // letters of "Markdown"
        const j = i - 2;
        const tj = 2.0 + j * 0.024;
        if (t < tj) { pop(el, t, ti, { f: 4.2, z: 0.45, from: 0.25, dy: 34 }); return; }
        const u = spring(t - tj, 1.9, 0.78);
        const g = glyphLocal[j];
        const tgt = toScreen(c, toB(g.x, g.y, scroll));
        const src = { x: MONO.x0 + i * 90 + 45, y: MONO.top + 75 };
        set(el, { x: (tgt.x - src.x) * u, y: (tgt.y - src.y) * u, s: 1 - 0.12 * clamp((t - tj) / 0.2), o: 1 - clamp((t - tj - 0.03) / 0.14), blur: 7 * clamp((t - tj) / 0.18) });
        void lt;
      });
      hashBurst.update(t);

      // ---- serif glyphs of "Markdown" follow the mono letters home
      const kk = c.k * S1;
      glyphs.forEach((gl, j) => {
        const tj = 2.0 + j * 0.024;
        if (t < tj) { set(gl, { o: 0 }); return; }
        const u = spring(t - tj, 1.9, 0.78);
        const g = glyphLocal[j];
        const tgt = toScreen(c, toB(g.x, g.y, scroll));
        const src = { x: MONO.x0 + (j + 2) * 90 + 45, y: MONO.top + 75 };
        // offset in local page units (rotation is negligible before 7 s)
        const dx = ((src.x - tgt.x) * (1 - u)) / kk, dy = ((src.y - tgt.y) * (1 - u)) / kk;
        set(gl, { x: dx, y: dy, s: 0.92 + 0.08 * u, o: clamp((t - tj - 0.02) / 0.16), blur: (1 - clamp((t - tj) / 0.3)) * 9 / kk });
      });

      // ---- the rest of the tagline, one word per pluck
      const wordT = [null, 2.25, 2.5, 3.0, 3.06, 3.25, 3.375, 3.5];
      heroWords.forEach((w, i) => {
        if (i === 0) return;
        const lt = t - wordT[i];
        const p = spring(lt, 2.3, 0.7);
        set(w, { y: (1 - p) * 30, o: clamp(lt / 0.18), blur: (1 - ease.outCubic(clamp(lt / 0.32))) * 5 });
      });
      draw(markPath, ease.brand(clamp((t - 3.5) / 0.48)));

      // ---- header, hero body and buttons
      set(logo, { y: (1 - spring(t - 4.2, 2.6, 0.7)) * -24, o: clamp((t - 4.2) / 0.15) });
      navs.forEach((n, i) => set(n, { y: (1 - spring(t - 4.26 - i * 0.035, 2.8, 0.66)) * -20, o: clamp((t - 4.26 - i * 0.035) / 0.15) }));
      pop(themeBtn, t, 4.5, { f: 3, z: 0.5 });
      set(eyebrow, { y: (1 - spring(t - ARRIVE.body + 0.04, 2.4, 0.72)) * 16, o: clamp((t - ARRIVE.body + 0.04) / 0.18) });
      set(heroBody, { y: (1 - spring(t - ARRIVE.body, 2.4, 0.72)) * 22, o: clamp((t - ARRIVE.body) / 0.2), blur: (1 - clamp((t - ARRIVE.body) / 0.3)) * 4 });
      buttons.forEach((b, i) => pop(b, t, ARRIVE.buttons[i], { f: 3.6, z: 0.42, from: 0.2, dy: 16 }));

      // ---- H2, cards and stats
      wordsIn(h2Words, t, ARRIVE.h2, { stagger: 0.04, em: 44, blur: 5, dy: 0.5 });
      cards.forEach((cd, i) => {
        const t0 = ARRIVE.cards[i];
        const p = pop(cd.el, t, t0, { f: 2.8, z: 0.55, from: 0.6, dy: 40, squash: 0.08 });
        void p;
        cd.ink.forEach((path, k) => draw(path, ease.brand(clamp((t - t0 - 0.08 - k * 0.07) / 0.45))));
        if (cd.fill) set(cd.fill, { o: clamp((t - t0 - 0.35) / 0.12) });
        cd.parts.forEach((part, k) => part && set(part, { y: (1 - spring(t - t0 - 0.06 - k * 0.04, 2.6, 0.7)) * 14, o: clamp((t - t0 - 0.06 - k * 0.04) / 0.15) }));
      });
      statEls.forEach((s, i) => {
        const t0 = ARRIVE.stats[i];
        set(s.rule, { sx: ease.brandExpo(clamp((t - t0) / 0.5)), o: t >= t0 ? 1 : 0 });
        set(s.value, { y: (1 - spring(t - t0, 2.8, 0.62)) * 36, o: clamp((t - t0) / 0.1), blur: (1 - clamp((t - t0) / 0.2)) * 6 });
        set(s.label, { y: (1 - spring(t - t0 - 0.05, 2.6, 0.7)) * 12, o: clamp((t - t0 - 0.05) / 0.15) });
      });

      // ---- editor
      const edIn = spring(t - 4.05, 1.7, 0.78);
      set(ed.el, { x: (1 - edIn) * -460, y: (1 - edIn) * 60, r: (1 - edIn) * -5, o: t < 4.05 ? 0 : 1 });
      ed.rows.forEach((r, i) => {
        const t0 = rowsAt[i] ?? 99;
        typeRow(r, t, t0, 0.006, ed.cw);
        const hl = hlAt[i];
        if (hl !== undefined) {
          const a = clamp((t - hl) / 0.06) * (1 - clamp((t - hl - 0.45) / 0.3));
          set(r.hl, { o: a, sx: ease.outExpo(clamp((t - hl) / 0.2)) });
        }
      });
      css(ed.tabDot, 'display', t < 6.5 ? 'block' : 'none');
      pop(tabCheck, t, 6.5, { f: 4, z: 0.4 });
      dots.forEach((d, i) => pop(d, t, 4.35 + i * 0.05, { f: 4, z: 0.5 }));
      set(url, { o: clamp((t - 4.45) / 0.2), y: (1 - spring(t - 4.45, 3, 0.7)) * 8 });

      // ---- packets
      for (const pk of packetList) {
        pk.dots.forEach((d, i) => {
          const lt = t - pk.t0 - i * 0.018;
          if (lt < 0 || lt > pk.dur + 0.05) { set(d, { o: 0 }); return; }
          const u = ease.inOutCubic(clamp(lt / pk.dur));
          const a = pk.from, b = pk.target;
          const mx = (a.x + b.x) / 2, my = Math.min(a.y, b.y) - 120;
          const x = (1 - u) * (1 - u) * a.x + 2 * (1 - u) * u * mx + u * u * b.x;
          const y = (1 - u) * (1 - u) * a.y + 2 * (1 - u) * u * my + u * u * b.y;
          const end = clamp((lt - pk.dur) / 0.05);
          set(d, { x, y, o: (1 - i * 0.17) * (1 - end), s: 1 + end * 1.5 });
        });
      }
    },
  });
}
