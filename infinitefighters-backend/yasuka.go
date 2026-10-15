package main

import "math"

// Yasuka sandbox (docs/levels/sandbox-yasuka.md): the Khabib sandbox rules with Yasuka as the boss.
// She loops three attacks: a counter-stance that sidesteps the first hit, rejection seals that send
// players back to the start, and a silence dome that bursts into paper when it breaks.
// Concept art: assets/concepts/boss-yasuka-v1.png.

const (
	yIdleTicks   = tickRate / 2
	yDodgeWind   = 30
	yDodgeTicks  = 150 // 2.5 s waiting for someone to try
	yEvadeDist   = 200.0
	yStampWind   = 36
	yStampTicks  = 20
	ySealWarn    = 48  // 0.8 s before a seal bites
	ySealLife    = 240 // then 4 s active
	ySealW       = 32.0
	ySealH       = 24.0
	ySealsAhead  = 3 // seals laid in a row in front of her, besides one under each player
	yDomeWind    = 30
	yDomeHP      = 8.0
	yDomeTicks   = 300 // 5 s
	yFlurryWind  = 30
	yFlurryTicks = 20
	yRejectGrace = 30
	yFormsWind   = 45
	yFormsTicks  = 20
	paperSpeed   = 220.0 // slower than a run: it can be outrun, not forgotten
	paperTurn    = 0.07  // how much of the way to the new heading it turns each tick
	paperLife    = 5 * tickRate
	paperW       = 24.0
	paperH       = 18.0
)

const (
	yDodgeWindS  = "dodge_windup"
	yDodgeS      = "dodge"
	yProcWind    = "procedure_windup"
	yProc        = "procedure"
	ySilWind     = "silence_windup"
	ySilence     = "silence"
	yFlurryWindS = "flurry_windup"
	yFlurryS     = "flurry"
	yFormsWindS  = "forms_windup"
	yFormsS      = "forms"
)

var yCycle = []string{yDodgeWindS, yProcWind, ySilWind, yFormsWindS}

// Seal is a rejection stamp on the floor or a desk. Touching it, once it is awake, sends a player back to their start.
type Seal struct {
	ID         int
	X, Y       float64 // bottom-centre
	warn, life int
}

func (r *Room) stepYasuka() {
	b := &r.boss
	set := func(s string) { b.State, b.ticks = s, 0 }
	r.settleBoss()
	r.stepSeals()
	r.stepPapers()
	switch b.State {
	case bossIdle:
		r.faceNearest()
		if b.ticks >= r.t(yIdleTicks) && len(r.cards) == 0 && len(r.papers) == 0 {
			b.target, b.walkFrom = r.nearestTarget(), b.X
			set(bossWalk)
		}
	case bossWalk:
		if r.walkBoss() {
			set(yCycle[b.cycle%len(yCycle)])
			b.cycle++
		}
	case yFormsWindS: // gửi lại biểu mẫu: one homing form for every player
		r.faceNearest()
		if b.ticks >= r.t(yFormsWind) {
			set(yFormsS)
		}
	case yFormsS:
		if b.ticks == 1 {
			r.throwForms()
		}
		if b.ticks >= r.t(yFormsTicks) {
			set(bossIdle)
		}
	case yDodgeWindS: // né yêu cầu: hands behind her back, waiting
		r.faceNearest()
		if b.ticks >= r.t(yDodgeWind) {
			set(yDodgeS)
		}
	case yDodgeS: // the first hit slides off her (see evade); nobody trying ends it
		r.faceNearest()
		if b.ticks >= r.t(yDodgeTicks) {
			set(bossIdle)
		}
	case yProcWind: // đúng quy trình
		r.faceNearest()
		if b.ticks >= r.t(yStampWind) {
			set(yProc)
		}
	case yProc:
		if b.ticks == 1 {
			r.stampSeals()
		}
		if b.ticks >= r.t(yStampTicks) {
			set(bossIdle)
		}
	case ySilWind: // im lặng là vàng
		if b.ticks >= r.t(yDomeWind) {
			b.dome = yDomeHP
			set(ySilence)
		}
	case ySilence:
		switch {
		case b.dome <= 0: // the dome broke: the backlog bursts out
			set(yFlurryWindS)
		case b.ticks >= r.t(yDomeTicks): // nobody broke it
			b.dome = 0
			set(khRecover)
		}
	case yFlurryWindS:
		if b.ticks >= r.t(yFlurryWind) {
			set(yFlurryS)
		}
	case yFlurryS:
		if b.ticks == 1 {
			r.flurry()
		}
		if b.ticks >= r.t(yFlurryTicks) {
			set(khRecover)
		}
	case khRecover:
		if b.ticks >= r.t(khRecoverTicks) {
			set(bossIdle)
		}
	}
}

// evade makes the counter-stance slide away from the hit, which is then wasted: she is open afterwards.
func (r *Room) evade() {
	b := &r.boss
	dir := 1.0
	if t := r.nearestTarget(); t != nil && t.X > b.X {
		dir = -1
	}
	x := max(r.bossMinX, min(r.bossMaxX, b.X+dir*yEvadeDist))
	if x == b.X { // cornered: slide the other way
		x = max(r.bossMinX, min(r.bossMaxX, b.X-dir*yEvadeDist))
	}
	b.X, b.State, b.ticks = x, khRecover, 0
}

// hitDome wears the silence dome down; the boss behind it takes nothing.
func (r *Room) hitDome(n float64) { r.boss.dome -= max(n, 1) }

// stampSeals lays a seal under each player and a row in front of her.
func (r *Room) stampSeals() {
	b, ts := &r.boss, float64(r.level.Tile)
	add := func(x, y float64) {
		r.nextShotID++
		r.seals = append(r.seals, &Seal{ID: r.nextShotID, X: x, Y: y, warn: r.t(ySealWarn), life: ySealLife})
	}
	for _, p := range r.alive() {
		add(p.X, r.groundBelow(p.X, p.Y))
	}
	for i := 1; i <= ySealsAhead; i++ {
		x := b.X + float64(b.Facing)*(b.W/2+float64(i)*3*ts)
		if x > ts && x < float64(r.level.W-1)*ts {
			add(x, r.level.Boss.Y)
		}
	}
}

func (r *Room) stepSeals() {
	r.seals = keep(r.seals, func(s *Seal) bool {
		if s.warn > 0 {
			s.warn--
			return true
		}
		box := rect{s.X - ySealW/2, s.Y - ySealH, s.X + ySealW/2, s.Y}
		for _, p := range r.players {
			if r.canHurt(p) && overlaps(box, playerBox(p)) {
				r.reject(p)
			}
		}
		s.life--
		return s.life > 0
	})
}

// reject sends the player back to their start, as if the form came back unsigned.
func (r *Room) reject(p *Player) {
	sp := r.level.Spawns[p.Slot%len(r.level.Spawns)]
	p.X, p.Y, p.VX, p.VY, p.invuln = sp.X, sp.Y, 0, 0, yRejectGrace
	p.Body.dashTicks = 0
	p.Rejects++
}

// flurry bursts the backlog out as paper along every desk height, both ways. The floor stays safe.
func (r *Room) flurry() {
	b := &r.boss
	for _, d := range []float64{-1, 1} {
		for _, top := range r.level.PlatformTops {
			r.nextShotID++
			r.cards = append(r.cards, &Shot{
				ID: r.nextShotID, Kind: "card", Word: "HO SO", X: b.X + d*(b.W/2+vCardW/2), Y: top - hitH/2,
				VX: d * cardSpeed * r.attackMul(), W: vCardW, H: vCardH,
			})
		}
	}
}

// Paper is a form that chases one player.
type Paper struct {
	ID     int
	X, Y   float64 // centre
	VX, VY float64
	target *Player
	life   int
}

// throwForms sends one homing paper at every player, fanned out so they do not stack.
func (r *Room) throwForms() {
	b := &r.boss
	for i, p := range r.alive() {
		r.nextShotID++
		a := float64(i%3-1) * 0.6 // fan: up, level, down-ish
		r.papers = append(r.papers, &Paper{
			ID: r.nextShotID, X: b.X + float64(b.Facing)*b.W/2, Y: b.Y - b.H*0.6,
			VX: float64(b.Facing) * paperSpeed * (1 - 0.5*abs(a)), VY: -paperSpeed * a, target: p, life: paperLife,
		})
	}
}

// stepPapers steers each paper toward its player, hurts on touch, and lets players swat or shoot them down.
func (r *Room) stepPapers() {
	r.papers = keep(r.papers, func(pp *Paper) bool {
		if pp.life--; pp.life <= 0 {
			return false
		}
		t := pp.target
		if t.client == nil || t.action == actDown { // the target is out: chase the nearest player instead
			t = nil
			for _, q := range r.alive() {
				if t == nil || abs(q.X-pp.X) < abs(t.X-pp.X) {
					t = q
				}
			}
			if t == nil {
				return false
			}
			pp.target = t
		}
		dx, dy := t.X-pp.X, (t.Y-hitH/2)-pp.Y
		if d := math.Hypot(dx, dy); d > 0 {
			sp := paperSpeed * r.attackMul()
			pp.VX += (dx/d*sp - pp.VX) * paperTurn
			pp.VY += (dy/d*sp - pp.VY) * paperTurn
		}
		pp.X += pp.VX * dt
		pp.Y += pp.VY * dt
		box := rect{pp.X - paperW/2, pp.Y - paperH/2, pp.X + paperW/2, pp.Y + paperH/2}
		for _, q := range r.players { // keyboard swings, once per swing
			elapsed := attackTicks - q.actionTicks
			if q.action == actAttack && !q.swingStamp && elapsed >= swingFrom && elapsed <= swingTo && overlaps(reachBox(q, swingReach), box) {
				q.swingStamp = true
				return false
			}
		}
		swatted := false
		r.shots = keep(r.shots, func(sh *Shot) bool { // code pulses and other shots
			if !swatted && overlaps(shotBox(sh), box) {
				swatted = true
				return false
			}
			return true
		})
		if swatted {
			return false
		}
		for _, q := range r.players {
			if r.canHurt(q) && overlaps(box, playerBox(q)) {
				r.hurt(q, 1, pp.X)
				return false
			}
		}
		return true
	})
}
