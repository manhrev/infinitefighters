package main

import "math"

// Combat, boss and match flow for the MVP level (dev/game-scenario.md).
//
// Rules this file guarantees:
//   - Boss HP and attack directions are computed once at match start from every slot in the
//     room, connected or not; a restart after a loss keeps them.
//   - Only one boss attack is ever live: the boss walks and starts the next windup only after
//     all hazards are gone.
//   - Every attack leaves a safe zone: volleys fly at platform heights (the floor is safe),
//     waves and charges run along the floor (any platform is safe).
//   - The boss keeps at least 3 tiles between its body and each wall; a charge with less than
//     3 tiles of room becomes a wind_clock before any telegraph is shown.
//   - Away (disconnected) and downed players are ignored by hazards and boss aiming,
//     and away players do not count for "everyone is down".

// Action lengths match the art (assets/sprites/manifest.json): frames × tickRate / fps.
const (
	attackTicks = 5 * tickRate / 18
	dashTicks   = 4 * tickRate / 16
	hurtTicks   = 2 * tickRate / 10
	reviveTicks = 4 * tickRate / 10
	castTicks   = 4 * tickRate / 16 // skill-cast.png
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
	pulseDamage     = 2.0
	pingTicks       = 2 * tickRate
	coffeeRespawn   = 20 * tickRate
	coffeeSize      = 16.0

	bossBaseHP      = 120
	bossHPPerPlayer = 60 // per player beyond the first
	bossW, bossH    = 70.0, 110.0
	bossIdleTicks   = tickRate
	windupTicks     = 54 // telegraph; well above the clients' ~70 ms render delay
	volleyTicks     = 30
	waveTicks       = 40
	windTicks       = 150 // vulnerable window: damage ×2
	extraAt         = 24  // 4+ players: a second salvo / wave this many ticks into the attack
	cardSpeed       = 480.0
	cardW, cardH    = 16.0, 12.0
	waveSpeed       = 420.0
	waveW, waveH    = 32.0, 32.0

	walkSpeed     = 0.4 * runSpeed
	walkMaxTicks  = 2 * tickRate
	walkStopTiles = 3 // stop this close to the target
	walkMaxTiles  = 6
	wallGapTiles  = 3 // between the boss body and each wall
	chargeTiles   = 8
	chargeMinTile = 3
	chargeTicks   = 36 // 0.6 s
	recoverTicks  = 90 // 1.5 s, no damage

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

	bossIdle          = "idle"
	bossWalk          = "walk"
	bossVolleyWindup  = "volley_windup"
	bossVolley        = "volley"
	bossWaveWindup    = "wave_windup"
	bossWave          = "wave"
	bossChargeWindup  = "charge_windup"
	bossCharge        = "charge"
	bossChargeRecover = "charge_recover"
	bossWind          = "wind_clock"
	bossDefeat        = "defeat"

	resultWon  = "won"
	resultLost = "lost"
)

// A charge ("Họp nhanh 5 phút") after every two volley/wave turns; the wind_clock window closes the loop.
var bossCycle = []string{
	bossVolleyWindup, bossWaveWindup, bossChargeWindup,
	bossVolleyWindup, bossWaveWindup, bossChargeWindup, bossWind,
}

// fighter is a player's per-match combat state; a restart resets it whole.
type fighter struct {
	HP           int
	action       string
	actionTicks  int
	swingHit     bool
	invuln       int
	dashCD       int
	pulseCD      int
	skillCD      int
	charge       int  // P1: ticks held
	reflectsLeft int  // P2: during a shield
	castHit      bool // P1 smash / P5 timecard: hits once per cast
	chargeDmg    float64
	reviveHeld   int
	ping         int
	pingTicks    int
}

// counts only grow, across restarts too. Clients play one-shot animations and sounds when one changes.
type counts struct {
	Attacks, Pulses, Dashes, Hurts uint32
	Hits                           uint32 // keyboard swing or P1 smash confirmed on the boss
	Casts, SkillHits, Cancels      uint32
	Reflects, Healed, Coffees      uint32
}

type Boss struct {
	X, Y       float64 // feet
	Facing     int
	HP, MaxHP  float64
	State      string
	ticks      int // in the current state
	cycle      int
	Hits       uint32 // resets with the match; clients treat a decrease as a reset
	Marks      uint32 // P3 mark applied (also resets)
	markTicks  int
	target     *Player // walk target, chosen when the walk starts
	walkFrom   float64
	chargeFrom float64 // locked when charge_windup starts
	chargeTo   float64
	chargeHit  map[int]bool
}

type Shot struct {
	ID       int
	Kind     string  // card, pulse, stamp, returned
	X, Y, VX float64 // centre
	W, H     float64
	Dmg      float64
	life     int // player shots only
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
	cards         []*Shot // boss task cards: hurt players
	shots         []*Shot // players' shots: hurt the boss
	waves         []*Wave
	zones         []*Zone
	nextShotID    int
	bossBaseHP    int   // configurable for playtests
	coffeeIn      []int // per level coffee point: ticks until it is back; 0 = on the shelf
	bossMinX      float64
	bossMaxX      float64
}

func (r *Room) resetMatch() {
	hp := float64(r.bossBaseHP + bossHPPerPlayer*(r.scaleN-1))
	b := r.level.Boss
	r.boss = Boss{X: b.X, Y: b.Y, Facing: -1, HP: hp, MaxHP: hp, State: bossIdle}
	r.cards, r.shots, r.waves, r.zones = nil, nil, nil, nil
	r.coffeeIn = make([]int, len(r.level.Coffees))
	r.deadline, r.result, r.resultTicks = r.deadlineTicks, "", 0
	r.bossMinX, r.bossMaxX = r.bossRange()
	for i, p := range r.players {
		s := r.level.Spawns[i%len(r.level.Spawns)]
		p.Body = Body{X: s.X, Y: s.Y, Facing: 1}
		p.fighter = fighter{HP: playerHP}
		p.pressed = presses{}
	}
}

// bossRange finds the walls left and right of the boss spawn and keeps the wall gap from both.
func (r *Room) bossRange() (lo, hi float64) {
	l, ts := r.level, float64(r.level.Tile)
	row := cell(l.Boss.Y-1, ts)
	left, right := cell(l.Boss.X, ts), cell(l.Boss.X, ts)
	for l.at(left-1, row) != tileSolid {
		left--
	}
	for l.at(right+1, row) != tileSolid {
		right++
	}
	gap := wallGapTiles*ts + bossW/2
	return float64(left)*ts + gap, float64(right+1)*ts - gap
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
	r.stepCoffee()
	r.stepZones()
	r.stepBoss()
	r.stepShots()
	r.stepWaves()

	switch {
	case r.boss.HP <= 0:
		r.boss.HP, r.boss.State, r.boss.ticks, r.boss.markTicks = 0, bossDefeat, 0, 0
		r.cards, r.shots, r.waves, r.zones = nil, nil, nil, nil
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
	for _, t := range []*int{&p.invuln, &p.dashCD, &p.pulseCD, &p.skillCD} {
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
	case actCharge:
		r.stepCharge(p, in, pr)
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
			p.action, p.actionTicks, p.dashCD = actDash, dashTicks, dashCooldown
			p.Body.dashTicks, p.Body.dashSpeed = dashTicks, dashSpeed
			if p.Slot == slotOrange { // shares one cooldown with "Em về đúng giờ"
				p.skillCD = max(p.skillCD, dashCooldown)
			}
			p.Dashes++
		case pr.special && p.skillCD == 0:
			if dir != 0 {
				p.Facing = dir
			}
			r.startSkill(p)
		case pr.attack:
			p.action, p.actionTicks, p.swingHit = actAttack, attackTicks, false
			p.Attacks++
		case pr.skill && p.pulseCD == 0:
			p.action, p.actionTicks, p.pulseCD = actPulse, attackTicks, pulseCooldown
			p.Pulses++
			r.addShot(p, "pulse", 32, 16, pulseSpeed, pulseDamage, pulseLife)
		}
	}

	facing := p.Facing
	p.slow = p.action == actCharge
	p.step(r.level, in, pr.jump && !isDashing(p.action))
	switch p.action {
	case actAttack, actPulse, actSmash, actShield, actStamp, actHotfix:
		p.Facing = facing // no turning mid-action
	}
	if p.action == actTimecard && p.Body.dashTicks == 0 {
		r.endTimecard(p) // stopped at a wall
	}
	if r.boss.State == bossDefeat {
		return
	}
	elapsed := attackTicks - p.actionTicks
	if p.action == actAttack && !p.swingHit && elapsed >= swingFrom && elapsed <= swingTo &&
		overlaps(reachBox(p, swingReach), r.bossBox()) {
		p.swingHit = true
		p.Hits++
		r.damageBoss(1, true)
	}
	r.skillHits(p)
}

func isDashing(action string) bool { return action == actDash || action == actTimecard }

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

// stepCoffee heals 1 HP to the first hurt player touching a coffee; it comes back later.
// Full-HP players walk past it, so it is not wasted.
func (r *Room) stepCoffee() {
	for i, c := range r.level.Coffees {
		if r.coffeeIn[i] > 0 {
			r.coffeeIn[i]--
			continue
		}
		box := rect{c.X - coffeeSize/2, c.Y - coffeeSize, c.X + coffeeSize/2, c.Y}
		for _, p := range r.players {
			if p.client != nil && p.action != actDown && p.HP < playerHP && overlaps(box, playerBox(p)) {
				p.HP++
				p.Coffees++
				r.coffeeIn[i] = coffeeRespawn
				break
			}
		}
	}
}

func (r *Room) stepBoss() {
	b := &r.boss
	b.ticks++
	if b.markTicks > 0 {
		b.markTicks--
	}
	set := func(s string) { b.State, b.ticks = s, 0 }
	switch b.State {
	case bossIdle:
		r.faceNearest()
		if b.ticks >= bossIdleTicks && len(r.cards) == 0 && len(r.waves) == 0 {
			b.target, b.walkFrom = r.nearestTarget(), b.X
			set(bossWalk)
		}
	case bossWalk:
		if r.walkBoss() {
			next := bossCycle[b.cycle%len(bossCycle)]
			b.cycle++
			if next == bossChargeWindup && !r.lockCharge() {
				next = bossWind // no room to charge safely
			}
			set(next)
		}
	case bossVolleyWindup:
		if b.ticks >= windupTicks {
			set(bossVolley)
		}
	case bossWaveWindup:
		if b.ticks >= windupTicks {
			set(bossWave)
		}
	case bossChargeWindup:
		if b.ticks >= windupTicks {
			b.chargeHit = map[int]bool{}
			set(bossCharge)
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
	case bossCharge:
		b.X = b.chargeFrom + (b.chargeTo-b.chargeFrom)*float64(b.ticks)/chargeTicks
		for _, p := range r.players { // only the moving body hurts, once per player per charge
			if !b.chargeHit[p.ID] && r.canHurt(p) && overlaps(playerBox(p), r.bossBox()) {
				b.chargeHit[p.ID] = true
				r.hurt(p, 1, b.X)
			}
		}
		if b.ticks >= chargeTicks {
			set(bossChargeRecover)
		}
	case bossChargeRecover:
		if b.ticks >= recoverTicks {
			set(bossIdle)
		}
	case bossWind:
		if b.ticks >= windTicks {
			set(bossIdle)
		}
	}
}

func (r *Room) faceNearest() {
	if t := r.nearestTarget(); t != nil {
		r.boss.Facing = 1
		if t.X < r.boss.X {
			r.boss.Facing = -1
		}
	}
}

// walkBoss moves toward the target chosen at the start of the walk. It reports true when the walk is over.
func (r *Room) walkBoss() bool {
	b, ts := &r.boss, float64(r.level.Tile)
	t := b.target
	if t == nil || t.client == nil || t.action == actDown || b.ticks > walkMaxTicks {
		return true
	}
	dx := t.X - b.X
	if math.Abs(dx) <= walkStopTiles*ts || math.Abs(b.X-b.walkFrom) >= walkMaxTiles*ts {
		return true
	}
	b.Facing = 1
	if dx < 0 {
		b.Facing = -1
	}
	x := math.Max(r.bossMinX, math.Min(r.bossMaxX, b.X+float64(b.Facing)*walkSpeed*dt))
	if x == b.X {
		return true // at the wall gap
	}
	b.X = x
	return false
}

// lockCharge fixes the charge direction and end point at the start of the telegraph.
// It reports false when there is less than chargeMinTile of room that way.
func (r *Room) lockCharge() bool {
	b, ts := &r.boss, float64(r.level.Tile)
	r.faceNearest()
	to := math.Max(r.bossMinX, math.Min(r.bossMaxX, b.X+float64(b.Facing)*chargeTiles*ts))
	if math.Abs(to-b.X) < chargeMinTile*ts {
		return false
	}
	b.chargeFrom, b.chargeTo = b.X, to
	return true
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
				ID: r.nextShotID, Kind: "card", X: r.boss.X + float64(d)*(bossW/2+cardW/2), Y: top - hitH/2,
				VX: float64(d) * cardSpeed, W: cardW, H: cardH,
			})
		}
	}
}

func (r *Room) addShot(p *Player, kind string, w, h, speed, dmg float64, life int) {
	r.nextShotID++
	r.shots = append(r.shots, &Shot{
		ID: r.nextShotID, Kind: kind, X: p.X + float64(p.Facing)*(hitW/2+w/2), Y: p.Y - hitH/2,
		VX: float64(p.Facing) * speed, W: w, H: h, Dmg: dmg, life: life,
	})
}

func (r *Room) stepShots() {
	r.cards = keep(r.cards, func(s *Shot) bool {
		if !r.moveShot(s) {
			return false
		}
		for _, p := range r.players {
			if r.reflect(p, s) {
				return false
			}
		}
		for _, p := range r.players {
			if r.canHurt(p) && overlaps(shotBox(s), playerBox(p)) {
				r.hurt(p, 1, s.X)
				return false
			}
		}
		return true
	})
	r.shots = keep(r.shots, func(s *Shot) bool {
		if s.life--; s.life <= 0 || !r.moveShot(s) {
			return false
		}
		if r.boss.State != bossDefeat && overlaps(shotBox(s), r.bossBox()) {
			r.damageBoss(s.Dmg, s.Kind == "pulse")
			if s.Kind == "stamp" {
				r.boss.markTicks = markTicks // refresh, never stack
				r.boss.Marks++
			}
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

// damageBoss applies the P3 mark (keyboard and code-pulse hits only) and the wind_clock window.
func (r *Room) damageBoss(n float64, markable bool) {
	if markable && r.boss.markTicks > 0 {
		n *= markBonus
	}
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
	case actTimecard:
		if p.actionTicks > timecardTicks-timecardInvuln {
			return false
		}
	}
	return p.client != nil && p.invuln == 0
}

func (r *Room) hurt(p *Player, dmg int, fromX float64) {
	if p.action == actCharge {
		r.cancelCharge(p)
	}
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

// reachBox is the area in front of a player, from their centre to reach past the hitbox edge.
func reachBox(p *Player, reach float64) rect {
	near, far := p.X, p.X+float64(p.Facing)*(hitW/2+reach)
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
