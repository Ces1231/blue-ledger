/**
 * Blue Ledger — PWA icon generator
 * Generates icon-192.png and icon-512.png for the web app manifest.
 * Uses `jimp` (pure-JS, no native binaries required).
 *
 * Run: node scripts/generate-icons.mjs
 */

import { Jimp } from 'jimp'
import { mkdirSync } from 'fs'
import { fileURLToPath } from 'url'
import { dirname, join } from 'path'

const __dir = dirname(fileURLToPath(import.meta.url))
const OUT_DIR = join(__dir, '..', 'public', 'icons')

mkdirSync(OUT_DIR, { recursive: true })

// Colours as RGBA integers
const NAVY     = 0x001A4DFF
const GOLD     = 0xC9A84CFF
const GOLD_DIM = 0xC9A84C88

function setPixel(img, x, y, color, w, h) {
  if (x >= 0 && x < w && y >= 0 && y < h) img.setPixelColor(color, x, y)
}

function hLine(img, x0, x1, y, color, thick, w, h) {
  const half = Math.floor(thick / 2)
  for (let dy = -half; dy <= half; dy++)
    for (let x = x0; x <= x1; x++)
      setPixel(img, x, y + dy, color, w, h)
}

function vLine(img, x, y0, y1, color, thick, w, h) {
  const half = Math.floor(thick / 2)
  for (let dx = -half; dx <= half; dx++)
    for (let y = y0; y <= y1; y++)
      setPixel(img, x + dx, y, color, w, h)
}

function bline(img, x0, y0, x1, y1, color, thick, w, h) {
  const dx = Math.abs(x1 - x0), sx = x0 < x1 ? 1 : -1
  const dy = -Math.abs(y1 - y0), sy = y0 < y1 ? 1 : -1
  let err = dx + dy, cx = x0, cy = y0
  const half = Math.floor(thick / 2)
  while (true) {
    for (let bx = -half; bx <= half; bx++)
      for (let by = -half; by <= half; by++)
        setPixel(img, cx + bx, cy + by, color, w, h)
    if (cx === x1 && cy === y1) break
    const e2 = 2 * err
    if (e2 >= dy) { err += dy; cx += sx }
    if (e2 <= dx) { err += dx; cy += sy }
  }
}

function circle(img, cx, cy, r, color, w, h) {
  for (let y = cy - r; y <= cy + r; y++)
    for (let x = cx - r; x <= cx + r; x++)
      if ((x - cx) ** 2 + (y - cy) ** 2 <= r * r)
        setPixel(img, x, y, color, w, h)
}

function ring(img, cx, cy, r, thick, color, w, h) {
  const inner = (r - thick) ** 2
  const outer = (r + thick) ** 2
  for (let y = cy - r - thick; y <= cy + r + thick; y++)
    for (let x = cx - r - thick; x <= cx + r + thick; x++) {
      const d2 = (x - cx) ** 2 + (y - cy) ** 2
      if (d2 >= inner && d2 <= outer) setPixel(img, x, y, color, w, h)
    }
}

async function generateIcon(size) {
  const img = new Jimp({ width: size, height: size, color: NAVY })
  const w = size, h = size
  const cx = Math.floor(w / 2), cy = Math.floor(h / 2)
  const t = Math.max(2, Math.floor(size / 60)) // thickness unit

  // ── Phi (Φ) — left third ─────────────────────────────────────────────────
  const pCX = Math.floor(w * 0.22)
  const pR  = Math.floor(w * 0.09)
  vLine(img, pCX, Math.floor(cy - pR * 1.6), Math.floor(cy + pR * 1.6), GOLD, t * 2, w, h)
  ring(img, pCX, cy, pR, t, GOLD, w, h)

  // ── Beta (Β) — center ────────────────────────────────────────────────────
  const bL   = Math.floor(w * 0.41)
  const bTop = Math.floor(cy - pR * 1.6)
  const bBot = Math.floor(cy + pR * 1.6)
  const bMid = Math.floor((bTop + bBot) / 2)
  const bW   = Math.floor(w * 0.14)

  vLine(img, bL, bTop, bBot, GOLD, t * 2, w, h)
  hLine(img, bL, bL + Math.floor(bW * 0.75), bTop, GOLD, t * 2, w, h)
  hLine(img, bL, bL + Math.floor(bW * 0.85), bMid, GOLD, t * 2, w, h)
  hLine(img, bL, bL + Math.floor(bW * 0.75), bBot, GOLD, t * 2, w, h)
  vLine(img, bL + Math.floor(bW * 0.75), bTop, bMid, GOLD, t, w, h)
  vLine(img, bL + Math.floor(bW * 0.85), bMid, bBot, GOLD, t, w, h)

  // ── Sigma (Σ) — right third ──────────────────────────────────────────────
  const sL = Math.floor(w * 0.62)
  const sR = Math.floor(w * 0.82)

  hLine(img, sL, sR, bTop, GOLD, t * 2, w, h)
  hLine(img, sL, sR, bBot, GOLD, t * 2, w, h)
  bline(img, sR, bTop, sL + Math.floor((sR - sL) * 0.1), cy, GOLD, t * 2, w, h)
  bline(img, sL + Math.floor((sR - sL) * 0.1), cy, sR, bBot, GOLD, t * 2, w, h)

  // ── Decorative dot row at bottom ─────────────────────────────────────────
  const dotY = Math.floor(h * 0.86)
  const dotR = Math.max(2, t)
  for (let i = 0; i < 5; i++) {
    const dotX = Math.floor(w * 0.28 + i * w * 0.11)
    circle(img, dotX, dotY, dotR, GOLD_DIM, w, h)
  }

  return img
}

for (const size of [192, 512]) {
  process.stdout.write(`Generating ${size}x${size}...`)
  const img = await generateIcon(size)
  const outPath = join(OUT_DIR, `icon-${size}.png`)
  await img.write(outPath)
  console.log(` done -> ${outPath}`)
}

console.log('\nIcons written to web/public/icons/')
