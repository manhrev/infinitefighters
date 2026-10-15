import Phaser from 'phaser'
import type { SceneAudio } from './audio'
import { world, type PlayerState, type Snapshot } from './net'

// The party PvP phase ("Anh em hết mình, tiền tình hết hồn"): safe, ATMs, coins, timer, rules,
// results and their sounds. Art and sounds that are not generated yet fall back to placeholders.

const CREAM = '#F2E3C2', GOLD = '#FFD25E', CORAL = '#FF5A5A', CYAN = '#3FE0FF', OUTLINE = '#0B1020'
const SLOT_HEX = ['#4FD1C5', '#B08CF0', '#FFD25E', '#8EF0B5', '#FFA25E']
const RULES = 'Nhặt tiền → đánh rơi tiền đối thủ → gửi ATM'

const label = (s: Phaser.Scene, x: number, y: number, size: number, color: string) =>
  s.add.text(x, y, '', { fontFamily: 'monospace', fontSize: `${size}px`, fontStyle: 'bold', color, stroke: OUTLINE, strokeThickness: 4 })

type Obj = Phaser.Types.Tilemaps.TiledObject
type ATMView = { bar: Phaser.GameObjects.Graphics; amount: Phaser.GameObjects.Text; screen?: Phaser.GameObjects.Sprite; x: number; y: number }

export class PartyView {
  private safe?: Phaser.GameObjects.Sprite
  private atms: ATMView[] = []
  private bundles = new Map<number, Phaser.GameObjects.GameObject & { setPosition(x: number, y: number): unknown }>()
  private center: Phaser.GameObjects.Text
  private banner: Phaser.GameObjects.Text

  constructor(private scene: Phaser.Scene, private sfx: SceneAudio, objects: Obj[]) {
    const s = scene
    const safe = objects.find(o => o.type === 'safe')
    if (safe) {
      if (s.textures.exists('sprites/party/safe')) {
        this.safe = s.add.sprite(safe.x!, safe.y!, 'sprites/party/safe', 0).setOrigin(0.5, 1).setDepth(1)
      } else {
        s.add.rectangle(safe.x!, safe.y!, 56, 52, 0x2f4a7f).setOrigin(0.5, 1).setStrokeStyle(3, 0xffd25e).setDepth(1)
        label(s, safe.x!, safe.y! - 18, 12, GOLD).setText('KÉT').setOrigin(0.5, 1).setDepth(1)
      }
    }
    for (const o of objects.filter(o => o.type === 'atm')) {
      let screen: Phaser.GameObjects.Sprite | undefined
      if (s.textures.exists('sprites/party/atm')) {
        screen = s.add.sprite(o.x!, o.y!, 'sprites/party/atm', 0).setOrigin(0.5, 1).setDepth(1)
      } else {
        s.add.rectangle(o.x!, o.y!, 40, 60, 0x22345f).setOrigin(0.5, 1).setStrokeStyle(2, 0x3fe0ff).setDepth(1)
        label(s, o.x!, o.y! - 36, 11, CYAN).setText('ATM').setOrigin(0.5, 1).setDepth(1)
      }
      this.atms.push({
        x: o.x!, y: o.y!, screen,
        bar: s.add.graphics().setDepth(6),
        amount: label(s, o.x!, o.y! - 84, 12, GOLD).setOrigin(0.5, 1).setDepth(6),
      })
    }
    this.center = label(s, 640, 300, 28, CREAM).setOrigin(0.5).setAlign('center').setScrollFactor(0).setDepth(20)
    this.banner = label(s, 640, 110, 22, GOLD).setOrigin(0.5).setScrollFactor(0).setDepth(20)
  }

  update(a: Snapshot, b: Snapshot, k: number) {
    const party = b.party
    if (!party) return
    // coins: interpolate while they fly
    const live = new Set(party.bundles.map(c => c.id))
    for (const c of party.bundles) {
      let v = this.bundles.get(c.id)
      if (!v) {
        if (this.scene.textures.exists('sprites/party/coin-bundle')) {
          const sp = this.scene.add.sprite(c.x, c.y, 'sprites/party/coin-bundle').setOrigin(0.5, 1).setDepth(3)
          if (this.scene.anims.exists('sprites/party/coin-bundle')) sp.play('sprites/party/coin-bundle')
          v = sp
        } else {
          v = this.scene.add.circle(c.x, c.y - 6, 6, 0xffd25e).setStrokeStyle(2, 0x8a5a00).setDepth(3)
        }
        this.bundles.set(c.id, v)
      }
      const ca = a.party?.bundles.find(o => o.id === c.id) ?? c
      const yOff = v instanceof Phaser.GameObjects.Sprite ? 0 : -6
      v.setPosition(Phaser.Math.Linear(ca.x, c.x, k), Phaser.Math.Linear(ca.y, c.y, k) + yOff)
    }
    for (const [id, v] of this.bundles) {
      if (!live.has(id)) { v.destroy(); this.bundles.delete(id) }
    }
    // ATMs: progress and the amount being deposited, readable by everyone
    party.atms.forEach((m, i) => {
      const v = this.atms[i]
      if (!v) return
      v.bar.clear()
      v.screen?.setFrame(m.user ? 1 : 0)
      if (m.user) {
        v.bar.fillStyle(0x0b1020).fillRect(v.x - 24, v.y - 80, 48, 8)
        v.bar.fillStyle(0x3fe0ff).fillRect(v.x - 23, v.y - 79, 46 * (m.ms / Math.max(1, m.of)), 6)
        v.amount.setText(`${m.amount} xu`)
      } else v.amount.setText('')
    })

    // centre text: countdown with the rules, then the results
    if (party.countdownMs > 0) {
      const n = Math.ceil(party.countdownMs / 1000)
      const head = 'ANH EM CÂY KHẾ\n' + (party.solo ? 'Gom thưởng trước giờ khóa sổ' : '“Anh em là chính!” — đồng đội → đối tượng cạnh tranh')
      this.center.setColor(CREAM).setText(`${head}\n${RULES}\nBắt đầu sau ${n}`)
    } else if (party.results) {
      this.center.setColor(GOLD).setText(this.resultsText(party.results))
    } else this.center.setText('')
    this.banner.setText(party.banner ?? '')
  }

  private resultsText(results: NonNullable<Snapshot['party']>['results'] & object) {
    const name = (id: number) => {
      const l = world.lobby.find(p => p.id === id)
      return l ? `P${l.slot + 1} ${l.name}` : `#${id}`
    }
    const lines = results.map(r => `${r.rank}. ${name(r.id)} — ${r.balance} xu${r.kills ? ` (${r.kills} hạ gục)` : ''}`)
    const top = results.filter(r => r.rank === 1 && r.balance > 0)
    const who = top.length ? `${top.map(r => name(r.id)).join(', ')} ôm vali!` : 'Không ai gửi được đồng nào.'
    return `AI ÔM TIỀN\n${lines.join('\n')}\n${who}\n“Khoản thưởng đang chờ Approve.”`
  }

  // timer reads the time left for the DEADLINE-style bar: [fraction, label, flash].
  timer(b: Snapshot, time: number): [number, string, boolean] {
    const p = b.party!
    const left = p.countdownMs > 0 ? 180_000 : p.timeLeftMs
    const secs = Math.ceil(left / 1000)
    const flash = left > 0 && left <= 10_000 && Math.floor(time / 250) % 2 === 0
    return [left / 180_000, `KHÓA SỔ ${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`, flash]
  }

  // chipLine is the money line under each player's name: HP, wallet, deposited.
  chipLine(p: PlayerState) {
    const v = p.pvp!
    if (v.left) return 'đã rời sân'
    return `♥${Math.ceil(v.hp)} VÍ${v.wallet} GỬI${v.balance}`
  }

  events(was: Snapshot, now: Snapshot) {
    const s = this.sfx, p0 = was.party, p1 = now.party
    if (!p0 || !p1) return
    const play = (key: string, fallback: string, o?: Parameters<SceneAudio['play']>[1]) => s.play(s.has(key) ? key : fallback, o)
    if (p1.countdownMs > 0 && Math.ceil(p1.countdownMs / 1000) !== Math.ceil(p0.countdownMs / 1000)) s.play('deadline-tick', { critical: true })
    if (p0.countdownMs > 0 && p1.countdownMs === 0) play('party-start', 'checkpoint', { critical: true })
    if (p1.sprays > p0.sprays) {
      play('safe-burst', 'coin-drop', { x: p1.safeX })
      this.safe?.play('sprites/party/safe')
    }
    if (p1.banner && !p0.banner) play('hr-announce', 'checkpoint', { critical: true })
    const sec = Math.ceil(p1.timeLeftMs / 1000)
    if (p1.countdownMs === 0 && p1.timeLeftMs > 0 && p1.timeLeftMs <= 10_000 && sec !== Math.ceil(p0.timeLeftMs / 1000)) {
      s.play('deadline-tick', { critical: true })
    }
    if (p1.results && !p0.results) s.play('victory', { critical: true })
    for (const p of now.players) {
      const q = was.players.find(o => o.id === p.id)
      if (!p.pvp || !q?.pvp) continue
      const o = { x: p.x, mine: p.id === world.myId }
      const up = (k: keyof NonNullable<PlayerState['pvp']>) => (p.pvp![k] as number) > (q.pvp![k] as number)
      if (up('pickups')) play('coin-pickup', 'coffee', { ...o, volume: 0.6 })
      if (up('drops')) {
        play('coin-drop', 'land', o)
        this.vfx('sprites/vfx/coin-burst', p.x, p.y)
      }
      if (up('atmStarts')) play('atm-start', 'ui-confirm', o)
      if (up('atmCancels')) play('atm-cancel', 'ui-back', o)
      if (up('deposits')) {
        play('atm-success', 'checkpoint', o)
        this.vfx('sprites/vfx/atm-success', p.x, p.y - 70)
      }
      if (up('respawns')) play('respawn', 'revive', o)
      if (up('marks')) s.play('amber-mark-hit', o)
    }
  }

  private vfx(key: string, x: number, y: number) {
    if (!this.scene.anims.exists(key)) return
    const sp = this.scene.add.sprite(x, y, key).setOrigin(0.5, 1).setDepth(5).play(key)
    sp.once(Phaser.Animations.Events.ANIMATION_COMPLETE, () => sp.destroy())
  }

  static hpColor(hp: number) { return hp > 50 ? 0x8ef0b5 : hp > 25 ? 0xffd25e : 0xff5a5a }
  static slotHex(slot: number) { return SLOT_HEX[slot] }
  static coral = CORAL
}
