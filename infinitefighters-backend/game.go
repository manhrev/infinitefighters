package main

import "math"

// Combat, boss and match flow for the MVP level (dev/game-scenario.md).
//
// Rules this file guarantees:
//   - Boss HP and attack directions are computed once at match start from every slot in the
//     room, connected or not; a restart after a loss keeps them.
//   - Only one boss attack is ever live: the next windup waits until all hazards are gone.
//   - Every attack leaves a safe zone: volleys fly at platform heights (the floor is safe),
//     waves run along the floor (any platform is safe).
//   - Away (disconnected) and downed players are ignored by hazards and boss aiming,
//     and away players do not count for "everyone is down".

// Action lengths match the art (assets/sprites/manifest.json): frames × tickRate / fps.
const (
	attackTicks = 5 * tickRate / 18
	dashTicks   = 4 * tickRate / 16
	hurtTicks   = 2 * tickRate / 10
	reviveTicks = 4 * tickRate / 10
)

// Tuning.
const (
	playerHP        = 5
	reviveHP        = 3
	reviveHoldTicks = 2 * tickRate // a teammate holds interact this long next to a downed player
	reviveReach     = 48.0
	invulnTicks     = tickRate // after a hit
	knockSpeed      = 220.0
	knockHop        = 380.0
	dashSpeed       = 700.0
	dashCooldown    = 48
	swingReach      = 36.0 // in front of the hitbox edge
	swingFrom       = 4    // ticks into the swing when it starts to hit
	swingTo         = 12
	pulseCooldown   = 90
	pulseSpeed      = 600.0
	pulseLife       = 27 // ticks, ≈ 270 px
	pulseDamage     = 2
	pingTicks       = 2 * tickRate

	bossBaseHP      = 120
	bossHPPerPlayer = 60 // per player beyond the first
	bossW, bossH    = 70.0, 110.0
	bossIdleTicks   = tickRate
	windupTicks     = 54 // telegraph; well above the clients' ~100 ms render delay
	volleyTicks     = 30
	waveTicks       = 40
	windTicks       = 150 // vulnerable window: damage ×2
	extraAt         = 24  // 4+ players: a second salvo / wave this many ticks into the attack
	cardSpeed       = 480.0
	cardW, cardH    = 16.0, 12.0
	waveSpeed       = 420.0
	waveW, waveH    = 32.0, 32.0

	lostTicks = 4 * tickRate // show why the team lost, then restart the level
	wonTicks  = 6 * tickRate // show the win, then back to the lobby
)

const (
	actAttack = "attack"
	actPulse  = "pulse"
	actDash   = "dash"
	actHurt   = "hurt"
	actDown   = "down" // "Nghỉ phép"
	actRevive = "revive"

	bossIdle         = "idle"
	bossVolleyWindup = "volley_windup"
	bossVolley       = "volley"
	bossWaveWindup   = "wave_windup"
	bossWave         = "wave"
	bossWind         = "wind_clock"
	bossDefeat       = "defeat"

	resultWon  = "won"
	resultLost = "lost"
)

var bossCycle = []string{bossVolleyWindup, bossWaveWindup, bossVolleyWindup, bossWaveWindup, bossWind}

// fighter is a player's per-match combat state; a restart resets it whole.
type fighter struct {
	HP          int
	action      string
	actionTicks int
	swingHit    bool
	invuln      int
	dashCD      int
	pulseCD     int
	reviveHeld  int
	ping        int
	pingTicks   int
}

// counts only grow. Clients fire one-shot animations when a counter changes.
type counts struct {
	Attacks, Pulses, Dashes, Hurts uint32
}

type Boss struct {
	X, Y   float64 // feet
	Facing int
	HP     int
	MaxHP  int
	State  string
	ticks  int // in the current state
	cycle  int
	Hits   uint32
}

type Shot struct {
	ID       int
	X, Y, VX float64 // centre
	W, H     float64
	life     int // pulses only
}

type Wave struct {
	ID  int
	X   float64 // centre
	Dir int
}

type match struct {
	scaleN        int // players counted at match start
	deadlineTicks int // configured length
	deadline      int // ticks left
	result        string
	resultTicks   int
	boss          Boss
	cards         []*Shot // boss task cards
	pulses        []*Shot // players' code pulses
	waves         []*Wave
	nextShotID    int
}

func (r *Room) resetMatch() {
	hp := bossBaseHP + bossHPPerPlayer*(r.scaleN-1)
	b := r.level.Boss
	r.boss = Boss{X: b.X, Y: b.Y, Facing: -1, HP: hp, MaxHP: hp, State: bossIdle}
	r.cards, r.pulses, r.waves = nil, nil, nil
	r.deadline, r.result, r.resultTicks = r.deadlineTicks, "", 0
	for i, p := range r.players {
		s := r.level.Spawns[i%len(r.level.Spawns)]
		p.Body = Body{X: s.X, Y: s.Y, Facing: 1}
		p.fighter = fighter{HP: playerHP}
		p.pressed = presses{}
	}
}

func (r *Room) stepMatch() {
	if r.result != "" {
		if r.resultTicks--; r.resultTicks <= 0 {
			if r.result == resultLost {
				r.resetMatch()
			} else {
				r.toLobby()
			}
		}
		return
	}
	r.deadline--
	for _, p := range r.players {
		r.stepPlayer(p)
	}
	r.stepRevives()
	r.stepBoss()
	r.stepShots()
	r.stepWaves() // no contact damage: the boss hurts only through telegraphed attacks

	switch {
	case r.boss.HP <= 0:
		r.boss.HP, r.boss.State, r.boss.ticks = 0, bossDefeat, 0
		r.cards, r.pulses, r.waves = nil, nil, nil
		r.result, r.resultTicks = resultWon, wonTicks
	case r.deadline <= 0 || r.allDown():
		r.result, r.resultTicks = resultLost, lostTicks
	}
}

func (r *Room) stepPlayer(p *Player) {
	pr := p.pressed
	p.pressed = presses{}
	if p.client == nil {
		return // away: frozen
	}
	for _, t := range []*int{&p.invuln, &p.dashCD, &p.pulseCD} {
		if *t > 0 {
			*t--
		}
	}
	if p.pingTicks > 0 {
		p.pingTicks--
	} else {
		p.ping = 0
	}
	if pr.ping != 0 {
		p.ping, p.pingTicks = pr.ping, pingTicks
	}
	if p.actionTicks > 0 {
		if p.actionTicks--; p.actionTicks == 0 {
			p.action = ""
		}
	}

	in := p.in
	switch p.action {
	case actDown, actHurt, actRevive:
		p.step(r.level, Input{}, false)
		return
	}

	if p.action == "" {
		dir := 0
		if in.Left {
			dir--
		}
		if in.Right {
			dir++
		}
		switch {
		case pr.dash && p.dashCD == 0:
			if dir != 0 {
				p.Facing = dir
			}
			p.action, p.actionTicks, p.Body.dashTicks, p.dashCD = actDash, dashTicks, dashTicks, dashCooldown
			p.Dashes++
		case pr.attack:
			p.action, p.actionTicks, p.swingHit = actAttack, attackTicks, false
			p.Attacks++
		case pr.skill && p.pulseCD == 0:
			p.action, p.actionTicks, p.pulseCD = actPulse, attackTicks, pulseCooldown
			p.Pulses++
			r.nextShotID++
			r.pulses = append(r.pulses, &Shot{
				ID: r.nextShotID, X: p.X + float64(p.Facing)*(hitW/2+16), Y: p.Y - hitH/2,
				VX: float64(p.Facing) * pulseSpeed, W: 32, H: 16, life: pulseLife,
			})
		}
	}

	facing := p.Facing
	p.step(r.level, in, pr.jump && p.action != actDash)
	if p.action == actAttack || p.action == actPulse {
		p.Facing = facing // no turning mid-swing
	}
	if p.action == actAttack && !p.swingHit && r.boss.State != bossDefeat {
		elapsed := attackTicks - p.actionTicks
		if elapsed >= swingFrom && elapsed <= swingTo && overlaps(swingBox(p), r.bossBox()) {
			p.swingHit = true
			r.damageBoss(1)
		}
	}
}

func (r *Room) stepRevives() {
	for _, d := range r.players {
		if d.action != actDown || d.client == nil {
			continue
		}
		helped := false
		for _, h := range r.players {
			if h != d && h.client != nil && h.action != actDown && h.in.Interact &&
				math.Abs(h.X-d.X) <= reviveReach && math.Abs(h.Y-d.Y) <= reviveReach {
				helped = true
			}
		}
		if !helped {
			d.reviveHeld = 0
			continue
		}
		if d.reviveHeld++; d.reviveHeld >= reviveHoldTicks {
			d.HP, d.reviveHeld = reviveHP, 0
			d.action, d.actionTicks, d.invuln = actRevive, reviveTicks, invulnTicks
		}
	}
}

func (r *Room) stepBoss() {
	b := &r.boss
	b.ticks++
	set := func(s string) { b.State, b.ticks = s, 0 }
	switch b.State {
	case bossIdle:
		if t := r.nearestTarget(); t != nil {
			b.Facing = 1
			if t.X < b.X {
				b.Facing = -1
			}
		}
		if b.ticks >= bossIdleTicks && len(r.cards) == 0 && len(r.waves) == 0 {
			set(bossCycle[b.cycle%len(bossCycle)])
			b.cycle++
		}
	case bossVolleyWindup:
		if b.ticks >= windupTicks {
			set(bossVolley)
		}
	case bossWaveWindup:
		if b.ticks >= windupTicks {
			set(bossWave)
		}
	case bossVolley:
		if b.ticks == 1 || (r.scaleN >= 4 && b.ticks == extraAt) {
			r.fireVolley()
		}
		if b.ticks >= volleyTicks {
			set(bossIdle)
		}
	case bossWave:
		if b.ticks == 1 || (r.scaleN >= 4 && b.ticks == extraAt) {
			for _, d := range []int{-1, 1} {
				r.nextShotID++
				r.waves = append(r.waves, &Wave{ID: r.nextShotID, X: b.X + float64(d)*(bossW/2+waveW/2), Dir: d})
			}
		}
		if b.ticks >= waveTicks {
			set(bossIdle)
		}
	case bossWind:
		if b.ticks >= windTicks {
			set(bossIdle)
		}
	}
}

// fireVolley throws one card per platform height, toward the facing side (both sides with 3+ players).
// Cards fly at the body height of someone standing on a platform, above anyone standing on the floor.
func (r *Room) fireVolley() {
	dirs := []int{r.boss.Facing}
	if r.scaleN >= 3 {
		dirs = append(dirs, -r.boss.Facing)
	}
	for _, d := range dirs {
		for _, top := range r.level.PlatformTops {
			r.nextShotID++
			r.cards = append(r.cards, &Shot{
				ID: r.nextShotID, X: r.boss.X + float64(d)*(bossW/2+cardW/2), Y: top - hitH/2,
				VX: float64(d) * cardSpeed, W: cardW, H: cardH,
			})
		}
	}
}

func (r *Room) stepShots() {
	r.cards = keep(r.cards, func(s *Shot) bool {
		if !r.moveShot(s) {
			return false
		}
		for _, p := range r.players {
			if r.canHurt(p) && overlaps(shotBox(s), playerBox(p)) {
				r.hurt(p, 1, s.X)
				return false
			}
		}
		return true
	})
	r.pulses = keep(r.pulses, func(s *Shot) bool {
		if s.life--; s.life <= 0 || !r.moveShot(s) {
			return false
		}
		if r.boss.State != bossDefeat && overlaps(shotBox(s), r.bossBox()) {
			r.damageBoss(pulseDamage)
			return false
		}
		return true
	})
}

// moveShot reports false when the shot hits a solid tile.
func (r *Room) moveShot(s *Shot) bool {
	s.X += s.VX * dt
	ts := float64(r.level.Tile)
	return r.level.at(cell(s.X, ts), cell(s.Y, ts)) != tileSolid
}

func (r *Room) stepWaves() {
	r.waves = keep(r.waves, func(w *Wave) bool {
		w.X += float64(w.Dir) * waveSpeed * dt
		ts := float64(r.level.Tile)
		if r.level.at(cell(w.X+float64(w.Dir)*waveW/2, ts), cell(r.boss.Y-1, ts)) == tileSolid {
			return false
		}
		box := rect{w.X - waveW/2, r.boss.Y - waveH, w.X + waveW/2, r.boss.Y}
		for _, p := range r.players {
			if r.canHurt(p) && overlaps(box, playerBox(p)) {
				r.hurt(p, 1, w.X)
			}
		}
		return true
	})
}

func (r *Room) damageBoss(n int) {
	if r.boss.State == bossWind {
		n *= 2
	}
	r.boss.HP -= n
	r.boss.Hits++
}

func (r *Room) canHurt(p *Player) bool {
	switch p.action {
	case actDown, actDash, actRevive:
		return false
	}
	return p.client != nil && p.invuln == 0
}

func (r *Room) hurt(p *Player, dmg int, fromX float64) {
	p.HP -= dmg
	p.Hurts++
	p.Body.dashTicks = 0
	if p.HP <= 0 {
		p.HP, p.action, p.actionTicks, p.reviveHeld = 0, actDown, 0, 0
		return
	}
	dir := 1.0
	if p.X < fromX {
		dir = -1
	}
	p.action, p.actionTicks, p.invuln = actHurt, hurtTicks, invulnTicks
	p.pushVX, p.pushTicks, p.VY = dir*knockSpeed, hurtTicks, -knockHop
}

func (r *Room) allDown() bool {
	for _, p := range r.players {
		if p.client != nil && p.action != actDown {
			return false
		}
	}
	return true
}

func (r *Room) nearestTarget() *Player {
	var best *Player
	for _, p := range r.players {
		if p.client == nil || p.action == actDown {
			continue
		}
		if best == nil || math.Abs(p.X-r.boss.X) < math.Abs(best.X-r.boss.X) {
			best = p
		}
	}
	return best
}

type rect struct{ x0, y0, x1, y1 float64 }

func overlaps(a, b rect) bool { return a.x0 < b.x1 && b.x0 < a.x1 && a.y0 < b.y1 && b.y0 < a.y1 }

func playerBox(p *Player) rect { return rect{p.X - hitW/2, p.Y - hitH, p.X + hitW/2, p.Y} }

func (r *Room) bossBox() rect {
	b := r.boss
	return rect{b.X - bossW/2, b.Y - bossH, b.X + bossW/2, b.Y}
}

func shotBox(s *Shot) rect { return rect{s.X - s.W/2, s.Y - s.H/2, s.X + s.W/2, s.Y + s.H/2} }

func swingBox(p *Player) rect {
	near, far := p.X, p.X+float64(p.Facing)*(hitW/2+swingReach)
	return rect{min(near, far), p.Y - hitH + 4, max(near, far), p.Y - 4}
}

func keep[T any](s []T, f func(T) bool) []T {
	out := s[:0]
	for _, v := range s {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}
