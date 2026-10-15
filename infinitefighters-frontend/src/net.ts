// Mirrors infinitefighters-backend/protocol.go.

export type Input = {
  left: boolean; right: boolean; down: boolean; jump: boolean
  attack: boolean; skill: boolean; dash: boolean; interact: boolean
  ping: number // 0 none, 1 Tôi nhận, 2 Cần hỗ trợ, 3 Chặn scope
}

export type LobbyPlayer = { id: number; name: string; slot: number; host: boolean; connected: boolean }

// x/y are the feet (bottom-centre of the hitbox) in level pixels, +y down.
export type PlayerState = {
  id: number; x: number; y: number; vx: number; vy: number
  facing: number; ground: boolean; away: boolean
  hp: number
  action: '' | 'attack' | 'pulse' | 'dash' | 'hurt' | 'down' | 'revive'
  revive: number // 0..1
  ping: number
  // Counters only grow: a change means "play the one-shot now".
  attacks: number; pulses: number; dashes: number; hurts: number
}

export type BossState = { x: number; y: number; facing: number; hp: number; maxHp: number; state: string; hits: number }
export type ShotState = { id: number; kind: 'card' | 'pulse'; x: number; y: number; vx: number }
export type WaveState = { id: number; x: number; dir: number }

export type Snapshot = {
  type: 'snapshot'; tick: number; serverTime: number
  deadlineMs: number; deadlineOf: number
  result: '' | 'won' | 'lost'
  players: PlayerState[]; boss: BossState; shots: ShotState[]; waves: WaveState[]
}

export type ServerMsg =
  | { type: 'welcome'; id: number; token: string; phase: string }
  | { type: 'lobby'; phase: 'lobby' | 'playing'; players: LobbyPlayer[] }
  | { type: 'start'; level: string }
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
