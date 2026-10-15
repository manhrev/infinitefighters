package main

// Vought sandbox (docs/levels/sandbox-vought.md): the Khabib sandbox rules with SA Vought as the
// boss. He loops four attacks: lobbed tasks, a shoulder pat, a DEV stamp and a reflecting shield.
// Concept art: assets/concepts/boss-sa-vought-v1.png.

const (
	vIdleTicks     = tickRate / 2
	vLobEvery      = 10
	vLobTicks      = 55  // 5 lobs fit
	vLobFlight     = 0.7 // seconds in the air
	vLobGravity    = 1500.0
	vCardW, vCardH = 72.0, 16.0
	vPatOutTicks   = 24
	vPatWarnTicks  = 30
	vPatHoldTicks  = 90 // he keeps the listener for 1.5 s
	vPatEscape     = 6  // attack or dash presses that break free
	vPatRelease    = 30 // invulnerable ticks after the pat
	vBurdenTicks   = 5 * tickRate
	vStampThrow    = 20
	vStampHP       = 3
	vStampTicks    = 210 // 3.5 s to break it
	vReflectWind   = 36
	vReflectTicks  = 120
	vCounterEvery  = 15
)

const (
	vDeflectWind = "deflect_windup"
	vDeflect     = "deflect"
	vPatOut      = "pat_out"
	vPatWarn     = "pat_warn"
	vPat         = "shoulder_pat"
	vStampWind   = "stamp_windup"
	vStamp       = "stamp_throw"
	vReflectWait = "reflect_windup"
	vReflect     = "reflect_shield"
)

var vCycle = []string{vDeflectWind, vPatOut, vStampWind, vReflectWait}

var vLobWords = []string{"DEV LO NHE", "DEV CHU DONG", "PHAN NAY DEV"}

// Stamp is a DEV stamp over a player. It hits after ticks unless hits break it first.
type Stamp struct {
	ID     int
	target *Player
	hp     int
	ticks  int
}

// stepSandboxBoss runs the boss of the sandbox the room was started with.
func (r *Room) stepSandboxBoss() {
	switch r.sandbox {
	case sandboxVought:
		r.stepVought()
		return
	case sandboxYasuka:
		r.stepYasuka()
		return
	}
	r.stepKhabib()
}

func (r *Room) stepVought() {
	b := &r.boss
	set := func(s string) { b.State, b.ticks = s, 0 }
	if b.counterCD > 0 {
		b.counterCD--
	}
	r.settleBoss()
	r.stepStamps()
	switch b.State {
	case bossIdle:
		r.faceNearest()
		if b.ticks >= r.t(vIdleTicks) && len(r.cards) == 0 {
			b.target, b.walkFrom = r.nearestTarget(), b.X
			set(bossWalk)
		}
	case bossWalk:
		if r.walkBoss() {
			set(vCycle[b.cycle%len(vCycle)])
			b.cycle++
			b.nags = 0
		}
	case vDeflectWind: // múa mồm chuyển việc: tasks lobbed over to the devs
		r.faceNearest()
		if b.ticks >= r.t(windupTicks) {
			set(vDeflect)
		}
	case vDeflect:
		r.faceNearest()
		if b.ticks%r.t(vLobEvery) == 1 {
			r.lob()
		}
		if b.ticks >= r.t(vLobTicks) {
			set(bossIdle)
		}
	case vPatOut: // vỗ vai tâm sự: gone, then right behind a player
		r.faceNearest()
		if b.ticks >= r.t(vPatOutTicks) {
			if b.patTarget = r.nearestTarget(); b.patTarget == nil {
				set(bossIdle)
				return
			}
			b.landX, b.landY = r.behind(b.patTarget)
			set(vPatWarn)
		}
	case vPatWarn:
		if b.ticks >= r.t(vPatWarnTicks) {
			p := b.patTarget
			b.X, b.Y, b.landX, b.landY = b.landX, b.landY, 0, 0
			if p == nil || !r.canHurt(p) {
				set(khRecover) // the listener got away
				return
			}
			b.Facing = 1
			if p.X < b.X {
				b.Facing = -1
			}
			p.action, p.actionTicks, p.patPresses = actPat, 3*tickRate, 0
			p.Body.dashTicks = 0
			b.patHits = b.Hits
			set(vPat)
		}
	case vPat:
		p := b.patTarget
		switch {
		case p == nil || p.action != actPat: // hurt by something else, or gone
			set(khRecover)
		case b.Hits != b.patHits || p.patPresses >= vPatEscape: // a teammate hit him, or the listener broke free
			r.endPat(p, false)
			set(khRecover)
		case b.ticks >= r.t(vPatHoldTicks): // the task lands on their back
			r.endPat(p, true)
			set(khRecover)
		}
	case vStampWind: // sang tên bug
		r.faceNearest()
		if b.ticks >= r.t(windupTicks) {
			set(vStamp)
		}
	case vStamp:
		if b.ticks == 1 {
			r.stampPlayer()
		}
		if b.ticks >= r.t(vStampThrow) {
			set(bossIdle)
		}
	case vReflectWait: // phản đòn yêu cầu: he inhales, then the shield opens
		r.faceNearest()
		if b.ticks >= r.t(vReflectWind) {
			set(vReflect)
		}
	case vReflect:
		if b.ticks >= r.t(vReflectTicks) {
			set(khRecover)
		}
	case khRecover:
		if b.ticks >= r.t(khRecoverTicks) {
			set(bossIdle)
		}
	}
}

// alive lists the players he can target.
func (r *Room) alive() []*Player {
	var out []*Player
	for _, p := range r.players {
		if p.client != nil && p.action != actDown {
			out = append(out, p)
		}
	}
	return out
}

// lob throws one task over in an arc to a player; alone, to three spots around them.
func (r *Room) lob() {
	b := &r.boss
	ps := r.alive()
	if len(ps) == 0 {
		return
	}
	p, dx := ps[b.nags%len(ps)], 0.0
	if len(ps) == 1 {
		dx = float64(b.nags%3-1) * 70
	}
	ts := float64(r.level.Tile)
	tx := max(ts, min(float64(r.level.W-1)*ts, p.X+dx))
	ty := p.Y - hitH/2
	x0, y0 := b.X, b.Y-b.H*0.7
	T := vLobFlight
	r.nextShotID++
	r.cards = append(r.cards, &Shot{
		ID: r.nextShotID, Kind: "card", Word: vLobWords[b.nags%len(vLobWords)],
		X: x0, Y: y0, VX: (tx - x0) / T, VY: (ty-y0)/T - 0.5*vLobGravity*T, G: vLobGravity, W: vCardW, H: vCardH,
	})
	b.nags++
}

// endPat frees the listener. A full pat leaves a task on their back: no dash and no skill for a while.
func (r *Room) endPat(p *Player, tasked bool) {
	p.action, p.actionTicks, p.invuln = "", 0, vPatRelease
	if tasked {
		p.dashCD = max(p.dashCD, vBurdenTicks)
		p.skillCD = max(p.skillCD, vBurdenTicks)
	}
}

// stampPlayer puts a DEV stamp on the nearest player who has none.
func (r *Room) stampPlayer() {
	var best *Player
	for _, p := range r.alive() {
		if r.hasStamp(p) {
			continue
		}
		if best == nil || abs(p.X-r.boss.X) < abs(best.X-r.boss.X) {
			best = p
		}
	}
	if best == nil {
		return
	}
	r.nextShotID++
	r.stamps = append(r.stamps, &Stamp{ID: r.nextShotID, target: best, hp: vStampHP, ticks: vStampTicks})
}

func (r *Room) hasStamp(p *Player) bool {
	for _, s := range r.stamps {
		if s.target == p {
			return true
		}
	}
	return false
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func stampBox(p *Player) rect { return rect{p.X - 18, p.Y - hitH - 30, p.X + 18, p.Y - hitH + 20} }

// stepStamps breaks stamps that were hit, and hands the bug to the player when one runs out.
func (r *Room) stepStamps() {
	r.stamps = keep(r.stamps, func(s *Stamp) bool {
		p := s.target
		if p.client == nil || p.action == actDown {
			return false
		}
		box := stampBox(p)
		for _, q := range r.players { // keyboard swings, once per swing
			elapsed := attackTicks - q.actionTicks
			if q.action == actAttack && !q.swingStamp && elapsed >= swingFrom && elapsed <= swingTo && overlaps(reachBox(q, swingReach), box) {
				q.swingStamp = true
				s.hp--
			}
		}
		r.shots = keep(r.shots, func(sh *Shot) bool { // code pulses and other shots
			if overlaps(shotBox(sh), box) {
				s.hp--
				return false
			}
			return true
		})
		if s.hp <= 0 {
			return false
		}
		if s.ticks--; s.ticks <= 0 { // the bug is yours now
			if r.canHurt(p) {
				r.hurt(p, 1, r.boss.X)
			}
			p.dashCD = max(p.dashCD, vBurdenTicks)
			p.skillCD = max(p.skillCD, vBurdenTicks)
			return false
		}
		return true
	})
}

// counter answers a hit on the shield with a red task at the nearest player, along their lane.
func (r *Room) counter() {
	b := &r.boss
	if b.counterCD > 0 {
		return
	}
	var t *Player
	for _, p := range r.alive() {
		if t == nil || abs(p.X-b.X) < abs(t.X-b.X) {
			t = p
		}
	}
	if t == nil {
		return
	}
	b.counterCD = vCounterEvery
	dir := 1.0
	if t.X < b.X {
		dir = -1
	}
	r.nextShotID++
	r.cards = append(r.cards, &Shot{
		ID: r.nextShotID, Kind: "card", Word: "EM LAM!", X: b.X + dir*(b.W/2+vCardW/2), Y: t.Y - hitH/2,
		VX: dir * cardSpeed, W: vCardW, H: vCardH,
	})
}

// reflectShot turns a player's shot into a red task flying back along its own lane.
func (r *Room) reflectShot(s *Shot) {
	r.nextShotID++
	dir := -1.0
	if s.VX < 0 {
		dir = 1
	}
	r.cards = append(r.cards, &Shot{
		ID: r.nextShotID, Kind: "card", Word: "EM LAM!", X: s.X, Y: s.Y, VX: dir * cardSpeed, W: vCardW, H: vCardH,
	})
}
