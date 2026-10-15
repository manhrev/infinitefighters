import Phaser from 'phaser'
import { lobbyMusic, saveSettings, settings } from './audio'
import { GameScene, VIEW_H, VIEW_W } from './game'
import { connect, send, world, type Join, type ServerMsg } from './net'

const $ = <T extends HTMLElement>(id: string) => document.getElementById(id) as T
const lobbyEl = $('lobby'), joinForm = $<HTMLFormElement>('join'), roomEl = $('room')
const playersEl = $('players'), levelsEl = $('levels'), waitEl = $('wait'), statusEl = $('status')

// Remembered so a reload or a dropped socket resumes the same character.
type Session = { name: string; code: string; token: string }
const STORE = 'taskvotan.session'
let session: Session = { name: '', code: '', token: '' }
try { session = { ...session, ...JSON.parse(localStorage.getItem(STORE) ?? '{}') } } catch { /* storage blocked */ }
function save() {
  try { localStorage.setItem(STORE, JSON.stringify(session)) } catch { /* storage blocked */ }
}

let game: Phaser.Game | undefined

function onMessage(m: ServerMsg) {
  switch (m.type) {
    case 'welcome':
      world.myId = m.id
      session.token = m.token
      save()
      statusEl.textContent = ''
      joinForm.hidden = true
      roomEl.hidden = false
      if (m.phase === 'lobby') lobbyMusic(true)
      break
    case 'lobby': {
      world.lobby = m.players
      const me = m.players.find(p => p.id === world.myId)
      playersEl.replaceChildren(...m.players.map(p => {
        const li = document.createElement('li')
        li.textContent = `P${p.slot + 1} ${p.name}${p.host ? ' (chủ phòng)' : ''}${p.id === world.myId ? ' — bạn' : ''}`
        li.classList.toggle('away', !p.connected)
        return li
      }))
      levelsEl.hidden = !me?.host
      waitEl.hidden = !!me?.host
      if (m.phase === 'lobby' && game) { // match won: back to the lobby
        game.destroy(true)
        game = undefined
        gameLevel = ''
        world.snapshots = []
        lobbyEl.hidden = false
        lobbyMusic(true)
      }
      break
    }
    case 'start':
      lobbyEl.hidden = true
      lobbyMusic(false)
      startGame(m.level, m.mode)
      break
  }
}

function onClose(reason: string, rejected: boolean) {
  if (reason === 'opened in another tab') {
    statusEl.textContent = 'Đang chơi ở tab khác.'
    return
  }
  if (rejected) {
    session.token = ''
    save()
    statusEl.textContent = reason
    if (!game) showJoin()
    return
  }
  statusEl.textContent = 'Mất kết nối, đang nối lại…'
  setTimeout(() => join(session), 1000)
}

function join(j: Join) {
  connect(j, onMessage, onClose)
}

function showJoin() {
  lobbyEl.hidden = false
  joinForm.hidden = false
  roomEl.hidden = true
  ;(joinForm.elements.namedItem('name') as HTMLInputElement).value = session.name
  ;(joinForm.elements.namedItem('code') as HTMLInputElement).value = session.code
}

joinForm.addEventListener('submit', e => {
  e.preventDefault()
  const data = new FormData(joinForm)
  session = { name: String(data.get('name')).trim(), code: String(data.get('code')).trim().toUpperCase(), token: '' }
  save()
  join(session)
})
levelsEl.querySelectorAll<HTMLButtonElement>('.level').forEach(b =>
  b.addEventListener('click', () => send({ type: 'start', level: b.dataset.level })))

let gameLevel = ''

function startGame(level: string, mode: 'coop' | 'party') {
  if (game && level === gameLevel) return // reconnect mid-match: the scene is already running
  if (game) { // boss beaten, on to the party: a fresh scene, no old snapshots
    game.destroy(true)
    world.snapshots = []
  }
  gameLevel = level
  game = new Phaser.Game({
    type: Phaser.AUTO,
    parent: 'game',
    width: VIEW_W,
    height: VIEW_H,
    backgroundColor: '#0B1530',
    pixelArt: true,
    // Fill the window, keep 16:9. pixelArt keeps nearest-neighbour, so a non-integer scale gives slightly uneven pixels, not blur.
    scale: { mode: Phaser.Scale.FIT, autoCenter: Phaser.Scale.CENTER_BOTH },
  })
  game.scene.add('game', GameScene, true, { level, mode })
}

// Audio settings panel: Master / Music / SFX / mute, remembered per browser.
for (const k of ['master', 'music', 'sfx'] as const) {
  const input = $<HTMLInputElement>(`vol-${k}`)
  input.value = String(Math.round(settings[k] * 100))
  input.addEventListener('input', () => { settings[k] = Number(input.value) / 100; saveSettings() })
}
const muteBox = $<HTMLInputElement>('vol-mute')
muteBox.checked = settings.muted
muteBox.addEventListener('change', () => { settings.muted = muteBox.checked; saveSettings() })
$('sound-toggle').addEventListener('click', () => { $('sound').hidden = !$('sound').hidden })
// Browsers block sound until the first interaction: retry the lobby music then.
addEventListener('pointerdown', () => { if (!game && !roomEl.hidden) lobbyMusic(true) }, { once: true })
addEventListener('keydown', () => { if (!game && !roomEl.hidden) lobbyMusic(true) }, { once: true })

if (session.token) join(session)
else showJoin()
