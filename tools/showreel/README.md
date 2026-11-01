# Showreel

A 30-second video that introduces Wavemark: [wavemark-intro.mp4](wavemark-intro.mp4) (1920×1080, 60 fps, stereo AAC). [poster.png](poster.png) is its last frame.

Everything on screen comes from the product. The pages are Wavemark's own HTML and CSS, the icons and strokes come from `internal/sketch`, the terminal text is real `wavemark` output, and the screenshots show the docs site and the sites blind agents wrote. The soundtrack is synthesized from a score at 120 BPM that also drives the animation, so cuts, pops and keystrokes land on the beat.

## Layout

| Path | Contents |
|---|---|
| `stage/` | The composition: an HTML page whose state is a pure function of time |
| `stage/score.mjs` | Tempo, chords, melody and drums, shared by the picture and the sound |
| `stage/scenes/` | One module per scene |
| `stage/gen/` | Product assets captured by `prepare.mjs` |
| `audio/synth.mjs` | Synthesizer, sound design and mastering, written to `out/soundtrack.wav` |
| `render.mjs` | Frame-by-frame renderer with sub-frame motion blur |

## Rebuild

Needs Node, Go, ffmpeg and Playwright's Chromium.

```bash
cd tools/showreel
npm install && npx playwright install chromium
node render.mjs --cues     # sound cues from the scenes → out/cues.json
node audio/synth.mjs       # → out/soundtrack.wav
node render.mjs --final    # → wavemark-intro.mp4 and poster.png (about 35 min on 2 cores)
```

`node render.mjs --sheet 8:10:0.25` and `node render.mjs --stills 8.2` write review images. `node audio/analyze.mjs` prints the level of each stem, the spectral balance and the kick timing. `node render.mjs` with no arguments lists every option.

`node prepare.mjs` captures the assets again. It needs the sites in `.e2e-agents/`, which are not in git, so on a fresh clone keep the committed `stage/gen/`.
