import Phaser from 'phaser'
import { createAnims, firstFrame, preloadAssets } from './assets'
import { SceneAudio, preloadAudio } from './audio'
import { send, world, type Input, type PlayerState, type Snapshot } from './net'

export const VIEW_W = 1920
export const VIEW_H = 1080

const SLOT_COLORS = ['teal', 'violet', 'amber', 'mint', 'orange']
const SLOT_HEX = ['#4FD1C5', '#B08CF0', '#FFD25E', '#8EF0B5', '#FFA25E']
const PORTRAITS = ['ui/portrait', 'ui/portrait-violet', 'ui/portrait-amber', 'ui/portrait-mint', 'ui/portrait-orange']
// Slot skills (dev/game-scenario.md, "Năm chiêu riêng"). Icons may not exist yet.
const SKILLS = [
  { name: 'Em sửa có một dòng', icon: 'ui/skills/one-line-fix', quote: 'Một dòng thôi, sao production tối rồi?' },
  { name: 'Không thuộc scope em', icon: 'ui/skills/out-of-scope', quote: 'Em xin phép loop anh vào xử lý.' },
  { name: 'Theo biên bản đã chốt', icon: 'ui/skills/meeting-minutes', quote: 'Anh đã đồng ý trong cuộc họp không có anh.' },
  { name: 'Máy em vẫn chạy', icon: 'ui/skills/works-on-my-machine', quote: 'Để em zip nguyên máy gửi anh.' },
  { name: 'Em về đúng giờ', icon: 'ui/skills/clock-out', quote: 'Task mai tính, cửa đang đóng!' },
]
const PINGS = ['', 'Tôi nhận', 'Cần hỗ trợ', 'Chặn scope']
const PLAYER_HP = 5 // server playerHP
const BOSS = 'sprites/boss/doc-cong/'
const SKILL_VFX = 'sprites/vfx/skills/'
// Boss states without their own art yet borrow the closest existing animation.
const BOSS_FALLBACK: Record<string, string> = { walk: 'idle', charge_windup: 'wave_windup', charge: 'idle', charge_recover: 'wind_clock' }
const CORAL = '#FF5A5A', CREAM = '#F2E3C2', CYAN = '#3FE0FF', OUTLINE = '#0B1020'
// Render this far behind the newest snapshot: two snapshot intervals at 30 Hz.
const RENDER_DELAY_MS = 70
// Own skill panel, top left under the five player chips, clear of the play area.
const SKILL_Y = 24 + 5 * 44 + 8

type Placed = Phaser.GameObjects.GameObject & { setPosition(x: number, y: number): unknown; setVisible(v: boolean): unknown }

type Avatar = {
  slot: number
  sprite: Phaser.GameObjects.Sprite
  label: Phaser.GameObjects.Text
  ping: Phaser.GameObjects.Text
  bar: Phaser.GameObjects.Graphics // revive progress or P1 charge
  fx?: Placed // effect following the player during a held skill
  fxAction?: string
  step: number // next footstep time
}

type Chip = { portrait: Phaser.GameObjects.Image; name: Phaser.GameObjects.Text; hearts: Phaser.GameObjects.Image[] }

const text = (s: Phaser.Scene, x: number, y: number, size: number, color: string) =>
  s.add.text(x, y, '', { fontFamily: 'monospace', fontSize: `${size}px`, fontStyle: 'bold', color, stroke: OUTLINE, strokeThickness: 4 })

export class GameScene extends Phaser.Scene {
  private levelUrl = ''
  private floorY = 0
  private avatars = new Map<number, Avatar>()
  private shots = new Map<number, Phaser.GameObjects.Sprite>()
  private waves = new Map<number, Phaser.GameObjects.Sprite>()
  private items = new Map<number, Phaser.GameObjects.Sprite>()
  private zones = new Map<number, Placed>()
  private boss!: Phaser.GameObjects.Sprite
  private bossMark!: Placed
  private bossLine!: Phaser.GameObjects.Text
  private bossStep = 0
  private lastBossX?: number
  private lastBossDust = -Infinity
  private warning!: Phaser.GameObjects.TileSprite // deadline wave
  private chargeWarning!: Phaser.GameObjects.TileSprite // locked charge lane
  private hud!: {
    deadline: Phaser.GameObjects.Graphics; deadlineText: Phaser.GameObjects.Text
    bossBar: Phaser.GameObjects.Graphics; result: Phaser.GameObjects.Text
    skillName: Phaser.GameObjects.Text; skillBar: Phaser.GameObjects.Graphics; skillIcon?: Phaser.GameObjects.Image
  }
  private chips: Chip[] = []
  private keys!: Record<string, Phaser.Input.Keyboard.Key>
  private lastInput = ''
  private sfx!: SceneAudio
  private prev?: Snapshot // last snapshot whose events were played
  private charging?: Phaser.Sound.BaseSound
  private urgent = false
  private lastSec = -1
  private quotes: Phaser.GameObjects.Text[] = []

  constructor() { super('game') }

  init(data: { level: string }) { this.levelUrl = data.level }

  preload() {
    preloadAssets(this)
    preloadAudio(this)
    this.load.tilemapTiledJSON('level', this.levelUrl)
  }

  create() {
    createAnims(this)
    this.sfx = new SceneAudio(this)
    const map = this.make.tilemap({ key: 'level' })
    // The whole arena fits the view, so the camera stays put, map bottom on view bottom.
    this.cameras.main.setScroll(0, map.heightInPixels - VIEW_H)
    const objects = map.getObjectLayer('objects')?.objects ?? []
    this.floorY = objects.find(o => o.type === 'spawn')?.y ?? map.heightInPixels

    // Background art is drawn standing on its bottom edge: put that edge on the level floor.
    for (const bg of ['bg-0-sky', 'bg-1-windows', 'bg-2-office']) {
      this.add.image(0, this.floorY - this.cameras.main.scrollY, `backgrounds/${bg}`).setOrigin(0, 1).setScrollFactor(0)
    }
    const tiles = map.addTilesetImage('office', 'tiles/office')!
    for (const layer of map.layers) map.createLayer(layer.name, tiles)

    for (const o of objects) {
      if (o.type === 'prop') {
        const key = `sprites/props/${o.name}`
        this.add.image(o.x!, o.y!, key, firstFrame(key)).setOrigin(0.5, 1)
      }
    }
    const bossAt = objects.find(o => o.type === 'boss')
    this.boss = this.add.sprite(bossAt?.x ?? 0, bossAt?.y ?? 0, BOSS + 'idle').setOrigin(0.5, 1).setDepth(1).play(BOSS + 'idle')
    this.bossMark = this.fx('boss-stamp-mark', 0, 0, () => this.add.circle(0, 0, 18).setStrokeStyle(3, 0xffd25e))
    this.bossMark.setVisible(false)
    this.bossLine = text(this, 0, 0, 16, CORAL).setOrigin(0.5, 1).setDepth(6)
    // Floor telegraphs sit on the floor line only, never over platform edges.
    this.warning = this.add.tileSprite(map.tileWidth, this.floorY, map.widthInPixels - 2 * map.tileWidth, 8, 'sprites/vfx/floor-warning', 0)
      .setOrigin(0, 1).setVisible(false)
    const lane = this.textures.exists('sprites/vfx/charge-warning') ? 'sprites/vfx/charge-warning' : 'sprites/vfx/floor-warning'
    this.chargeWarning = this.add.tileSprite(0, this.floorY, 64, lane.endsWith('charge-warning') ? 16 : 8, lane, 0)
      .setOrigin(0.5, 1).setDepth(2).setVisible(false)

    this.hud = {
      deadline: this.add.graphics().setScrollFactor(0).setDepth(10),
      deadlineText: text(this, VIEW_W / 2, 44, 18, CREAM).setOrigin(0.5, 0).setScrollFactor(0).setDepth(11),
      bossBar: this.add.graphics().setScrollFactor(0).setDepth(10),
      result: text(this, VIEW_W / 2, VIEW_H / 2 - 80, 56, CREAM).setOrigin(0.5).setScrollFactor(0).setDepth(20).setAlign('center'),
      skillName: text(this, 64, SKILL_Y, 16, CREAM).setScrollFactor(0).setDepth(11),
      skillBar: this.add.graphics().setScrollFactor(0).setDepth(10),
    }
    this.add.image(VIEW_W / 2 - 316, 30, 'ui/deadline-icon').setScale(2).setScrollFactor(0).setDepth(11)
    text(this, VIEW_W / 2, VIEW_H - 64, 16, CREAM).setText('ĐỐC CÔNG HÔM QUA').setOrigin(0.5, 1).setScrollFactor(0).setDepth(11)

    this.keys = this.input.keyboard!.addKeys(
      'W,A,S,D,UP,LEFT,DOWN,RIGHT,SPACE,J,K,L,SHIFT,E,R,ONE,TWO,THREE',
    ) as Record<string, Phaser.Input.Keyboard.Key>
    this.keys.R.on('down', () => this.trySkill())

    this.startMusic()
  }

  update(time: number) {
    this.sendInput()
    const pair = this.pair()
    if (!pair) return
    const [a, b, k] = pair
    if (b !== this.prev) {
      if (this.prev) this.playEvents(this.prev, b)
      this.prev = b
    }
    this.drawPlayers(a, b, k, time)
    this.drawBoss(b, time)
    this.drawHazards(a, b, k)
    this.drawHud(b, time)
  }

  private sendInput() {
    const k = this.keys
    const input: Input = {
      left: k.A.isDown || k.LEFT.isDown,
      right: k.D.isDown || k.RIGHT.isDown,
      down: k.S.isDown || k.DOWN.isDown,
      jump: k.SPACE.isDown || k.W.isDown || k.UP.isDown,
      attack: k.J.isDown,
      skill: k.K.isDown,
      special: k.R.isDown, // E remains the existing teammate-revive interaction.
      dash: k.L.isDown || k.SHIFT.isDown,
      interact: k.E.isDown,
      ping: k.ONE.isDown ? 1 : k.TWO.isDown ? 2 : k.THREE.isDown ? 3 : 0,
    }
    const s = JSON.stringify(input)
    if (s !== this.lastInput) {
      this.lastInput = s
      send({ type: 'input', ...input })
    }
  }

  // Local cue when the skill is not ready. The server ignores the press.
  private trySkill() {
    const me = this.prev?.players.find(p => p.id === world.myId)
    if (me && me.skillMs > 0) this.sfx.throttled('unavailable', 300, 'skill-unavailable')
  }

  // The two snapshots around the render time, and how far between them it is.
  private pair(): [Snapshot, Snapshot, number] | undefined {
    const snaps = world.snapshots
    if (snaps.length === 0) return
    const t = performance.now() - world.clockOffset - RENDER_DELAY_MS // in server time
    while (snaps.length > 2 && snaps[1].serverTime <= t) snaps.shift()
    const a = snaps[0]
    const b = snaps[1] ?? a
    return [a, b, b === a ? 1 : Phaser.Math.Clamp((t - a.serverTime) / (b.serverTime - a.serverTime), 0, 1)]
  }

  private startMusic() {
    this.urgent = false
    this.sfx.loop('music', 'boss', 'music')
    this.sfx.loop('ambience', 'office', 'music')
  }

  private stopMatchLoops() {
    for (const id of ['music', 'ambience', 'urgency', 'wind']) this.sfx.stop(id)
    for (const z of this.zones.keys()) this.sfx.stop(`zone${z}`)
  }

  // playEvents plays every sound for what changed between two snapshots, once. A reconnect starts
  // from the current snapshot, so old hits, heals and casts are never replayed.
  private playEvents(was: Snapshot, now: Snapshot) {
    const s = this.sfx
    for (const p of now.players) {
      const q = was.players.find(o => o.id === p.id)
      if (!q) continue
      const mine = p.id === world.myId
      const o = { x: p.x, mine }
      const up = (k: keyof PlayerState) => (p[k] as number) > (q[k] as number)
      const slot = this.slotOf(p.id)
      if (up('attacks')) s.play('keyboard-swing', o)
      if (up('hits') && !(slot === 0 && up('skillHits'))) s.play('keyboard-hit', o)
      if (up('pulses')) s.play('code-pulse', o)
      if (up('dashes')) s.play('dash', o)
      if (up('coffees')) s.play('coffee', o)
      if (p.action === 'down' && q.action !== 'down') s.play('player-down', { ...o, critical: true })
      else if (up('hurts')) s.play('player-hurt', o)
      if (p.action === 'revive' && q.action !== 'revive') s.play('revive', o)
      if (p.ping && p.ping !== q.ping) s.throttled(`ping${p.id}`, 1000, 'ping', o)
      if (!p.ground && q.ground && p.vy < 0 && !p.away) s.play('jump', o)
      if (p.ground && !q.ground && p.action === '') s.play('land', o)
      if (mine && q.skillMs > 0 && p.skillMs === 0) s.play('skill-ready')
      if (up('casts')) this.flashQuote(p)

      switch (slot) {
        case 0: // the charge one-shot stops on release or cancel, never loops
          if (p.action === 'charge' && q.action !== 'charge') {
            const c = s.play('teal-charge', o)
            if (mine) this.charging = c
          }
          if (up('casts')) { if (mine) s.stopOne(this.charging); s.play('teal-smash', o) }
          if (up('skillHits')) s.play('teal-smash-hit', o)
          if (up('cancels')) { if (mine) s.stopOne(this.charging); s.play('teal-charge-cancel', o) }
          break
        case 1:
          if (up('casts')) s.play('violet-shield', o)
          if (up('reflects')) s.play('violet-reflect', o) // one per snapshot, however many bounced
          if (q.action === 'shield' && p.action !== 'shield') s.play('violet-shield-end', o)
          break
        case 2:
          if (up('casts')) s.play('amber-stamp-cast', o)
          break
        case 3:
          if (up('casts')) s.play('mint-hotfix-deploy', o)
          break
        case 4:
          if (up('casts')) s.play('orange-timecard-dash', o)
          if (up('skillHits')) s.play('orange-timecard-hit', o)
          if (q.action === 'timecard' && p.action !== 'timecard') s.play('orange-timecard-stop', o)
          break
      }
    }
    // One heal cue per snapshot, however many players were healed.
    if (now.players.some(p => p.healed > (was.players.find(o => o.id === p.id)?.healed ?? p.healed))) s.play('mint-hotfix-heal')

    // Boss. Its counters reset with each attempt: only growth is an event.
    const b = now.boss, c = was.boss
    if (b.marks > c.marks) s.play('amber-mark-hit', { x: b.x })
    if (c.markMs > 0 && b.markMs === 0 && b.marks === c.marks && now.result === '') s.play('amber-mark-expire', { x: b.x, volume: 0.6 })
    if (b.hits > c.hits) s.throttled('boss-hurt', 120, 'boss-hurt', { x: b.x })
    if (b.state !== c.state) {
      const crit = { x: b.x, critical: true }
      switch (b.state) {
        case 'volley_windup': s.play('boss-volley-windup', crit); break
        case 'volley': s.play('boss-volley', crit); break
        case 'wave_windup': s.play('boss-wave-windup', crit); break
        case 'wave': s.play('boss-wave-slam', crit); break
        case 'charge_windup':
          s.play(s.has('boss-charge-windup') ? 'boss-charge-windup' : 'checkpoint', crit) // placeholder bell until the sound exists
          this.say('Có 5 phút thôi, xuống đây anh trao đổi!')
          break
        case 'charge': s.play(s.has('boss-charge') ? 'boss-charge' : 'dash', crit); break
        case 'defeat': s.play('boss-defeat', crit); break
      }
      if (b.state === 'wind_clock') s.loop('wind', 'boss-wind-clock')
      else s.stop('wind')
      if (b.state.endsWith('windup')) { // duck the music briefly under a telegraph
        s.duck('music', 0.55)
        this.time.delayedCall(500, () => s.duck('music', this.urgent ? 0.7 : 1))
      }
    }

    // Match result and DEADLINE cues.
    if (now.result !== was.result) {
      if (now.result === 'won') {
        this.stopMatchLoops()
        this.time.delayedCall(600, () => s.play('victory', { critical: true })) // after boss-defeat
      } else if (now.result === 'lost') {
        this.stopMatchLoops()
        s.play('defeat', { critical: true })
      } else this.startMusic() // the server restarted the level
    }
    if (now.result === '') {
      const left = now.deadlineMs
      if (!this.urgent && left > 0 && left < Math.min(60_000, now.deadlineOf * 0.1)) {
        this.urgent = true
        s.loop('urgency', 'urgency', 'music')
        s.duck('music', 0.7)
      }
      const sec = Math.ceil(left / 1000)
      if (left > 0 && left <= 10_000 && sec !== this.lastSec) s.play('deadline-tick', { critical: true })
      this.lastSec = sec
    }
  }

  private slotOf(id: number) { return world.lobby.find(l => l.id === id)?.slot ?? 0 }

  private drawPlayers(a: Snapshot, b: Snapshot, k: number, time: number) {
    const seen = new Set<number>()
    for (const pb of b.players) {
      const pa = a.players.find(p => p.id === pb.id) ?? pb
      seen.add(pb.id)
      this.drawPlayer(pb, pa, Phaser.Math.Linear(pa.x, pb.x, k), Phaser.Math.Linear(pa.y, pb.y, k), time)
    }
    for (const [id, av] of this.avatars) {
      if (!seen.has(id)) {
        for (const o of [av.sprite, av.label, av.ping, av.bar, av.fx]) o?.destroy()
        this.avatars.delete(id)
      }
    }
    this.unstackLabels()
  }

  // Players standing close together: lift each label above the ones it would overlap.
  private unstackLabels() {
    const placed: Phaser.Geom.Rectangle[] = []
    const avs = [...this.avatars.values()].sort((a, b) => b.label.y - a.label.y)
    for (const av of avs) {
      let box = av.label.getBounds()
      while (placed.some(p => Phaser.Geom.Intersects.RectangleToRectangle(p, box))) {
        av.label.y -= 16
        box = av.label.getBounds()
      }
      placed.push(box)
      av.ping.y = av.label.y - 18
    }
  }

  private drawPlayer(p: PlayerState, last: PlayerState, x: number, y: number, time: number) {
    const info = world.lobby.find(l => l.id === p.id)
    const slot = info?.slot ?? 0
    let av = this.avatars.get(p.id)
    if (av?.slot !== slot) { // new player, or the lobby info arrived after the first snapshot
      if (av) for (const o of [av.sprite, av.label, av.ping, av.bar, av.fx]) o?.destroy()
      av = {
        slot, step: 0,
        sprite: this.add.sprite(x, y, `sprites/player/${SLOT_COLORS[slot]}/idle`).setOrigin(0.5, 1).setDepth(2),
        label: text(this, x, y, 14, SLOT_HEX[slot]).setOrigin(0.5, 1).setDepth(3),
        ping: text(this, x, y, 16, CREAM).setOrigin(0.5, 1).setDepth(3),
        bar: this.add.graphics().setDepth(3),
      }
      this.avatars.set(p.id, av)
    }
    const s = av.sprite
    const anim = (name: string) => `sprites/player/${SLOT_COLORS[slot]}/${name}`
    // Skill art may not exist yet: fall back to the closest base animation.
    const play = (name: string, fallback: string, restart = false) =>
      s.play(this.anims.exists(anim(name)) ? anim(name) : anim(fallback), !restart)

    // One-shots restart on a counter change; otherwise keep whatever is playing.
    const changed = (k: keyof PlayerState) => p[k] !== last[k]
    if (changed('casts')) {
      if (p.action === 'shield') play('skill-windup', 'attack', true) // the braced guard holds the windup pose
      else play('skill-cast', p.action === 'timecard' ? 'dash' : 'attack', true)
    } else if (changed('attacks') || changed('pulses')) {
      play('attack', 'attack', true)
      if (changed('attacks')) this.vfx('sprites/vfx/swing-arc', x + p.facing * 22, y - 22, p.facing < 0)
    } else if (changed('dashes')) play('dash', 'dash', true)
    else if (changed('hurts')) {
      play('hurt', 'hurt', true)
      s.setTintFill(0xff5a5a)
      this.time.delayedCall(100, () => s.clearTint())
    } else if (p.action === 'revive' && last.action !== 'revive') play('revive', 'revive', true)
    else if (p.action === 'down') play('down', 'down')
    else if (p.action === 'charge') play('skill-windup', 'attack')
    else if (p.action === '') play(!p.ground ? (p.vy < 0 ? 'jump' : 'fall') : p.vx !== 0 ? 'run' : 'idle', 'idle')

    if (p.ground && !last.ground && p.action === '') this.vfx('sprites/vfx/land-dust', x, y, false, 0.5, 1)
    if (p.ground && p.vx !== 0 && p.action === '' && time >= av.step) { // footsteps: lowest priority
      av.step = time + 280
      this.sfx.play(Math.floor(time / 280) % 2 ? 'footstep-1' : 'footstep-2', { x, mine: p.id === world.myId, volume: 0.6 })
    }
    if (changed('skillHits') && slot === 4) this.fxOnce('timecard-impact', x + p.facing * 24, y - 22, false, () => {})
    if (changed('casts') && slot === 0) {
      this.fxOnce('keyboard-cleave', x + p.facing * 34, y - 30, p.facing < 0,
        () => this.vfx('sprites/vfx/swing-arc', x + p.facing * 30, y - 26, p.facing < 0)?.setScale(1.8))
    }
    this.attachedFx(av, p, x, y)

    s.setPosition(x, y).setFlipX(p.facing < 0).setAlpha(p.away ? 0.4 : 1)
    const me = p.id === world.myId ? '▼ ' : ''
    const top = Math.round(y) - 60 // above the tallest (54 px) design
    av.label.setText(`${me}P${slot + 1} ${info?.name ?? ''}`).setPosition(Math.round(x), top)
    av.ping.setText(PINGS[p.ping] ? `「${PINGS[p.ping]}」` : '').setPosition(Math.round(x), top - 18)

    av.bar.clear()
    const bar = (frac: number, color: number) => {
      av.bar.fillStyle(0x0b1020).fillRect(x - 20, y - 66, 40, 6)
      av.bar.fillStyle(color).fillRect(x - 19, y - 65, 38 * frac, 4)
    }
    if (p.action === 'down') {
      bar(p.revive, 0x8ef0b5)
      av.label.setText(`${av.label.text} · Nghỉ phép`)
    } else if (p.action === 'charge') bar(p.charge, 0x3fe0ff)
  }

  // attachedFx keeps one effect on a player while a held skill lasts (P1 charge, P2 shield, P5 lunge).
  private attachedFx(av: Avatar, p: PlayerState, x: number, y: number) {
    const want = p.action === 'charge' || p.action === 'shield' || p.action === 'timecard' ? p.action : undefined
    if (av.fxAction !== want) {
      av.fx?.destroy()
      av.fx = undefined
      av.fxAction = want
      if (want === 'charge') av.fx = this.fx('keyboard-charge', x, y, () => this.add.circle(0, 0, 20, 0x3fe0ff, 0.25))
      if (want === 'shield') av.fx = this.fx('forward-shield', x, y, () => this.add.rectangle(0, 0, 10, 56).setStrokeStyle(3, 0x3fe0ff))
      if (want === 'timecard') av.fx = this.fx('timecard-trail', x, y, () => this.add.rectangle(0, 0, 96, 6, 0xf2e3c2, 0.6))
    }
    if (!av.fx || !want) return
    const f = p.facing
    const pos: Record<string, [number, number]> = { charge: [x, y - 50], shield: [x + f * 22, y - 30], timecard: [x - f * 48, y - 22] }
    av.fx.setPosition(...pos[want])
    if (av.fx instanceof Phaser.GameObjects.Sprite) av.fx.setFlipX(f < 0)
  }

  private drawBoss(b: Snapshot, time: number) {
    const boss = b.boss
    this.boss.setPosition(boss.x, boss.y).setFlipX(boss.facing < 0) // art faces right
    const own = BOSS + boss.state
    this.boss.play(this.anims.exists(own) ? own : BOSS + (BOSS_FALLBACK[boss.state] ?? 'idle'), true)
    if (this.prev && b !== this.prev && boss.hits > this.prev.boss.hits) {
      this.boss.setTintFill(0xffffff)
      this.time.delayedCall(60, () => this.boss.clearTint())
      this.vfx('sprites/vfx/hit-spark', boss.x + Phaser.Math.Between(-24, 24), boss.y - Phaser.Math.Between(40, 90))
    }
    this.bossMark.setVisible(boss.markMs > 0)
    this.bossMark.setPosition(boss.x, boss.y - 132)
    this.bossLine.setPosition(boss.x, boss.y - 150)

    this.warning.setVisible(boss.state === 'wave_windup').setFrame(Math.floor(time / 166) % 2)
    const from = boss.chargeFrom, to = boss.chargeTo
    const showCharge = b.result === '' && (boss.state === 'charge_windup' || boss.state === 'charge')
      && from !== undefined && to !== undefined && from !== to
    this.chargeWarning.setVisible(showCharge)
    if (showCharge && from !== undefined && to !== undefined) {
      // The server sends centre positions; include the 70px-wide boss body at both ends.
      const frames = Math.max(1, this.chargeWarning.texture.frameTotal - 1)
      this.chargeWarning.setPosition((from + to) / 2, this.floorY)
        .setSize(Math.abs(to - from) + 70, this.chargeWarning.height)
        .setFlipX(to < from).setFrame(Math.floor(time / 125) % frames)
    }
    const moving = this.lastBossX !== undefined && Math.abs(boss.x - this.lastBossX) > 0.1
    if (b.result === '' && moving && (boss.state === 'walk' || boss.state === 'charge')
      && time - this.lastBossDust >= (boss.state === 'charge' ? 120 : 280)) {
      this.vfx('sprites/vfx/boss-foot-dust', boss.x - boss.facing * 30, boss.y, boss.facing < 0, 0.5, 1)?.setDepth(1)
      this.lastBossDust = time
    }
    if (boss.state === 'walk' && moving && time >= this.bossStep) {
      this.bossStep = time + 450
      this.sfx.play(this.sfx.has('boss-footstep') ? 'boss-footstep' : 'footstep-1', { x: boss.x, rate: 0.6, volume: 0.5 })
    }
    this.lastBossX = boss.x
  }

  private say(line: string) {
    this.bossLine.setText(`“${line}”`)
    this.time.delayedCall(1800, () => this.bossLine.setText(''))
  }

  private flashQuote(p: PlayerState) {
    const av = this.avatars.get(p.id)
    if (!av) return
    const t = text(this, av.sprite.x, av.label.y - 22, 14, CREAM).setText(SKILLS[av.slot].quote).setOrigin(0.5, 1).setDepth(6)
    // Several casts at once: lift this quote above the ones still showing.
    while (this.quotes.some(q => Phaser.Geom.Intersects.RectangleToRectangle(q.getBounds(), t.getBounds()))) t.y -= 20
    this.quotes.push(t)
    this.tweens.add({
      targets: t, y: t.y - 16, alpha: 0, delay: 900, duration: 600,
      onComplete: () => { this.quotes = this.quotes.filter(q => q !== t); t.destroy() },
    })
  }

  private drawHazards(a: Snapshot, b: Snapshot, k: number) {
    const shotArt: Record<string, [string, string, number?]> = { // own art, fallback art, fallback tint
      card: ['sprites/vfx/task-card', 'sprites/vfx/task-card'],
      pulse: ['sprites/vfx/code-pulse', 'sprites/vfx/code-pulse'],
      stamp: [SKILL_VFX + 'stamp-projectile', 'sprites/vfx/code-pulse', 0xffd25e],
      returned: [SKILL_VFX + 'returned-task', 'sprites/vfx/task-card', 0x3fe0ff], // own art replaces the tint once it exists
    }
    sync(this.shots, b.shots, s => {
      const [own, fallback, tint] = shotArt[s.kind]
      const key = this.textures.exists(own) ? own : fallback
      const sp = this.add.sprite(s.x, s.y, key).setFlipX(s.vx < 0).setDepth(4)
      if (this.anims.exists(key)) sp.play(key)
      if (key !== own && tint) sp.setTint(tint)
      return sp
    }, (sprite, s) => {
      const sa = a.shots.find(o => o.id === s.id) ?? s
      sprite.setPosition(Phaser.Math.Linear(sa.x, s.x, k), s.y)
    })
    sync(this.items, b.items, it => {
      const key = 'sprites/items/coffee'
      return this.add.sprite(it.x, it.y, key).setOrigin(0.5, 1).play(key).setDepth(1)
    }, () => {})
    sync(this.waves, b.waves, w => {
      const key = 'sprites/vfx/deadline-wave'
      return this.add.sprite(w.x, this.floorY, key).setOrigin(0.5, 1).play(key).setFlipX(w.dir < 0).setDepth(4)
    }, (sprite, w) => {
      const wa = a.waves.find(o => o.id === w.id) ?? w
      sprite.setPosition(Phaser.Math.Linear(wa.x, w.x, k), this.floorY)
    })

    // Hotfix zones: the loop belongs to the zone id, so a reconnect restores it without replaying the deploy.
    const live = new Set(b.zones.map(z => z.id))
    for (const z of b.zones) {
      if (this.zones.has(z.id)) continue
      this.zones.set(z.id, this.fx('hotfix-zone', z.x, z.y - 24,
        () => this.add.rectangle(0, 0, 128, 48, 0x8ef0b5, 0.25).setStrokeStyle(2, 0x8ef0b5)))
      if (this.prev) this.fxOnce('hotfix-burst', z.x, z.y - 24, false, () => {})
      if (b.result === '') this.sfx.loop(`zone${z.id}`, 'mint-hotfix-zone')
    }
    for (const [id, zone] of this.zones) {
      if (live.has(id)) continue
      zone.destroy()
      this.zones.delete(id)
      this.sfx.stop(`zone${id}`)
      if (b.result === '') this.sfx.play('mint-hotfix-end', { volume: 0.7 })
    }
  }

  private drawHud(b: Snapshot, time: number) {
    const { deadline, deadlineText, bossBar, result } = this.hud
    const left = Math.max(0, b.deadlineMs) / Math.max(1, b.deadlineOf)
    const w = 600, x = VIEW_W / 2 - w / 2
    deadline.clear().fillStyle(0x0b1020).fillRect(x - 2, 20, w + 4, 20)
    deadline.fillStyle(left < 0.2 ? 0xff5a5a : 0x3fe0ff).fillRect(x, 22, w * left, 16)
    const secs = Math.ceil(Math.max(0, b.deadlineMs) / 1000)
    const flash = b.deadlineMs > 0 && b.deadlineMs <= 10_000 && Math.floor(time / 250) % 2 === 0 // with the last-10s tick
    deadlineText.setText(`DEADLINE ${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`).setColor(flash ? CORAL : CREAM)

    const boss = b.boss
    const bw = 800, bx = VIEW_W / 2 - bw / 2
    const open = boss.state === 'wind_clock' || boss.state === 'charge_recover' // good moment to hit
    bossBar.clear().fillStyle(0x0b1020).fillRect(bx - 2, VIEW_H - 58, bw + 4, 18)
    bossBar.fillStyle(open ? 0xffd25e : 0xff5a5a).fillRect(bx, VIEW_H - 56, bw * (boss.hp / Math.max(1, boss.maxHp)), 14)

    this.drawChips(b)
    this.drawSkill(b)

    if (b.result === 'won') result.setColor(CYAN).setText('APPROVED!\nCả team được tan làm.')
    else if (b.result === 'lost') {
      result.setColor(CORAL).setText(b.deadlineMs <= 0 ? 'TRỄ DEADLINE\nLàm lại từ đầu…' : 'CẢ TEAM NGHỈ PHÉP\nLàm lại từ đầu…')
    } else result.setText('')
  }

  // Own skill, bottom left: icon (once the art exists), name, key and cooldown.
  private drawSkill(b: Snapshot) {
    const me = b.players.find(p => p.id === world.myId)
    if (!me) return
    const skill = SKILLS[this.slotOf(me.id)]
    const { skillName, skillBar } = this.hud
    if (!this.hud.skillIcon && this.textures.exists(skill.icon)) {
      this.hud.skillIcon = this.add.image(24, SKILL_Y, skill.icon).setOrigin(0).setScale(1.5).setScrollFactor(0).setDepth(11)
    }
    const ready = me.skillMs === 0
    skillName.setText(`[R] ${skill.name}${ready ? '' : `  ${(me.skillMs / 1000).toFixed(1)}s`}`).setColor(ready ? CREAM : '#8899AA')
    skillBar.clear().fillStyle(0x0b1020).fillRect(62, SKILL_Y + 24, 244, 10)
    skillBar.fillStyle(ready ? 0x3fe0ff : 0x2f4a7f).fillRect(64, SKILL_Y + 26, 240 * (1 - me.skillMs / Math.max(1, me.skillOf)), 6)
  }

  // Status chips, top left: portrait, P# name, hearts.
  private drawChips(b: Snapshot) {
    b.players.forEach((p, i) => {
      const slot = this.slotOf(p.id)
      let chip = this.chips[i]
      if (!chip) {
        const y = 24 + i * 44
        chip = {
          portrait: this.add.image(24, y, PORTRAITS[slot]).setOrigin(0).setScale(1.5).setScrollFactor(0).setDepth(10),
          name: text(this, 64, y - 2, 14, SLOT_HEX[slot]).setScrollFactor(0).setDepth(10),
          hearts: Array.from({ length: PLAYER_HP }, (_, h) =>
            this.add.image(64 + h * 22, y + 20, 'ui/heart', 0).setOrigin(0).setScale(2).setScrollFactor(0).setDepth(10)),
        }
        this.chips[i] = chip
      }
      const name = world.lobby.find(l => l.id === p.id)?.name ?? ''
      chip.portrait.setTexture(PORTRAITS[slot]).setAlpha(p.away ? 0.4 : 1)
      chip.name.setColor(SLOT_HEX[slot]).setText(`P${slot + 1} ${name}${p.away ? ' (mất kết nối)' : ''}`)
      chip.hearts.forEach((h, n) => h.setFrame(n < p.hp ? 0 : 1))
    })
    for (const chip of this.chips.splice(b.players.length)) {
      chip.portrait.destroy(); chip.name.destroy(); chip.hearts.forEach(h => h.destroy())
    }
  }

  // fx makes a looping skill effect from its art, or from a fallback shape until the art exists.
  private fx(name: string, x: number, y: number, fallback: () => Phaser.GameObjects.Shape): Placed {
    const key = SKILL_VFX + name
    if (!this.textures.exists(key)) return fallback().setPosition(x, y).setDepth(5)
    const s = this.add.sprite(x, y, key).setDepth(5)
    if (this.anims.exists(key)) s.play(key)
    return s
  }

  // fxOnce plays a one-shot skill effect, or the fallback.
  private fxOnce(name: string, x: number, y: number, flip: boolean, fallback: () => void) {
    const key = SKILL_VFX + name
    if (this.anims.exists(key)) this.vfx(key, x, y, flip)
    else fallback()
  }

  // vfx plays a one-shot effect; missing art plays nothing.
  private vfx(key: string, x: number, y: number, flip = false, ox = 0.5, oy = 0.5) {
    if (!this.anims.exists(key)) return
    const s = this.add.sprite(x, y, key).setOrigin(ox, oy).setFlipX(flip).setDepth(5).play(key)
    s.once(Phaser.Animations.Events.ANIMATION_COMPLETE, () => s.destroy())
    return s
  }
}

// sync keeps one sprite per server object id.
function sync<T extends { id: number }>(
  sprites: Map<number, Phaser.GameObjects.Sprite>, items: T[],
  create: (t: T) => Phaser.GameObjects.Sprite, update: (s: Phaser.GameObjects.Sprite, t: T) => void,
) {
  const seen = new Set<number>()
  for (const it of items) {
    seen.add(it.id)
    let s = sprites.get(it.id)
    if (!s) sprites.set(it.id, (s = create(it)))
    update(s, it)
  }
  for (const [id, s] of sprites) {
    if (!seen.has(id)) { s.destroy(); sprites.delete(id) }
  }
}
