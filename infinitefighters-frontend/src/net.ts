// Mirrors infinitefighters-backend/protocol.go.

export type Input = {
  left: boolean; right: boolean; down: boolean; jump: boolean
  attack: boolean; skill: boolean; special: boolean; dash: boolean; interact: boolean
  ping: number // 0 none, 1 Tôi nhận, 2 Cần hỗ trợ, 3 Chặn scope
}

// The sandbox bosses (coop scene, no DEADLINE): see docs/levels/sandbox-*.md.
export type SandboxBoss = 'khabib' | 'vought' | 'yasuka'

export type LobbyPlayer = { id: number; name: string; slot: number; host: boolean; connected: boolean }

// x/y are the feet (bottom-centre of the hitbox) in level pixels, +y down.
export type PlayerState = {
  id: number; x: number; y: number; vx: number; vy: number
  facing: number; ground: boolean; away: boolean
  hp: number
  action: '' | 'attack' | 'pulse' | 'dash' | 'hurt' | 'down' | 'revive'
    | 'charge' | 'smash' | 'shield' | 'stamp' | 'hotfix' | 'timecard' | 'pat'
  revive: number // 0..1
  ping: number
  skillMs: number; skillOf: number // slot skill cooldown left / full
  charge: number // P1: 0..1 while charging
  // Counters only grow: a change means "play the one-shot now".
  attacks: number; pulses: number; dashes: number; hurts: number; hits: number
  pvp?: PvPState // party only
  casts: number; skillHits: number; cancels: number; reflects: number; healed: number; coffees: number; rejects: number
}

// Party PvP phase (mirrors pvpMsg / partyMsg in protocol.go).
export type PvPState = {
  hp: number; wallet: number; balance: number; kills: number; dead: boolean; left: boolean
  respawnMs: number; protectMs: number; markedBy: number
  pickups: number; drops: number; deposits: number; atmStarts: number; atmCancels: number; respawns: number; marks: number
}
export type PartyState = {
  countdownMs: number; timeLeftMs: number; solo: boolean; safeX: number; safeY: number; sprays: number
  banner?: string
  bundles: { id: number; x: number; y: number; value: number }[]
  atms: { x: number; y: number; user: number; ms: number; of: number; amount: number }[]
  results?: { id: number; balance: number; rank: number; kills: number }[]
}

export type BossState = {
  x: number; y: number; facing: number; hp: number; maxHp: number; state: string
  hits: number; marks: number; markMs: number // counters reset each attempt: a decrease is a reset
  chargeFrom?: number; chargeTo?: number // locked charge corridor
  shield?: number // Yasuka's silence dome strength, 1 down to 0, while it is up
  landX?: number; landY?: number // Khabib teleport: where he lands (feet), set while he is out of sight
}
export type ShotState = { id: number; kind: 'card' | 'pulse' | 'stamp' | 'returned' | 'paper'; x: number; y: number; vx: number; word?: string }
// Vought's DEV stamp over a player: break it with hits before ms runs out.
export type StampState = { id: number; x: number; y: number; hp: number; ms: number }
// Yasuka's rejection seal: asleep while ms > 0, then it sends a player back to their start.
export type SealState = { id: number; x: number; y: number; ms: number }
export type ZoneState = { id: number; x: number; y: number; ms: number }
// A locked column of the drop attack: warn from top to bottom until the stack falls (ms = time left).
export type DropState = { id: number; x: number; top: number; bottom: number; y: number; ms: number }
export type WaveState = { id: number; x: number; y: number; dir: number }
export type ItemState = { id: number; kind: 'coffee'; x: number; y: number }

export type Snapshot = {
  type: 'snapshot'; tick: number; serverTime: number
  deadlineMs: number; deadlineOf: number
  result: '' | 'won' | 'lost'
  players: PlayerState[]; boss: BossState; shots: ShotState[]; waves: WaveState[]; items: ItemState[]; zones: ZoneState[]; drops: DropState[]; stamps?: StampState[]; seals?: SealState[]
  mode: 'coop' | 'party'; party?: PartyState
}

export type ServerMsg =
  | { type: 'welcome'; id: number; token: string; phase: string }
  | { type: 'lobby'; phase: 'lobby' | 'playing'; players: LobbyPlayer[] }
  | { type: 'start'; level: string; mode: 'coop' | 'party'; boss?: SandboxBoss }
  | Snapshot

export type Join = { name?: string; code?: string; token?: string }

// Shared between the lobby UI and the game scene.
export const world = {
  myId: 0,
  lobby: [] as LobbyPlayer[],
  snapshots: [] as Snapshot[],
  clockOffset: Infinity, // min(local receive time - serverTime): maps server time to performance.now()
}

let ws: WebSocket | undefined

export function connect(join: Join, onMessage: (m: ServerMsg) => void, onClose: (reason: string, rejected: boolean) => void) {
  const sock = new WebSocket(`${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws`)
  sock.onopen = () => sock.send(JSON.stringify({ type: 'join', ...join }))
  sock.onmessage = e => {
    const m = JSON.parse(e.data) as ServerMsg
    if (m.type === 'snapshot') {
      world.clockOffset = Math.min(world.clockOffset, performance.now() - m.serverTime)
      world.snapshots.push(m)
    }
    onMessage(m)
  }
  sock.onclose = e => onClose(e.reason, e.code === 1008) // 1008: server rejected the join
  ws = sock
}

export function send(m: object) {
  if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify(m))
}

// Browser tests read the live state from the console with ?debug in the URL.
if (new URLSearchParams(location.search).has('debug')) Object.assign(window, { __world: world })
