import Phaser from 'phaser'
import manifest from '../../assets/audio/manifest.json'

// Runtime audio from assets/audio/manifest.json. Only these folders ship: the WAV sources and the
// preview page stay out of the build. Player themes are for listening, not used in a match.
const urls = import.meta.glob<string>('../../assets/audio/{music,sfx,ambience}/*.{ogg,mp3}', {
  eager: true, query: '?url', import: 'default',
})

type Asset = (typeof manifest.assets)[number]
const byKey = new Map<string, Asset>(manifest.assets.map(a => [a.key, a]))
const url = (path: string) => urls['../../' + path]
const mix = manifest.mixing

// Settings live per browser.
export type Settings = { master: number; music: number; sfx: number; muted: boolean }
const STORE = 'taskvotan.audio'
export const settings: Settings = { master: mix.master, music: mix.music, sfx: mix.sfx, muted: false }
try { Object.assign(settings, JSON.parse(localStorage.getItem(STORE) ?? '{}')) } catch { /* storage blocked */ }
const listeners = new Set<() => void>()
export function saveSettings() {
  try { localStorage.setItem(STORE, JSON.stringify(settings)) } catch { /* storage blocked */ }
  for (const f of listeners) f()
}
const level = (group: 'music' | 'sfx') => settings.muted ? 0 : settings.master * settings[group]

export function preloadAudio(scene: Phaser.Scene) {
  for (const a of manifest.assets) {
    if (a.key.startsWith('player-') || a.key === 'lobby' || a.key === 'sprint') continue
    const files = [url(a.path), url(a.fallback)].filter(Boolean)
    if (files.length) scene.load.audio(a.key, files)
  }
}

// Lobby music plays outside Phaser (the game only exists during a match).
let lobby: HTMLAudioElement | undefined
export function lobbyMusic(on: boolean) {
  const a = byKey.get('lobby')
  if (!a) return
  lobby ??= Object.assign(new Audio(url(a.path) ?? url(a.fallback)), { loop: true })
  lobby.volume = Math.min(1, level('music') * a.gain / mix.music)
  if (on) lobby.play().catch(() => { /* not allowed before the first click; retried on the next */ })
  else lobby.pause()
}
listeners.add(() => { if (lobby && !lobby.paused) lobbyMusic(true) })

type Opts = { x?: number; mine?: boolean; critical?: boolean; rate?: number; volume?: number }

// Sounds for one match scene. Every event plays once; loops are owned by an id.
export class SceneAudio {
  private loops = new Map<string, Phaser.Sound.BaseSound>()
  private last = new Map<string, number>() // throttles
  private voices = 0

  constructor(private scene: Phaser.Scene) {
    listeners.add(this.apply)
    scene.events.once(Phaser.Scenes.Events.SHUTDOWN, () => { listeners.delete(this.apply); this.stopAll() })
  }

  has(key: string) { return this.scene.cache.audio.exists(key) }

  // play a one-shot. Teammates are quieter and panned; footsteps and teammates drop first when busy.
  play(key: string, o: Opts = {}) {
    const a = byKey.get(key)
    if (!a || !this.has(key) || this.scene.sound.locked) return // no queued burst after an unlock
    const critical = o.critical || (a.priority ?? 9) <= 1
    if (this.voices >= mix.maxSfxVoices - (critical ? 0 : mix.reservedCriticalVoices)) return
    const mine = o.mine ?? true
    const pan = o.x === undefined ? 0 : Phaser.Math.Clamp((o.x / 1920) * 2 - 1, -1, 1) * 0.5
    const s = this.scene.sound.add(key, {
      volume: level('sfx') * a.gain * (mine || critical ? 1 : mix.remotePlayerGain) * (o.volume ?? 1),
      rate: o.rate ?? 1,
    }) as Phaser.Sound.WebAudioSound
    if ('setPan' in s) s.setPan(pan)
    this.voices++
    s.once(Phaser.Sound.Events.COMPLETE, () => { this.voices--; s.destroy() })
    s.play()
    return s
  }

  // throttled plays at most once per ms for the id.
  throttled(id: string, ms: number, key: string, o?: Opts) {
    const now = performance.now()
    if (now - (this.last.get(id) ?? -Infinity) < ms) return
    this.last.set(id, now)
    this.play(key, o)
  }

  loop(id: string, key: string, group: 'music' | 'sfx' = 'sfx') {
    if (this.loops.has(id) || !this.has(key)) return
    const a = byKey.get(key)!
    const s = this.scene.sound.add(key, { loop: true, volume: level(group) * a.gain / (group === 'music' ? mix.music : 1) })
    this.loops.set(id, s)
    s.play() // while locked, Phaser starts it on unlock
  }

  stop(id: string) {
    const s = this.loops.get(id)
    if (s) { s.stop(); s.destroy(); this.loops.delete(id) }
  }

  hasLoop(id: string) { return this.loops.has(id) }

  // stopOne ends a one-shot that may still be playing (P1 charge).
  stopOne(s?: Phaser.Sound.BaseSound) { if (s?.isPlaying) s.stop() }

  duck(id: string, factor: number) {
    const s = this.loops.get(id) as Phaser.Sound.WebAudioSound | undefined
    const a = s && byKey.get(s.key)
    if (s && a) s.setVolume(level('music') * a.gain / mix.music * factor)
  }

  stopAll() { for (const id of [...this.loops.keys()]) this.stop(id) }

  private apply = () => {
    for (const s of this.loops.values()) {
      const a = byKey.get(s.key)
      if (!a) continue
      const group = a.category === 'sfx' ? 'sfx' : 'music'
      ;(s as Phaser.Sound.WebAudioSound).setVolume(level(group) * a.gain / (group === 'music' ? mix.music : 1))
    }
  }
}
