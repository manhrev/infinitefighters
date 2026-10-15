package main

import (
	"log"
	"math"
)

// Combat, boss and match flow for the MVP level (docs/levels/01-sprint.md).
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
	reviveHoldTicks = 3 * tickRate // a teammate holds interact this long next to a downed player
	reviveReach     = 48.0
	invulnTicks     = 36 // 0.6 s after a hit
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
	coffeeRespawn   = 30 * tickRate
	coffeeSize      = 16.0

	bossBaseHP      = 120
	bossHPPerPlayer = 60 // per player beyond the first
	bossW, bossH    = 70.0, 110.0
	bossIdleTicks   = tickRate / 3
	windupTicks     = 45 // 0.75 s telegraph; well above the clients' ~70 ms render delay
	volleyTicks     = 24
	waveTicks       = 32
	windTicks       = 90  // vulnerable window: damage ×2
	enrageHPFrac    = 0.5 // below this: always a second salvo / wave, no rest between turns
	extraAt         = 16  // 3+ players: a second salvo / wave this many ticks into the attack
	cardSpeed       = 600.0
	cardW, cardH    = 16.0, 12.0
	waveSpeed       = 540.0
	waveW, waveH    = 32.0, 32.0

	walkSpeed     = 0.4 * runSpeed
	walkMaxTicks  = tickRate
	walkStopTiles = 3 // stop this close to the target
	walkMaxTiles  = 6
	wallGapTiles  = 3 // between the boss body and each wall
	chargeTiles   = 8
	chargeMinTile = 3
	chargeTicks   = 36 // 0.6 s
	recoverTicks  = 60 // 1 s, no damage

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
	actPat    = "pat" // Vought's shoulder pat: held in place

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
	patPresses   int  // Vought: attack or dash presses while held, to break free
	swingStamp   bool // Vought, Yasuka: this swing already broke a DEV stamp or a homing paper
	downTicks    int  // sandbox: ticks spent down, for the free respawn
	ping         int
	pingTicks    int
}

// counts only grow, across restarts too. Clients play one-shot animations and sounds when one changes.
type counts struct {
	Attacks, Pulses, Dashes, Hurts uint32
	Hits                           uint32 // keyboard swing or P1 smash confirmed on the boss
	Casts, SkillHits, Cancels      uint32
	Reflects, Healed, Coffees      uint32
	Rejects                        uint32 // Yasuka sent this player back to their start
}

type Boss struct {
	X, Y       float64 // feet
	Facing     int
	HP, MaxHP  float64
	State      string
	ticks      int // in the current state
	cycle      int
	W, H       float64 // hitbox
	Hits       uint32  // resets with the match; clients treat a decrease as a reset
	Marks      uint32  // P3 mark applied (also resets)
	markTicks  int
	target     *Player // walk target, chosen when the walk starts
	walkFrom   float64
	chargeFrom float64 // locked when charge_windup starts
	chargeTo   float64
	chargeHit  map[int]bool
	landX      float64 // Khabib: where the teleport lands (feet), locked while he is out of sight
	landY      float64
	VY         float64 // Khabib: falling speed
	chain      int     // Khabib: teleport hops still to land
	patTarget  *Player // Vought: who is being patted
	patHits    uint32  // Vought: Hits when the pat began; any growth is a teammate's interruption
	counterCD  int     // Vought: ticks until the shield may counter again
	dome       float64 // Yasuka: silence dome strength left
	queue      []int   // Khabib: player IDs the hops go to, in order
	nags       int     // Khabib: talk-spam words fired so far
}

type Shot struct {
	ID       int
	Kind     string  // card, pulse, stamp, returned
	X, Y, VX float64 // centre
	W, H     float64
	Dmg      float64
	life     int // player shots only
	// party: who fired it (no self-hits, kill credit, cleared on death) and whether P2 may bounce it
	owner       int
	reflectable bool
	VY, G       float64 // lobbed cards: vertical speed and gravity
	Word        string  // Khabib's talk spam: the nag drawn on the card
}

type Wave struct {
	ID  int
	X   float64 // centre
	Dir int
	Y   float64      // feet: the boss's floor, or a desk for Khabib
	hit map[int]bool // each wave hurts a player once, even without invulnerability frames
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
	stamps        []*Stamp // Vought's DEV stamps on players
	seals         []*Seal  // Yasuka's rejection seals
	papers        []*Paper // Yasuka's homing forms
	zones         []*Zone
	nextShotID    int
	bossBaseHP    int   // configurable for playtests
	hard          bool  // hard mode: falling files and enrage from the start, no invulnerability after a hit
	coffeeIn      []int // per level coffee point: ticks until it is back; 0 = on the shelf
	bossMinX      float64
	bossMaxX      float64
	drops         []*Drop
	dropHit       map[int]bool // players hit in the current drop turn
	dropCD        int          // ticks until the drop attack may be used again
}

func (r *Room) resetMatch() {
	hp := float64(r.bossBaseHP + bossHPPerPlayer*(r.scaleN-1))
	b := r.level.Boss
	r.boss = Boss{X: b.X, Y: b.Y, Facing: -1, HP: hp, MaxHP: hp, State: bossIdle, W: bossW, H: bossH}
	switch r.sandbox { // hitboxes from docs/sprite-spec.md
	case sandboxKhabib:
		r.boss.W = 80
	case sandboxVought:
		r.boss.W, r.boss.H = 36, 95
	case sandboxYasuka:
		r.boss.W, r.boss.H = 52, 100
	}
	r.cards, r.shots, r.waves, r.zones, r.drops = nil, nil, nil, nil, nil
	r.dropCD = 0
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
	left, right := r.arenaWalls()
	gap := wallGapTiles*float64(r.level.Tile) + r.boss.W/2
	return left + gap, right - gap
}

// arenaWalls is the open floor between the walls left and right of the boss spawn, in px.
func (r *Room) arenaWalls() (left, right float64) {
	l, ts := r.level, float64(r.level.Tile)
	row := cell(l.Boss.Y-1, ts)
	c0, c1 := cell(l.Boss.X, ts), cell(l.Boss.X, ts)
	for l.at(c0-1, row) != tileSolid {
		c0--
	}
	for l.at(c1+1, row) != tileSolid {
		c1++
	}
	return float64(c0) * ts, float64(c1+1) * ts
}

func (r *Room) stepMatch() {
	if r.result != "" {
		if r.resultTicks--; r.resultTicks <= 0 {
			switch {
			case r.result == resultLost:
				r.resetMatch()
			case r.partyLevel != nil:
				log.Printf("boss beaten, on to the party")
				r.startParty() // the extension: the survivors are invited to the year-end party
			default:
				r.toLobby()
			}
		}
		return
	}
	if r.sandbox == "" {
		r.deadline--
	}
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
	case r.sandbox != "":
		r.stepSandbox()
	case r.boss.HP <= 0:
		r.boss.HP, r.boss.State, r.boss.ticks, r.boss.markTicks = 0, bossDefeat, 0, 0
		r.cards, r.shots, r.waves, r.zones, r.drops = nil, nil, nil, nil, nil
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
	case actDown, actHurt, actRevive, actPat:
		if p.action == actPat && (pr.attack || pr.dash) {
			p.patPresses++
		}
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
			p.action, p.actionTicks, p.swingHit, p.swingStamp = actAttack, attackTicks, false, false
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
	if r.dropCD > 0 {
		r.dropCD--
	}
	if r.sandbox != "" {
		r.stepSandboxBoss()
		return
	}
	set := func(s string) { b.State, b.ticks = s, 0 }
	switch b.State {
	case bossIdle:
		r.faceNearest()
		if (b.ticks >= r.t(bossIdleTicks) || r.enraged()) && len(r.cards) == 0 && len(r.waves) == 0 {
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
			if next == bossVolleyWindup && r.lockDrops() {
				next = bossDropWindup // someone is camping on a desk
			}
			set(next)
		}
	case bossVolleyWindup:
		if b.ticks >= r.t(windupTicks) {
			set(bossVolley)
		}
	case bossWaveWindup:
		if b.ticks >= r.t(windupTicks) {
			set(bossWave)
		}
	case bossChargeWindup:
		if b.ticks >= r.t(windupTicks) {
			b.chargeHit = map[int]bool{}
			set(bossCharge)
		}
	case bossVolley:
		if b.ticks == 1 || ((r.scaleN >= 3 || r.enraged()) && b.ticks == r.t(extraAt)) {
			r.fireVolley()
		}
		if b.ticks >= r.t(volleyTicks) {
			set(bossIdle)
		}
	case bossWave:
		if b.ticks == 1 || ((r.scaleN >= 3 || r.enraged()) && b.ticks == r.t(extraAt)) {
			for _, d := range []int{-1, 1} {
				r.nextShotID++
				r.waves = append(r.waves, &Wave{ID: r.nextShotID, X: b.X + float64(d)*(bossW/2+waveW/2), Y: b.Y, Dir: d, hit: map[int]bool{}})
			}
		}
		if b.ticks >= r.t(waveTicks) {
			set(bossIdle)
		}
	case bossCharge:
		dur := r.chargeDur()
		b.X = b.chargeFrom + (b.chargeTo-b.chargeFrom)*float64(min(b.ticks, dur))/float64(dur)
		r.contactHurt() // only the moving body hurts, once per player per charge
		if b.ticks >= dur {
			set(bossChargeRecover)
		}
	case bossDropWindup:
		if b.ticks >= r.t(dropWindupTicks) {
			set(bossDrop)
		}
	case bossDrop:
		if r.stepDrops() {
			set(bossDropJam)
		}
	case bossDropJam:
		if b.ticks >= r.t(dropJamTicks) {
			set(bossIdle)
		}
	case bossChargeRecover:
		if b.ticks >= r.t(recoverTicks) {
			set(bossIdle)
		}
	case bossWind:
		if b.ticks >= r.t(windTicks) {
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
	if t == nil || t.client == nil || t.action == actDown || b.ticks > r.t(walkMaxTicks) {
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
	x := math.Max(r.bossMinX, math.Min(r.bossMaxX, b.X+float64(b.Facing)*walkSpeed*r.moveMul()*dt))
	if x == b.X {
		return true // at the wall gap
	}
	b.X = x
	return false
}

// lockCharge fixes the charge direction and end point at the start of the telegraph: toward the
// nearest player, or the other way when that side has less than chargeMinTile of room.
// It reports false only when neither side has room.
func (r *Room) lockCharge() bool {
	b, ts := &r.boss, float64(r.level.Tile)
	r.faceNearest()
	for _, dir := range []int{b.Facing, -b.Facing} {
		to := math.Max(r.bossMinX, math.Min(r.bossMaxX, b.X+float64(dir)*chargeTiles*ts))
		if math.Abs(to-b.X) >= chargeMinTile*ts {
			b.Facing, b.chargeFrom, b.chargeTo = dir, b.X, to
			return true
		}
	}
	return false
}

// Phase 2 starts at half HP in every mode: the boss moves twice as fast and attacks 1.5× as often
// (every boss timer at 2/3, task cards and waves 1.5× faster so the arena clears sooner).
func (r *Room) phase2() bool { return r.boss.HP < 0.5*r.boss.MaxHP }

// t scales a boss timer for the current phase.
func (r *Room) t(ticks int) int {
	if r.phase2() {
		return max(1, ticks*2/3)
	}
	return ticks
}

func (r *Room) attackMul() float64 {
	if r.phase2() {
		return 1.5
	}
	return 1
}

func (r *Room) moveMul() float64 {
	if r.phase2() {
		return 2
	}
	return 1
}

// chargeDur is the charge travel time: half as long, so twice as fast, in phase 2.
func (r *Room) chargeDur() int {
	if r.phase2() {
		return chargeTicks / 2
	}
	return chargeTicks
}

func (r *Room) enraged() bool { return r.hard || r.boss.HP < enrageHPFrac*r.boss.MaxHP }

// fireVolley throws one card per platform height, toward the facing side (both sides with 2+ players).
// Cards fly at the body height of someone standing on a platform, above anyone standing on the floor.
func (r *Room) fireVolley() {
	dirs := []int{r.boss.Facing}
	if r.scaleN >= 2 {
		dirs = append(dirs, -r.boss.Facing)
	}
	for _, d := range dirs {
		for _, top := range r.level.PlatformTops {
			r.nextShotID++
			r.cards = append(r.cards, &Shot{
				ID: r.nextShotID, Kind: "card", X: r.boss.X + float64(d)*(bossW/2+cardW/2), Y: top - hitH/2,
				VX: float64(d) * cardSpeed * r.attackMul(), W: cardW, H: cardH,
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
		if r.boss.State == vReflect && overlaps(shotBox(s), r.bossBox()) {
			r.reflectShot(s)
			return false
		}
		if r.boss.State == ySilence && overlaps(shotBox(s), r.bossBox()) {
			r.hitDome(s.Dmg)
			return false
		}
		if r.boss.State == yDodgeS && overlaps(shotBox(s), r.bossBox()) {
			r.evade() // the shot sails through the empty space she left
			return true
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
	s.VY += s.G * dt
	s.Y += s.VY * dt
	ts := float64(r.level.Tile)
	return r.level.at(cell(s.X, ts), cell(s.Y, ts)) != tileSolid
}

func (r *Room) stepWaves() {
	r.waves = keep(r.waves, func(w *Wave) bool {
		w.X += float64(w.Dir) * waveSpeed * r.attackMul() * dt
		ts := float64(r.level.Tile)
		if r.level.at(cell(w.X+float64(w.Dir)*waveW/2, ts), cell(w.Y-1, ts)) == tileSolid ||
			r.level.at(cell(w.X, ts), cell(w.Y, ts)) == tileEmpty { // a wall, or the edge of a desk
			return false
		}
		box := rect{w.X - waveW/2, w.Y - waveH, w.X + waveW/2, w.Y}
		for _, p := range r.players {
			if !w.hit[p.ID] && r.canHurt(p) && overlaps(box, playerBox(p)) {
				w.hit[p.ID] = true
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
	switch r.boss.State {
	case vReflect:
		r.counter() // the speech shield soaks the hit
		return
	case ySilence:
		r.hitDome(n) // the silence dome soaks it
		return
	case yDodgeS:
		r.evade() // the hit slides off her
		return
	}
	if r.boss.State == bossWind || r.boss.State == bossDropJam || r.boss.State == khRecover {
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
	if r.hard {
		p.invuln = 0 // hard mode: no grace after a hit, back-to-back hits land
	}
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
	if b.State == khTeleOut || b.State == khTeleWarn || b.State == vPatOut || b.State == vPatWarn {
		return rect{} // out of sight: nothing can hit him
	}
	return rect{b.X - b.W/2, b.Y - b.H, b.X + b.W/2, b.Y}
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
