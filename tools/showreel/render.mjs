// Renders the showreel stage frame by frame with headless Chromium.
//
//   node render.mjs --cues                      write out/cues.json (sound cues from the scenes)
//   node render.mjs --stills 2.1,8.2            PNG stills in out/stills/
//   node render.mjs --sheet 8:10:0.25           contact sheet of a time range in out/sheets/
//   node render.mjs --preview [--from 0 --to 30 --fps 30 --scale 0.5]
//                                               quick MP4 without motion blur (out/preview.mp4)
//   node render.mjs --final [--workers 2 --samples 4]
//                                               1920×1080, 60 fps, sub-frame motion blur, muxed
//                                               with out/soundtrack.wav → wavemark-intro.mp4 and poster.png
//
// Time is the only input: each frame calls window.__seek(t) and captures.
// Motion blur averages several captures inside a 180° shutter.
import { chromium } from 'playwright';
import http from 'node:http';
import fs from 'node:fs';
import path from 'node:path';
import zlib from 'node:zlib';
import { spawn } from 'node:child_process';
import { fileURLToPath } from 'node:url';

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, '../..');
const out = path.join(here, 'out');
fs.mkdirSync(out, { recursive: true });

const argv = process.argv.slice(2);
const flag = (k) => argv.includes(`--${k}`);
const opt = (k, d) => { const i = argv.indexOf(`--${k}`); return i >= 0 && argv[i + 1] && !argv[i + 1].startsWith('--') ? argv[i + 1] : d; };

// ---------------------------------------------------------------- server
const types = { '.html': 'text/html', '.mjs': 'text/javascript', '.js': 'text/javascript', '.css': 'text/css', '.json': 'application/json', '.woff2': 'font/woff2', '.png': 'image/png', '.svg': 'image/svg+xml' };
const allowed = [path.join(root, 'internal/theme/assets'), path.join(here, 'stage')];
function serve() {
  return new Promise((resolve) => {
    const srv = http.createServer((req, res) => {
      const p = path.normalize(path.join(root, decodeURIComponent(new URL(req.url, 'http://x').pathname)));
      if (!allowed.some((a) => p.startsWith(a + path.sep)) || !fs.existsSync(p) || fs.statSync(p).isDirectory()) { res.writeHead(404); res.end(); return; }
      res.writeHead(200, { 'content-type': types[path.extname(p)] || 'application/octet-stream', 'cache-control': 'no-store' });
      fs.createReadStream(p).pipe(res);
    });
    srv.listen(0, '127.0.0.1', () => resolve(srv));
  });
}

async function openStage(port, scale = 1) {
  const browser = await chromium.launch({ args: ['--font-render-hinting=none', '--force-color-profile=srgb', '--disable-lcd-text', '--hide-scrollbars'] });
  // deviceScaleFactor below 1 is not honoured by screenshots; scale the capture clip instead.
  const page = await browser.newPage({ viewport: { width: 1920, height: 1080 }, deviceScaleFactor: 1 });
  page.on('pageerror', (e) => console.error('page error:', e.message));
  page.on('console', (m) => { if (m.type() === 'error') console.error('console:', m.text()); });
  await page.goto(`http://127.0.0.1:${port}/tools/showreel/stage/index.html`);
  await page.waitForFunction(() => window.__ready || window.__error, null, { timeout: 120000 });
  const err = await page.evaluate(() => window.__error);
  if (err) throw new Error(err);
  const client = await page.context().newCDPSession(page);
  return {
    browser, page,
    async shot(t) {
      await page.evaluate((t) => window.__seek(t), t);
      const { data } = await client.send('Page.captureScreenshot', { format: 'png', optimizeForSpeed: true, ...(scale !== 1 ? { clip: { x: 0, y: 0, width: 1920, height: 1080, scale } } : {}) });
      return Buffer.from(data, 'base64');
    },
  };
}

// ---------------------------------------------------------------- PNG decode
function decodePNG(buf) {
  let pos = 8, width = 0, height = 0, ctype = 0;
  const idat = [];
  while (pos < buf.length) {
    const len = buf.readUInt32BE(pos);
    const type = buf.toString('ascii', pos + 4, pos + 8);
    const data = buf.subarray(pos + 8, pos + 8 + len);
    if (type === 'IHDR') { width = data.readUInt32BE(0); height = data.readUInt32BE(4); ctype = data[9]; if (data[8] !== 8 || data[12] !== 0) throw new Error('unsupported PNG'); }
    else if (type === 'IDAT') idat.push(data);
    else if (type === 'IEND') break;
    pos += 12 + len;
  }
  const bpp = ctype === 6 ? 4 : ctype === 2 ? 3 : 0;
  if (!bpp) throw new Error(`unsupported PNG color type ${ctype}`);
  const raw = zlib.inflateSync(Buffer.concat(idat));
  const stride = width * bpp;
  const rgb = Buffer.allocUnsafe(width * height * 3);
  let prev = Buffer.alloc(stride), cur = Buffer.alloc(stride);
  for (let y = 0; y < height; y++) {
    const f = raw[y * (stride + 1)];
    const line = raw.subarray(y * (stride + 1) + 1, (y + 1) * (stride + 1));
    for (let i = 0; i < stride; i++) {
      const a = i >= bpp ? cur[i - bpp] : 0, b = prev[i], c = i >= bpp ? prev[i - bpp] : 0;
      let v = line[i];
      if (f === 1) v += a;
      else if (f === 2) v += b;
      else if (f === 3) v += (a + b) >> 1;
      else if (f === 4) { const p = a + b - c, pa = Math.abs(p - a), pb = Math.abs(p - b), pc = Math.abs(p - c); v += pa <= pb && pa <= pc ? a : pb <= pc ? b : c; }
      cur[i] = v & 255;
    }
    for (let x = 0, o = y * width * 3; x < width; x++, o += 3) { rgb[o] = cur[x * bpp]; rgb[o + 1] = cur[x * bpp + 1]; rgb[o + 2] = cur[x * bpp + 2]; }
    [prev, cur] = [cur, prev];
  }
  return { width, height, rgb };
}

function ffmpeg(args) {
  const p = spawn('ffmpeg', ['-hide_banner', '-loglevel', 'error', '-y', ...args], { stdio: ['pipe', 'inherit', 'inherit'] });
  const done = new Promise((res, rej) => p.on('close', (c) => (c === 0 ? res() : rej(new Error(`ffmpeg exited ${c}`)))));
  return { stdin: p.stdin, done };
}
const write = (stream, buf) => new Promise((res) => (stream.write(buf) ? res() : stream.once('drain', res)));

// Renders frames [f0, f1) into a lossless segment, on an already open stage.
// Fast moves (whips, bursts, the pull-back) get more sub-frames so their blur
// is a smooth streak rather than a few ghost copies.
const FAST = [[3.98, 4.5], [7.45, 7.8], [7.98, 8.35], [9.68, 10.32], [11.74, 12.06], [13.18, 13.55], [13.78, 14.08], [15.74, 16.1], [17.72, 18.14], [19.0, 19.36], [19.7, 20.14], [21.0, 21.5], [21.55, 22.0], [23.76, 24.1], [26.95, 27.66], [27.98, 28.35]];
const samplesAt = (t, base) => (base > 1 && FAST.some(([a, b]) => t >= a && t <= b) ? Math.max(base, 12) : base);
async function renderSegment(st, { f0, f1, fps, samples, scale, file, label }) {
  const w = Math.round(1920 * scale), hgt = Math.round(1080 * scale);
  const enc = ffmpeg(['-f', 'rawvideo', '-pix_fmt', 'rgb24', '-video_size', `${w}x${hgt}`, '-framerate', String(fps), '-i', '-', '-c:v', 'ffv1', '-level', '3', '-pix_fmt', 'bgr0', file]);
  const acc = new Uint16Array(w * hgt * 3);
  const frameOut = Buffer.allocUnsafe(w * hgt * 3);
  const t0 = Date.now();
  for (let f = f0; f < f1; f++) {
    const t = f / fps;
    const n = samplesAt(t, samples);
    if (n <= 1) {
      const { rgb, width } = decodePNG(await st.shot(t));
      if (width !== w) throw new Error(`captured ${width}px wide, expected ${w}`);
      await write(enc.stdin, rgb);
    } else {
      acc.fill(0);
      // 180° shutter centred on the frame time.
      for (let s = 0; s < n; s++) {
        const ts = Math.max(0, t + ((s + 0.5) / n - 0.5) * (0.5 / fps));
        const { rgb, width } = decodePNG(await st.shot(ts));
        if (width !== w) throw new Error(`captured ${width}px wide, expected ${w}`);
        for (let i = 0; i < rgb.length; i++) acc[i] += rgb[i];
      }
      const half = n >> 1;
      for (let i = 0; i < acc.length; i++) frameOut[i] = (acc[i] + half) / n;
      await write(enc.stdin, frameOut);
    }
  }
  enc.stdin.end();
  await enc.done;
  console.log(`${label} frames ${f0}–${f1 - 1} in ${((Date.now() - t0) / 1000).toFixed(0)} s`);
}

const srv = await serve();
const port = srv.address().port;
try {
  if (flag('cues')) {
    const st = await openStage(port);
    const cues = await st.page.evaluate(() => window.__cues());
    fs.writeFileSync(path.join(out, 'cues.json'), JSON.stringify(cues, null, 1));
    console.log(`${cues.length} cues → out/cues.json`);
    await st.browser.close();
  } else if (flag('stills')) {
    const dir = path.join(out, 'stills');
    fs.mkdirSync(dir, { recursive: true });
    const st = await openStage(port, +opt('scale', 1));
    for (const t of opt('stills').split(',').map(Number)) {
      fs.writeFileSync(path.join(dir, `t${t.toFixed(3)}.png`), await st.shot(t));
      console.log(`out/stills/t${t.toFixed(3)}.png`);
    }
    await st.browser.close();
  } else if (flag('sheet')) {
    const [a, b, step] = opt('sheet').split(':').map(Number);
    const dir = path.join(out, 'sheets');
    fs.mkdirSync(dir, { recursive: true });
    const scale = +opt('scale', 0.25);
    const st = await openStage(port, scale);
    const times = [];
    for (let t = a; t <= b + 1e-6; t += step) times.push(+t.toFixed(4));
    const tmp = fs.mkdtempSync(path.join(out, 'sheet-'));
    for (const [i, t] of times.entries()) fs.writeFileSync(path.join(tmp, `${String(i).padStart(4, '0')}.png`), await st.shot(t));
    await st.browser.close();
    const cols = +opt('cols', 4);
    const rows = Math.ceil(times.length / cols);
    const name = path.join(dir, `sheet-${a}-${b}.png`);
    const drawtext = (n) => n; // timestamps are in the file names of the stills
    await ffmpeg(['-framerate', '1', '-i', path.join(tmp, '%04d.png'), '-vf', `tile=${cols}x${rows}:padding=6:margin=6:color=0x808080`, '-frames:v', '1', name]).done;
    fs.rmSync(tmp, { recursive: true });
    console.log(`${drawtext(name)}  (${times.map((t) => t.toFixed(2)).join(' ')})`);
  } else if (flag('preview') || flag('final')) {
    const final = flag('final');
    const fps = +opt('fps', final ? 60 : 30);
    const scale = +opt('scale', final ? 1 : 0.5);
    const samples = +opt('samples', final ? 4 : 1);
    const workers = +opt('workers', final ? 2 : 1);
    const from = +opt('from', 0), to = +opt('to', 30);
    const F0 = Math.round(from * fps), F1 = Math.round(to * fps);
    const segDir = path.join(out, final ? 'segments' : 'preview-seg');
    fs.rmSync(segDir, { recursive: true, force: true });
    fs.mkdirSync(segDir, { recursive: true });
    // Short segments in a shared queue keep both workers busy to the end.
    const SEG = Math.max(30, Math.round(fps * 2));
    const segs = [];
    for (let f0 = F0, k = 0; f0 < F1; f0 += SEG, k++) segs.push({ f0, f1: Math.min(F1, f0 + SEG), fps, samples, scale, file: path.join(segDir, `seg${String(k).padStart(3, '0')}.mkv`), label: `[seg ${k}]` });
    const t0 = Date.now();
    let next = 0, done = 0;
    await Promise.all(Array.from({ length: workers }, async (_, wk) => {
      const st = await openStage(port, scale);
      try {
        while (next < segs.length) {
          const s = segs[next++];
          await renderSegment(st, { ...s, label: `[w${wk}] ${s.label}` });
          done++;
          const el = (Date.now() - t0) / 60000;
          console.log(`  ${done}/${segs.length} segments, ${el.toFixed(1)} min elapsed, ~${((el / done) * (segs.length - done)).toFixed(1)} min left`);
        }
      } finally {
        await st.browser.close();
      }
    }));
    fs.writeFileSync(path.join(segDir, 'list.txt'), segs.map((s) => `file '${s.file}'`).join('\n'));
    const audio = path.join(out, 'soundtrack.wav');
    const hasAudio = fs.existsSync(audio);
    const target = final ? path.join(here, 'wavemark-intro.mp4') : path.join(out, 'preview.mp4');
    const args = ['-f', 'concat', '-safe', '0', '-i', path.join(segDir, 'list.txt')];
    if (hasAudio) args.push('-ss', String(from), '-t', String(to - from), '-i', audio);
    args.push('-map', '0:v');
    if (hasAudio) args.push('-map', '1:a', '-c:a', 'aac', '-b:a', '320k', '-ar', '48000');
    args.push('-c:v', 'libx264', '-preset', final ? 'slow' : 'veryfast', '-crf', final ? '14' : '20', '-pix_fmt', 'yuv420p', '-profile:v', 'high', '-tune', 'animation', '-colorspace', 'bt709', '-color_primaries', 'bt709', '-color_trc', 'bt709', '-movflags', '+faststart', '-r', String(fps), target);
    await ffmpeg(args).done;
    // The end card, held, doubles as the poster frame.
    if (final) await ffmpeg(['-ss', '29.5', '-i', target, '-frames:v', '1', path.join(here, 'poster.png')]).done;
    console.log(`${path.relative(here, target)}  ${(F1 - F0)} frames in ${((Date.now() - t0) / 60000).toFixed(1)} min${hasAudio ? ' (with soundtrack)' : ''}`);
  } else {
    console.log(fs.readFileSync(fileURLToPath(import.meta.url), 'utf8').split('\n').slice(0, 14).join('\n'));
  }
} finally {
  srv.close();
}
