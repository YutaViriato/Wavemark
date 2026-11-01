// Transition timings shared by the outgoing and incoming scenes.
import { lerp, ease, clamp } from './engine.mjs';

// Brand → chapter 1: a clay band with hand-drawn wavy edges sweeps left to
// right; the next scene appears behind its trailing edge.
export const WAVE = {
  t0: 9.7,
  lead: (t, ahead = 0) => lerp(-320, 2320, ease.inOutCubic(clamp((t - 9.7 + ahead) / 0.42))),
  trail: (t) => lerp(-320, 2320, ease.inOutCubic(clamp((t - 9.9) / 0.42))),
  phase: (t) => t * 5.5,
};
