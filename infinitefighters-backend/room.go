package main

import (
	"context"
	"crypto/rand"
	"log"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	maxPlayers    = 5
	tickRate      = 60
	snapshotEvery = 2 // ticks: 60 Hz simulation, 30 Hz snapshots
	maxNameRunes  = 16
	levelURL      = "/levels/sprint.json"

	phaseLobby   = "lobby"
	phasePlaying = "playing"
)

type Player struct {
	ID      int
	Name    string
	Token   string
	Slot    int
	client  *client // nil while disconnected
	in      Input   // held buttons
	pressed presses // press edges latched until the next tick consumes them
	Body
	fighter
	counts
}

type presses struct {
	jump, attack, skill, special, dash bool
	ping                               int
}

// Room is the only game. All its state is touched only by the run goroutine;
// other goroutines send closures through inbox.
type Room struct {
	code    string
	level   *Level
	inbox   chan func()
	phase   string
	players []*Player // join order
	nextID  int
	hostID  int
	tick    uint64
	match
}

func newRoom(code string, level *Level) *Room {
	r := &Room{code: code, level: level, inbox: make(chan func(), 256), phase: phaseLobby, nextID: 1}
	r.deadlineTicks = 10 * 60 * tickRate
	r.bossBaseHP = bossBaseHP
	return r
}

func (r *Room) run(ctx context.Context) {
	t := time.NewTicker(time.Second / tickRate)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case f := <-r.inbox:
			f()
		case <-t.C:
			r.update()
		}
	}
}

func (r *Room) update() {
	// shortcut: if everyone stays away the match pauses forever; add a timeout back to the lobby if that happens in practice.
	if r.phase != phasePlaying || !r.anyConnected() {
		return
	}
	r.stepMatch()
	if r.phase != phasePlaying {
		return // the win sent everyone back to the lobby
	}
	r.tick++
	if r.tick%snapshotEvery == 0 {
		r.broadcast(encode(r.snapshot()))
	}
}

func (r *Room) handle(c *client, m clientMsg) {
	if c.closing {
		return
	}
	switch m.Type {
	case "join":
		if c.player == nil {
			r.join(c, m)
		}
	case "start":
		if p := c.player; p != nil && p.ID == r.hostID && r.phase == phaseLobby {
			r.start()
		}
	case "input":
		if p := c.player; p != nil {
			pr := &p.pressed
			pr.jump = pr.jump || (m.Jump && !p.in.Jump)
			pr.attack = pr.attack || (m.Attack && !p.in.Attack)
			pr.skill = pr.skill || (m.Skill && !p.in.Skill)
			pr.dash = pr.dash || (m.Dash && !p.in.Dash)
			pr.special = pr.special || (m.Special && !p.in.Special)
			if m.Ping != 0 && m.Ping != p.in.Ping && m.Ping <= 3 {
				pr.ping = m.Ping
			}
			p.in = m.Input
		}
	}
}

func (r *Room) join(c *client, m clientMsg) {
	if m.Token != "" {
		if p := r.byToken(m.Token); p != nil {
			if p.client != nil { // same player opened a new tab: the newest socket wins
				p.client.player = nil
				p.client.reject("opened in another tab")
			}
			r.attach(p, c)
			return
		}
	}
	switch {
	case strings.ToUpper(strings.TrimSpace(m.Code)) != r.code:
		c.reject("wrong room code")
	case r.phase != phaseLobby:
		c.reject("match in progress")
	case len(r.players) >= maxPlayers:
		c.reject("room full")
	default:
		p := &Player{ID: r.nextID, Name: cleanName(m.Name), Token: rand.Text(), Slot: r.freeSlot()}
		r.nextID++
		r.players = append(r.players, p)
		if r.hostID == 0 {
			r.hostID = p.ID
		}
		log.Printf("join: P%d %q", p.Slot+1, p.Name)
		r.attach(p, c)
	}
}

func (r *Room) attach(p *Player, c *client) {
	p.client, c.player = c, p
	p.in, p.pressed = Input{}, presses{}
	c.push(encode(welcomeMsg{Type: "welcome", ID: p.ID, Token: p.Token, Phase: r.phase}))
	if r.phase == phasePlaying {
		c.push(encode(startMsg{Type: "start", Level: levelURL}))
	}
	r.broadcast(encode(r.lobby()))
}

func (r *Room) disconnect(c *client) {
	p := c.player
	if p == nil || p.client != c {
		return
	}
	p.client, c.player = nil, nil
	p.in, p.pressed = Input{}, presses{}
	if r.phase == phaseLobby { // nothing to keep yet: free the slot
		for i, q := range r.players {
			if q == p {
				r.players = append(r.players[:i], r.players[i+1:]...)
				break
			}
		}
	}
	if p.ID == r.hostID {
		r.hostID = 0
		for _, q := range r.players {
			if q.client != nil {
				r.hostID = q.ID
				break
			}
		}
	}
	log.Printf("left: P%d %q", p.Slot+1, p.Name)
	r.broadcast(encode(r.lobby()))
}

func (r *Room) start() {
	r.phase, r.tick = phasePlaying, 0
	r.scaleN = len(r.players) // fixed for the whole match, restarts included
	r.resetMatch()
	log.Printf("match started with %d players", len(r.players))
	r.broadcast(encode(startMsg{Type: "start", Level: levelURL}))
}

// toLobby ends the match after a win. Away players lose their slot.
func (r *Room) toLobby() {
	r.phase = phaseLobby
	r.players = keep(r.players, func(p *Player) bool { return p.client != nil })
	if r.byID(r.hostID) == nil {
		r.hostID = 0
		if len(r.players) > 0 {
			r.hostID = r.players[0].ID
		}
	}
	log.Printf("match won, back to lobby")
	r.broadcast(encode(r.lobby()))
}

func (r *Room) broadcast(b []byte) {
	for _, p := range r.players {
		if p.client != nil {
			p.client.push(b)
		}
	}
}

func (r *Room) lobby() lobbyMsg {
	m := lobbyMsg{Type: "lobby", Phase: r.phase, Players: []lobbyPlayer{}}
	for _, p := range r.players {
		m.Players = append(m.Players, lobbyPlayer{ID: p.ID, Name: p.Name, Slot: p.Slot, Host: p.ID == r.hostID, Connected: p.client != nil})
	}
	return m
}

func (r *Room) snapshot() snapshotMsg {
	b := r.boss
	ms := func(ticks int) int64 { return int64(ticks) * 1000 / tickRate }
	m := snapshotMsg{
		Type: "snapshot", Tick: r.tick, ServerTime: time.Now().UnixMilli(),
		DeadlineMs: ms(r.deadline), DeadlineOf: ms(r.deadlineTicks), Result: r.result,
		Boss: bossState{
			X: b.X, Y: b.Y, Facing: b.Facing, HP: b.HP, MaxHP: b.MaxHP, State: b.State,
			Hits: b.Hits, Marks: b.Marks, MarkMs: ms(b.markTicks),
		},
		Players: []playerState{}, Shots: []shotState{}, Waves: []waveState{}, Items: []itemState{}, Zones: []zoneState{}, Drops: []dropState{},
	}
	if b.State == bossChargeWindup || b.State == bossCharge {
		m.Boss.ChargeFrom, m.Boss.ChargeTo = b.chargeFrom, b.chargeTo
	}
	for _, p := range r.players {
		m.Players = append(m.Players, playerState{
			ID: p.ID, X: p.X, Y: p.Y, VX: p.VX, VY: p.VY,
			Facing: p.Facing, Ground: p.OnGround, Away: p.client == nil,
			HP: p.HP, Action: p.action, Revive: float64(p.reviveHeld) / reviveHoldTicks, Ping: p.ping,
			SkillMs: ms(p.skillCD), SkillOf: ms(skillCooldown[p.Slot]), Charge: float64(p.charge) / chargeMaxTicks,
			Attacks: p.Attacks, Pulses: p.Pulses, Dashes: p.Dashes, Hurts: p.Hurts, Hits: p.Hits,
			Casts: p.Casts, SkillHits: p.SkillHits, Cancels: p.Cancels,
			Reflects: p.Reflects, Healed: p.Healed, Coffees: p.Coffees,
		})
	}
	for _, s := range append(append([]*Shot{}, r.cards...), r.shots...) {
		m.Shots = append(m.Shots, shotState{ID: s.ID, Kind: s.Kind, X: s.X, Y: s.Y, VX: s.VX})
	}
	for _, w := range r.waves {
		m.Waves = append(m.Waves, waveState{ID: w.ID, X: w.X, Dir: w.Dir})
	}
	for _, d := range r.drops {
		wait := d.delay
		if b.State == bossDropWindup {
			wait += dropWindupTicks - b.ticks
		}
		m.Drops = append(m.Drops, dropState{ID: d.ID, X: d.X, Top: d.Top, Bottom: d.Bottom, Y: d.Y, Ms: ms(wait)})
	}
	for _, z := range r.zones {
		m.Zones = append(m.Zones, zoneState{ID: z.ID, X: z.X, Y: z.Y, Ms: ms(z.ticks)})
	}
	for i, c := range r.level.Coffees {
		if r.coffeeIn[i] == 0 {
			m.Items = append(m.Items, itemState{ID: i, Kind: "coffee", X: c.X, Y: c.Y})
		}
	}
	return m
}

func (r *Room) anyConnected() bool {
	for _, p := range r.players {
		if p.client != nil {
			return true
		}
	}
	return false
}

func (r *Room) byID(id int) *Player {
	for _, p := range r.players {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (r *Room) byToken(t string) *Player {
	for _, p := range r.players {
		if p.Token == t {
			return p
		}
	}
	return nil
}

func (r *Room) freeSlot() int {
	used := [maxPlayers]bool{}
	for _, p := range r.players {
		used[p.Slot] = true
	}
	for i, u := range used {
		if !u {
			return i
		}
	}
	return 0 // unreachable: join checks the player count first
}

func cleanName(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "Dev"
	}
	if utf8.RuneCountInString(s) > maxNameRunes {
		s = string([]rune(s)[:maxNameRunes])
	}
	return s
}

// newRoomCode avoids 0/O and 1/I. 32 symbols, so byte%32 is unbiased.
func newRoomCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	rand.Read(b)
	for i := range b {
		b[i] = alphabet[b[i]%32]
	}
	return string(b)
}
