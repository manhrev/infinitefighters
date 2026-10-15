import Phaser from 'phaser'
import { createAnims, firstFrame, preloadAssets } from './assets'
import { send, world, type Input, type PlayerState, type Snapshot } from './net'

export const VIEW_W = 1920
export const VIEW_H = 1080

const SLOT_COLORS = ['teal', 'violet', 'amber', 'mint', 'orange']
const SLOT_HEX = ['#4FD1C5', '#B08CF0', '#FFD25E', '#8EF0B5', '#FFA25E']
const PORTRAITS = ['ui/portrait', 'ui/portrait-violet', 'ui/portrait-amber', 'ui/portrait-mint', 'ui/portrait-orange']
const PINGS = ['', 'Tôi nhận', 'Cần hỗ trợ', 'Chặn scope']
const PLAYER_HP = 5 // server playerHP
const BOSS = 'sprites/boss/doc-cong/'
const CORAL = '#FF5A5A', CREAM = '#F2E3C2', CYAN = '#3FE0FF', OUTLINE = '#0B1020'
// Render this far behind the newest snapshot: two snapshot intervals at 30 Hz.
const RENDER_DELAY_MS = 70

type Avatar = {
  slot: number
  sprite: Phaser.GameObjects.Sprite
  label: Phaser.GameObjects.Text
  ping: Phaser.GameObjects.Text
  revive: Phaser.GameObjects.Graphics
  last?: PlayerState // previous rendered state, to spot counter changes and landings
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
  private boss!: Phaser.GameObjects.Sprite
  private bossHits = -1 // -1 until the first snapshot, so joining mid-match does not flash
  private warning!: Phaser.GameObjects.TileSprite
  private hud!: {
    deadline: Phaser.GameObjects.Graphics; deadlineText: Phaser.GameObjects.Text
    bossBar: Phaser.GameObjects.Graphics; result: Phaser.GameObjects.Text
  }
  private chips: Chip[] = []
  private keys!: Record<string, Phaser.Input.Keyboard.Key>
  private lastInput = ''

  constructor() { super('game') }

  init(data: { level: string }) { this.levelUrl = data.level }

  preload() {
    preloadAssets(this)
    this.load.tilemapTiledJSON('level', this.levelUrl)
  }

  create() {
    createAnims(this)
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
    this.boss = this.add.sprite(bossAt?.x ?? 0, bossAt?.y ?? 0, BOSS + 'idle').setOrigin(0.5, 1).play(BOSS + 'idle')
    // Shown under the floor line while the boss winds up a deadline wave.
    this.warning = this.add.tileSprite(map.tileWidth, this.floorY, map.widthInPixels - 2 * map.tileWidth, 8, 'sprites/vfx/floor-warning', 0)
      .setOrigin(0, 1).setVisible(false)

    this.hud = {
      deadline: this.add.graphics().setScrollFactor(0).setDepth(10),
      deadlineText: text(this, VIEW_W / 2, 18, 18, CREAM).setOrigin(0.5, 0).setScrollFactor(0).setDepth(11),
      bossBar: this.add.graphics().setScrollFactor(0).setDepth(10),
      result: text(this, VIEW_W / 2, VIEW_H / 2 - 80, 56, CREAM).setOrigin(0.5).setScrollFactor(0).setDepth(20).setAlign('center'),
    }
    this.add.image(VIEW_W / 2 - 316, 30, 'ui/deadline-icon').setScale(2).setScrollFactor(0).setDepth(11)
    text(this, VIEW_W / 2, VIEW_H - 64, 16, CREAM).setText('ĐỐC CÔNG HÔM QUA').setOrigin(0.5, 1).setScrollFactor(0).setDepth(11)

    this.keys = this.input.keyboard!.addKeys(
      'W,A,S,D,UP,LEFT,DOWN,RIGHT,SPACE,J,K,L,SHIFT,E,ONE,TWO,THREE',
    ) as Record<string, Phaser.Input.Keyboard.Key>
  }

  update(time: number) {
    this.sendInput()
    const pair = this.pair()
    if (!pair) return
    const [a, b, k] = pair
    this.drawPlayers(a, b, k)
    this.drawBoss(b, time)
    this.drawHazards(a, b, k)
    this.drawHud(b)
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

  private drawPlayers(a: Snapshot, b: Snapshot, k: number) {
    const seen = new Set<number>()
    for (const pb of b.players) {
      const pa = a.players.find(p => p.id === pb.id) ?? pb
      seen.add(pb.id)
      this.drawPlayer(pb, Phaser.Math.Linear(pa.x, pb.x, k), Phaser.Math.Linear(pa.y, pb.y, k))
    }
    for (const [id, av] of this.avatars) {
      if (!seen.has(id)) {
        for (const o of [av.sprite, av.label, av.ping, av.revive]) o.destroy()
        this.avatars.delete(id)
      }
    }
  }

  private drawPlayer(p: PlayerState, x: number, y: number) {
    const info = world.lobby.find(l => l.id === p.id)
    const slot = info?.slot ?? 0
    let av = this.avatars.get(p.id)
    if (av?.slot !== slot) { // new player, or the lobby info arrived after the first snapshot
      if (av) for (const o of [av.sprite, av.label, av.ping, av.revive]) o.destroy()
      av = {
        slot,
        sprite: this.add.sprite(x, y, `sprites/player/${SLOT_COLORS[slot]}/idle`).setOrigin(0.5, 1).setDepth(2),
        label: text(this, x, y, 14, SLOT_HEX[slot]).setOrigin(0.5, 1).setDepth(3),
        ping: text(this, x, y, 16, CREAM).setOrigin(0.5, 1).setDepth(3),
        revive: this.add.graphics().setDepth(3),
      }
      this.avatars.set(p.id, av)
    }
    const last = av.last ?? p
    av.last = p
    const color = SLOT_COLORS[slot]
    const anim = (name: string) => `sprites/player/${color}/${name}`
    const s = av.sprite

    // One-shots restart on a counter change; otherwise keep whatever is playing.
    const swung = p.attacks !== last.attacks || p.pulses !== last.pulses
    if (swung) {
      s.play(anim('attack'))
      if (p.attacks !== last.attacks) this.vfx('sprites/vfx/swing-arc', x + p.facing * 22, y - 22, p.facing < 0)
    } else if (p.dashes !== last.dashes) s.play(anim('dash'))
    else if (p.hurts !== last.hurts) {
      s.play(anim('hurt'))
      s.setTintFill(0xff5a5a)
      this.time.delayedCall(100, () => s.clearTint())
    } else if (p.action === 'revive' && last.action !== 'revive') s.play(anim('revive'))
    else if (p.action === 'down') s.play(anim('down'), true)
    else if (p.action === '') {
      s.play(anim(!p.ground ? (p.vy < 0 ? 'jump' : 'fall') : p.vx !== 0 ? 'run' : 'idle'), true)
    }
    if (p.ground && !last.ground && p.action === '') this.vfx('sprites/vfx/land-dust', x, y, false, 0.5, 1)

    s.setPosition(x, y).setFlipX(p.facing < 0).setAlpha(p.away ? 0.4 : 1)
    const me = p.id === world.myId ? '▼ ' : ''
    const top = Math.round(y) - 52
    av.label.setText(`${me}P${slot + 1} ${info?.name ?? ''}`).setPosition(Math.round(x), top)
    av.ping.setText(PINGS[p.ping] ? `「${PINGS[p.ping]}」` : '').setPosition(Math.round(x), top - 18)

    av.revive.clear()
    if (p.action === 'down') {
      av.revive.fillStyle(0x0b1020).fillRect(x - 20, y - 60, 40, 6)
      av.revive.fillStyle(0x8ef0b5).fillRect(x - 19, y - 59, 38 * p.revive, 4)
      av.label.setText(`${av.label.text} · Nghỉ phép`)
    }
  }

  private drawBoss(b: Snapshot, time: number) {
    const boss = b.boss
    this.boss.setPosition(boss.x, boss.y).setFlipX(boss.facing < 0) // art faces right
    this.boss.play(BOSS + boss.state, true)
    if (boss.hits !== this.bossHits) {
      if (this.bossHits >= 0 && boss.hits > this.bossHits) {
        this.boss.setTintFill(0xffffff)
        this.time.delayedCall(60, () => this.boss.clearTint())
        this.vfx('sprites/vfx/hit-spark', boss.x + Phaser.Math.Between(-24, 24), boss.y - Phaser.Math.Between(40, 90))
      }
      this.bossHits = boss.hits
    }
    this.warning.setVisible(boss.state === 'wave_windup').setFrame(Math.floor(time / 166) % 2)
  }

  private drawHazards(a: Snapshot, b: Snapshot, k: number) {
    sync(this.shots, b.shots, s => {
      const key = s.kind === 'card' ? 'sprites/vfx/task-card' : 'sprites/vfx/code-pulse'
      return this.add.sprite(s.x, s.y, key).play(key).setFlipX(s.vx < 0).setDepth(4)
    }, (sprite, s) => {
      const sa = a.shots.find(o => o.id === s.id) ?? s
      sprite.setPosition(Phaser.Math.Linear(sa.x, s.x, k), s.y)
    })
    sync(this.waves, b.waves, w => {
      const key = 'sprites/vfx/deadline-wave'
      return this.add.sprite(w.x, this.floorY, key).setOrigin(0.5, 1).play(key).setFlipX(w.dir < 0).setDepth(4)
    }, (sprite, w) => {
      const wa = a.waves.find(o => o.id === w.id) ?? w
      sprite.setPosition(Phaser.Math.Linear(wa.x, w.x, k), this.floorY)
    })
  }

  private drawHud(b: Snapshot) {
    const { deadline, deadlineText, bossBar, result } = this.hud
    const left = Math.max(0, b.deadlineMs) / Math.max(1, b.deadlineOf)
    const w = 600, x = VIEW_W / 2 - w / 2
    deadline.clear().fillStyle(0x0b1020).fillRect(x - 2, 20, w + 4, 20)
    deadline.fillStyle(left < 0.2 ? 0xff5a5a : 0x3fe0ff).fillRect(x, 22, w * left, 16)
    const secs = Math.ceil(Math.max(0, b.deadlineMs) / 1000)
    deadlineText.setText(`DEADLINE ${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`).setY(44)

    const boss = b.boss
    const bw = 800, bx = VIEW_W / 2 - bw / 2
    bossBar.clear().fillStyle(0x0b1020).fillRect(bx - 2, VIEW_H - 58, bw + 4, 18)
    bossBar.fillStyle(boss.state === 'wind_clock' ? 0xffd25e : 0xff5a5a).fillRect(bx, VIEW_H - 56, bw * (boss.hp / Math.max(1, boss.maxHp)), 14)

    this.drawChips(b)

    if (b.result === 'won') result.setColor(CYAN).setText('APPROVED!\nCả team được tan làm.')
    else if (b.result === 'lost') {
      result.setColor(CORAL).setText(b.deadlineMs <= 0 ? 'TRỄ DEADLINE\nLàm lại từ đầu…' : 'CẢ TEAM NGHỈ PHÉP\nLàm lại từ đầu…')
    } else result.setText('')
  }

  // Status chips, top left: portrait, P# name, hearts.
  private drawChips(b: Snapshot) {
    b.players.forEach((p, i) => {
      const slot = world.lobby.find(l => l.id === p.id)?.slot ?? 0
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

  private vfx(key: string, x: number, y: number, flip = false, ox = 0.5, oy = 0.5) {
    const s = this.add.sprite(x, y, key).setOrigin(ox, oy).setFlipX(flip).setDepth(5).play(key)
    s.once(Phaser.Animations.Events.ANIMATION_COMPLETE, () => s.destroy())
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
