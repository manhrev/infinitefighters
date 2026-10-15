import Phaser from 'phaser'
import manifest from '../../assets/sprites/manifest.json'

// Every runtime asset is listed in the manifest written by scripts/build_sprites.py.
const urls = import.meta.glob<string>('../../assets/{sprites,tiles,backgrounds,ui}/**/*.png', {
  eager: true, query: '?url', import: 'default',
})

// Texture and animation key: path without "assets/" and ".png", e.g. "sprites/player/teal/run".
export const assets = manifest.entries.map(e => ({
  ...e,
  key: e.path.replace(/^assets\//, '').replace(/\.png$/, ''),
  url: urls['../../' + e.path],
}))

const sheets = new Set(assets.filter(a => a.frames > 1).map(a => a.key))

export function preloadAssets(scene: Phaser.Scene) {
  for (const a of assets) {
    if (!a.url) { // the generator may list art before it is written: skip it, the game falls back
      console.warn(`asset missing on disk: ${a.path}`)
      continue
    }
    if (a.frames > 1) scene.load.spritesheet(a.key, a.url, { frameWidth: a.frameWidth, frameHeight: a.frameHeight })
    else scene.load.image(a.key, a.url)
  }
}

export function createAnims(scene: Phaser.Scene) {
  for (const a of assets) {
    if (a.frames > 1 && a.fps > 0 && scene.textures.exists(a.key)) {
      scene.anims.create({ key: a.key, frames: scene.anims.generateFrameNumbers(a.key), frameRate: a.fps, repeat: a.loop ? -1 : 0 })
    }
  }
}

// First frame of a strip, or the whole image.
export function firstFrame(key: string) {
  return sheets.has(key) ? 0 : undefined
}
