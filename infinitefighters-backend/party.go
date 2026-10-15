package main

import (
	"math"
	"math/rand/v2"
	"sort"
)

// Party PvP level "Anh em cây khế" (dev/game-scenario.md, extension "Anh em hết mình, tiền tình hết hồn").
// It follows a boss win, or the host starts it directly from the lobby.
// Everyone fights everyone for coins: grab coins the safe sprays, hit rivals to knock coins out of their
// wallets, and deposit at an ATM. Only deposited coins count when the 180 s run out.
//
// It runs instead of the co-op rules (stepMatch): its own damage, skills and money, reusing only the
// physics, input latching and geometry helpers. Rules this file guarantees:
//   - Coins are conserved: everything the safe sprays is on the floor, in a wallet or in a balance.
//   - A hit that lands drops ceil(10%) of the wallet; a kill drops ceil(50%) of the wallet before the
//     hit and nothing more. Hits blocked by invulnerability (after a hit, dashing, spawn protection)
//     drop nothing and do not cancel a deposit.
//   - Each swing or cast hits each target at most once. Nobody hits themselves.
//   - Offline players can be hit for 5 s, then drop the rest of their wallet and leave the field.

const (
	modeCoop      = "coop"
	modeParty     = "party"
	partyLevelURL = "/levels/party.json"

	pvpHP            = 100.0
	partyCountdown   = 5 * tickRate // rules on screen, nothing can be hit, picked up or deposited
	partyTicks       = 180 * tickRate
	partyResultTicks = 10 * tickRate
	sprayEvery       = 15 * tickRate
	sprayLast        = 135 * tickRate // sprays at 0, 15, ..., 135 s
	sprayPerPlayer   = 50
	sprayBundles     = 5
	finalBurstAt     = 150 * tickRate
	finalPerPlayer   = 150
	finalBundles     = 15
	bundleSize       = 16.0
	bundleLock       = tickRate // a dropped bundle's previous owner waits 1 s
	atmTicks         = 3 * tickRate
	atmReach         = 32.0 // 1 tile
	pvpRespawnTicks  = 3 * tickRate
	spawnProtect     = 2 * tickRate
	offlineTicks     = 5 * tickRate
	pvpInvuln        = 36 // 0.6 s after a hit
	pvpSwing         = 10.0
	pvpSwingCD       = 24 // 0.4 s
	pvpPulse         = 10.0
	pvpPulseCD       = 72 // 1.2 s
	pvpDashTicks     = 12 // 0.2 s
	pvpDashCD        = 2 * tickRate
	pvpSmashMin      = 15.0
	pvpSmashMax      = 30.0
	pvpReflected     = 10.0
	pvpStamp         = 10.0
	pvpHotfix        = 10.0
	pvpHotfixHeal    = 20.0 // per cast, to P4 only
	pvpHealPerTick   = 20.0 / tickRate
	pvpTimecard      = 15.0
	pvpCardDashLock  = 7 * tickRate // P5 skill locks the dash
	pvpDashCardLock  = 2 * tickRate // a dash locks the P5 skill
	hrBanner         = "Giải ngân trước khi khóa sổ!"
)

// pvp is a player's party state; the phase start resets it.
type pvp struct {
	Health    float64
	Wallet    int
	Balance   int
	Kills     int
	dead      bool
	respawn   int // ticks until respawn
	protect   int // spawn protection ticks
	attackCD  int
	offline   int  // ticks offline
	left      bool // off the field after 5 s offline
	atm       int  // index+1 of the ATM in use
	hitSet    map[int]bool
	markOn    int // P3: player carrying this player's mark
	markTicks int
	// Counters for one-shot sounds; they only grow within a phase.
	Pickups, Drops, Deposits, AtmStarts, AtmCancels, Respawns, Marks uint32
}

type Bundle struct {
	ID     int
	X, Y   float64 // bottom-centre
	vx, vy float64
	Value  int
	owner  int // previous owner of dropped coins; 0 for sprayed coins
	lock   int
}

type ATM struct {
	X, Y  float64
	user  int // player ID, 0 when free
	ticks int
}

type partyState struct {
	t         int // ticks since the phase started, countdown included
	n         int // players at the start, fixed for the phase
	bundles   []*Bundle
	atms      []*ATM
	shots     []*Shot
	zones     []*Zone
	sprayed   int
	Sprays    uint32
	banner    int // ticks left showing the HR banner
	over      bool
	overTicks int
}

func (r *Room) startParty() {
	r.mode, r.level, r.phase = modeParty, r.partyLevel, phasePlaying
	r.players = keep(r.players, func(p *Player) bool { return p.client != nil }) // away players miss the party
	if r.byID(r.hostID) == nil && len(r.players) > 0 {
		r.hostID = r.players[0].ID
	}
	r.party = partyState{n: len(r.players)}
	for _, a := range r.level.ATMs {
		r.party.atms = append(r.party.atms, &ATM{X: a.X, Y: a.Y})
	}
	for i, p := range r.players {
		s := r.level.Spawns[i%len(r.level.Spawns)]
		p.Body = Body{X: s.X, Y: s.Y, Facing: 1}
		p.fighter = fighter{}
		p.pvp = pvp{Health: pvpHP}
		p.pressed = presses{}
	}
	r.tick = 0
	r.broadcast(encode(r.lobby()))
	r.broadcast(encode(startMsg{Type: "start", Level: partyLevelURL, Mode: modeParty}))
}

func (r *Room) stepParty() {
	pt := &r.party
	if pt.over {
		if pt.overTicks--; pt.overTicks <= 0 {
			r.toLobby()
		}
		return
	}
	pt.t++
	live := pt.t > partyCountdown
	for _, p := range r.players {
		r.stepPvP(p, live)
	}
	if !live {
		return
	}
	mt := pt.t - partyCountdown // 1 on the first live tick
	if at := mt - 1; at%sprayEvery == 0 && at <= sprayLast {
		r.spray(sprayPerPlayer*pt.n, sprayBundles)
	}
	if mt == finalBurstAt {
		r.spray(finalPerPlayer*pt.n, finalBundles)
		pt.banner = 3 * tickRate
	}
	if pt.banner > 0 {
		pt.banner--
	}
	r.stepPvPShots()
	r.stepPvPZones()
	r.stepBundles()
	if mt >= partyTicks { // deposits finished on this tick already counted above
		for _, p := range r.players {
			r.cancelATM(p)
		}
		pt.over, pt.overTicks = true, partyResultTicks
	}
}

func (r *Room) stepPvP(p *Player, live bool) {
	pr := p.pressed
	p.pressed = presses{}
	if p.left {
		if p.client != nil { // back after leaving the field: respawn with an empty wallet
			r.respawnPvP(p)
		}
		return
	}
	for _, t := range []*int{&p.invuln, &p.dashCD, &p.pulseCD, &p.skillCD, &p.attackCD, &p.protect, &p.markTicks} {
		if *t > 0 {
			*t--
		}
	}
	if p.markTicks == 0 {
		p.markOn = 0
	}
	if p.client == nil {
		r.cancelATM(p)
		if p.action == actCharge {
			r.cancelChargePvP(p)
		}
		if p.offline++; p.offline >= offlineTicks {
			r.dropMoney(p, p.Wallet) // the rest of the wallet stays on the field
			r.clearEffects(p)
			p.left, p.dead = true, false
			return
		}
		if !p.dead {
			p.step(r.level, Input{}, false) // frozen, but still falls
		}
		return
	}
	p.offline = 0
	if p.dead {
		if p.respawn--; p.respawn <= 0 {
			r.respawnPvP(p)
		}
		return
	}
	if p.actionTicks > 0 {
		if p.actionTicks--; p.actionTicks == 0 {
			p.action = ""
		}
	}
	in := p.in
	if !live {
		p.step(r.level, Input{Left: in.Left, Right: in.Right, Down: in.Down}, pr.jump)
		return
	}
	if p.action == actHurt {
		p.step(r.level, Input{}, false)
		return
	}
	if p.action == actCharge {
		switch {
		case pr.dash:
			r.cancelChargePvP(p)
		case !in.Special || p.charge >= chargeMaxTicks:
			power := float64(p.charge) / chargeMaxTicks
			p.action, p.actionTicks, p.castHit = actSmash, castTicks, false
			p.chargeDmg = pvpSmashMin + (pvpSmashMax-pvpSmashMin)*power
			p.skillCD = skillCooldown[slotTeal]
			p.Casts++
		default:
			p.charge++
		}
	}
	if p.action == "" && p.atm == 0 {
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
			p.action, p.actionTicks, p.dashCD = actDash, pvpDashTicks, pvpDashCD
			p.Body.dashTicks, p.Body.dashSpeed = pvpDashTicks, dashSpeed
			if p.Slot == slotOrange {
				p.skillCD = max(p.skillCD, pvpDashCardLock)
			}
			p.Dashes++
		case pr.special && p.skillCD == 0:
			p.protect = 0 // using a skill ends spawn protection
			if dir != 0 {
				p.Facing = dir
			}
			r.startPvPSkill(p)
		case pr.attack && p.attackCD == 0:
			p.protect = 0
			p.action, p.actionTicks, p.attackCD = actAttack, attackTicks, pvpSwingCD
			p.hitSet = map[int]bool{}
			p.Attacks++
		case pr.skill && p.pulseCD == 0:
			p.protect = 0
			p.action, p.actionTicks, p.pulseCD = actPulse, attackTicks, pvpPulseCD
			p.Pulses++
			r.addPvPShot(p, "pulse", 32, 16, pulseSpeed, pvpPulse, pulseLife, true)
		}
	}

	r.stepDeposit(p, in, pr)
	facing := p.Facing
	p.slow = p.action == actCharge
	if p.atm != 0 {
		p.step(r.level, Input{}, false)
	} else {
		p.step(r.level, in, pr.jump && !isDashing(p.action))
	}
	switch p.action {
	case actAttack, actPulse, actSmash, actShield, actStamp, actHotfix:
		p.Facing = facing
	}
	if p.action == actTimecard && p.Body.dashTicks == 0 {
		p.action, p.actionTicks = "", 0
	}
	r.pvpMelee(p)
}

// pvpMelee lands swings, the P1 smash and the P5 lunge on rivals, each target at most once per action.
func (r *Room) pvpMelee(p *Player) {
	switch p.action {
	case actAttack:
		elapsed := attackTicks - p.actionTicks
		if elapsed < swingFrom || elapsed > swingTo {
			return
		}
		for _, q := range r.rivals(p) {
			if !p.hitSet[q.ID] && overlaps(reachBox(p, swingReach), playerBox(q)) && r.pvpDamage(p, q, pvpSwing, true) {
				p.hitSet[q.ID] = true
			}
		}
	case actSmash: // one rival: the nearest in front
		if p.castHit {
			return
		}
		var best *Player
		for _, q := range r.rivals(p) {
			if overlaps(reachBox(p, smashReach), playerBox(q)) && (best == nil || math.Abs(q.X-p.X) < math.Abs(best.X-p.X)) {
				best = q
			}
		}
		if best != nil && r.pvpDamage(p, best, p.chargeDmg, true) {
			p.castHit = true
			p.Hits++
			p.SkillHits++
		}
	case actTimecard:
		for _, q := range r.rivals(p) {
			if !p.hitSet[q.ID] && overlaps(playerBox(p), playerBox(q)) && r.pvpDamage(p, q, pvpTimecard, false) {
				p.hitSet[q.ID] = true
				p.SkillHits++
			}
		}
	}
}

func (r *Room) startPvPSkill(p *Player) {
	switch p.Slot {
	case slotTeal:
		p.action, p.actionTicks, p.charge = actCharge, 0, 0
		return // the cast is counted on release
	case slotViolet:
		p.action, p.actionTicks, p.reflectsLeft = actShield, shieldTicks, maxReflects
	case slotAmber:
		p.action, p.actionTicks = actStamp, castTicks
		r.addPvPShot(p, "stamp", 32, 32, stampSpeed, pvpStamp, stampLife, true)
	case slotMint:
		p.action, p.actionTicks = actHotfix, castTicks
		r.nextShotID++
		z := &Zone{ID: r.nextShotID, X: p.X, Y: p.Y, ticks: hotfixTicks, owner: p.ID, healLeft: pvpHotfixHeal}
		r.party.zones = append(r.party.zones, z)
		for _, q := range r.rivals(p) {
			if overlaps(zoneBox(z), playerBox(q)) {
				r.pvpDamage(p, q, pvpHotfix, false)
			}
		}
	case slotOrange:
		p.action, p.actionTicks = actTimecard, timecardTicks
		p.Body.dashTicks, p.Body.dashSpeed = timecardTicks, timecardSpeed
		p.dashCD = max(p.dashCD, pvpCardDashLock)
		p.hitSet = map[int]bool{}
	}
	p.skillCD = skillCooldown[p.Slot]
	p.Casts++
}

func (r *Room) cancelChargePvP(p *Player) {
	p.action, p.charge = "", 0
	p.skillCD = skillCooldown[slotTeal]
	p.Cancels++
}

// rivals are the other players on the field who can be hit: alive and not gone, online or not.
func (r *Room) rivals(p *Player) []*Player {
	var out []*Player
	for _, q := range r.players {
		if q != p && !q.dead && !q.left {
			out = append(out, q)
		}
	}
	return out
}

// pvpDamage applies a hit and reports whether it landed. markable hits (keyboard, P1 smash, code
// pulse) get +15% when the attacker is the P3 whose mark the target carries.
func (r *Room) pvpDamage(by, q *Player, dmg float64, markable bool) bool {
	if q.dead || q.left || q.protect > 0 || q.invuln > 0 || q.action == actDash ||
		(q.action == actTimecard && q.actionTicks > timecardTicks-timecardInvuln) {
		return false
	}
	if markable && by.Slot == slotAmber && by.markOn == q.ID && by.markTicks > 0 {
		dmg *= markBonus
	}
	before := q.Wallet
	q.Health -= dmg
	q.invuln = pvpInvuln
	q.Hurts++
	r.cancelATM(q)
	if q.action == actCharge {
		r.cancelChargePvP(q)
	}
	q.Body.dashTicks = 0
	if q.Health <= 0 {
		q.Health = 0
		r.dropMoney(q, (before+1)/2) // ceil 50%, nothing more
		q.dead, q.respawn = true, pvpRespawnTicks
		q.action, q.actionTicks = actDown, 0
		by.Kills++
		r.clearEffects(q)
		return true
	}
	r.dropMoney(q, (before+9)/10) // ceil 10%
	dir := 1.0
	if q.X < by.X {
		dir = -1
	}
	q.action, q.actionTicks = actHurt, hurtTicks
	q.pushVX, q.pushTicks, q.VY = dir*knockSpeed, hurtTicks, -knockHop
	return true
}

// clearEffects removes what a fallen or departed player created: zones, shots, marks. It only marks
// shots and zones as expired, since it can run while those lists are being stepped.
func (r *Room) clearEffects(p *Player) {
	for _, s := range r.party.shots {
		if s.owner == p.ID {
			s.life = 0
		}
	}
	for _, z := range r.party.zones {
		if z.owner == p.ID {
			z.ticks = 0
		}
	}
	p.markOn, p.markTicks = 0, 0
	for _, q := range r.players {
		if q.markOn == p.ID {
			q.markOn, q.markTicks = 0, 0
		}
	}
	r.cancelATM(p)
}

func (r *Room) respawnPvP(p *Player) {
	best, bestD := r.level.Respawns[0], -1.0
	for _, s := range r.level.Respawns {
		d := math.Inf(1)
		for _, q := range r.rivals(p) {
			d = math.Min(d, math.Hypot(q.X-s.X, q.Y-s.Y))
		}
		if d > bestD {
			best, bestD = s, d
		}
	}
	if p.left {
		p.Wallet = 0 // the wallet was dropped when leaving
	}
	p.Body = Body{X: best.X, Y: best.Y, Facing: p.Facing}
	p.Health, p.dead, p.left, p.offline = pvpHP, false, false, 0
	p.action, p.actionTicks, p.charge = "", 0, 0
	p.protect, p.invuln = spawnProtect, 0
	p.Respawns++
}

// dropMoney moves coins from a wallet to a bundle at the player, locked for its previous owner.
func (r *Room) dropMoney(p *Player, amount int) {
	amount = min(amount, p.Wallet)
	if amount <= 0 {
		return
	}
	p.Wallet -= amount
	p.Drops++
	r.nextShotID++
	r.party.bundles = append(r.party.bundles, &Bundle{
		ID: r.nextShotID, X: p.X, Y: p.Y - hitH/2, vx: (rand.Float64() - 0.5) * 160, vy: -260,
		Value: amount, owner: p.ID, lock: bundleLock,
	})
}

func (r *Room) spray(total, n int) {
	s := r.level.Safe
	for range n {
		r.nextShotID++
		r.party.bundles = append(r.party.bundles, &Bundle{
			ID: r.nextShotID, X: s.X, Y: s.Y - 40,
			vx: (rand.Float64()*2 - 1) * 320, vy: -350 - rand.Float64()*300, Value: total / n,
		})
	}
	r.party.sprayed += total
	r.party.Sprays++
}

func (r *Room) stepBundles() {
	ts := float64(r.level.Tile)
	r.party.bundles = keep(r.party.bundles, func(b *Bundle) bool {
		if b.lock > 0 {
			b.lock--
		}
		if b.vx != 0 || b.vy != 0 { // flying: bounce off walls, land on any surface
			b.vy = min(b.vy+gravity*dt, maxFall)
			if x := b.X + b.vx*dt; r.level.at(cell(x, ts), cell(b.Y-1, ts)) == tileSolid {
				b.vx = -b.vx * 0.4
			} else {
				b.X = x
			}
			prev := b.Y
			b.Y += b.vy * dt
			if b.vy > 0 {
				row := cell(b.Y-eps, ts)
				top := float64(row) * ts
				if k := r.level.at(cell(b.X, ts), row); k != tileEmpty && prev <= top+eps {
					b.Y, b.vx, b.vy = top, 0, 0
				}
			}
		}
		box := rect{b.X - bundleSize/2, b.Y - bundleSize, b.X + bundleSize/2, b.Y}
		for _, p := range r.players {
			if p.client == nil || p.dead || p.left || p.protect > 0 || p.atm != 0 || (b.owner == p.ID && b.lock > 0) {
				continue
			}
			if overlaps(box, playerBox(p)) {
				p.Wallet += b.Value
				p.Pickups++
				return false
			}
		}
		return true
	})
}

// stepDeposit starts, keeps or cancels a deposit. Depositing means holding interact, standing still
// next to a free ATM for 3 s; anything else cancels it without losing coins.
func (r *Room) stepDeposit(p *Player, in Input, pr presses) {
	if p.atm != 0 {
		if !in.Interact || in.Left || in.Right || pr.jump || pr.attack || pr.skill || pr.special || pr.dash || p.action != "" {
			r.cancelATM(p)
			return
		}
		a := r.party.atms[p.atm-1]
		if a.ticks++; a.ticks >= atmTicks {
			p.Balance += p.Wallet
			p.Wallet = 0
			p.Deposits++
			a.user, a.ticks, p.atm = 0, 0, 0
		}
		return
	}
	if !in.Interact || p.Wallet == 0 || p.protect > 0 || p.action != "" || !p.OnGround || in.Left || in.Right {
		return
	}
	for i, a := range r.party.atms {
		if a.user == 0 && math.Abs(p.X-a.X) <= atmReach && math.Abs(p.Y-a.Y) < 1 {
			a.user, a.ticks, p.atm = p.ID, 0, i+1
			p.AtmStarts++
			return
		}
	}
}

func (r *Room) cancelATM(p *Player) {
	if p.atm == 0 {
		return
	}
	a := r.party.atms[p.atm-1]
	a.user, a.ticks, p.atm = 0, 0, 0
	p.AtmCancels++
}

func (r *Room) addPvPShot(p *Player, kind string, w, h, speed, dmg float64, life int, reflectable bool) {
	r.nextShotID++
	r.party.shots = append(r.party.shots, &Shot{
		ID: r.nextShotID, Kind: kind, X: p.X + float64(p.Facing)*(hitW/2+w/2), Y: p.Y - hitH/2,
		VX: float64(p.Facing) * speed, W: w, H: h, Dmg: dmg, life: life, owner: p.ID, reflectable: reflectable,
	})
}

func (r *Room) stepPvPShots() {
	var bounced []*Shot // appended after the loop: keep reuses the slice it walks
	r.party.shots = keep(r.party.shots, func(s *Shot) bool {
		if s.life--; s.life <= 0 || !r.moveShot(s) {
			return false
		}
		owner := r.byID(s.owner)
		if owner == nil || owner.dead || owner.left {
			return false
		}
		for _, q := range r.rivals(owner) {
			if s.reflectable && q.action == actShield && q.reflectsLeft > 0 && s.VX*float64(q.Facing) < 0 &&
				overlaps(shotBox(s), shieldBox(q)) {
				q.reflectsLeft--
				q.Reflects++
				dir := 1.0 // back toward the shooter, at the moment of the reflection
				if owner.X < s.X {
					dir = -1
				}
				r.nextShotID++
				bounced = append(bounced, &Shot{
					ID: r.nextShotID, Kind: "returned", X: s.X, Y: s.Y, VX: dir * cardSpeed,
					W: cardW, H: cardH, Dmg: pvpReflected, life: 2 * tickRate, owner: q.ID,
				})
				return false
			}
		}
		for _, q := range r.rivals(owner) {
			if overlaps(shotBox(s), playerBox(q)) && r.pvpDamage(owner, q, s.Dmg, s.Kind == "pulse") {
				if s.Kind == "stamp" { // one mark per P3: the newest target replaces the old one
					owner.markOn, owner.markTicks = q.ID, markTicks
					owner.Marks++
				}
				return false
			}
		}
		return true
	})
	r.party.shots = append(r.party.shots, bounced...)
}

func (r *Room) stepPvPZones() {
	r.party.zones = keep(r.party.zones, func(z *Zone) bool {
		if z.ticks <= 0 { // cleared when its P4 fell
			return false
		}
		if p := r.byID(z.owner); p != nil && !p.dead && z.healLeft > 0 && p.Health < pvpHP && overlaps(zoneBox(z), playerBox(p)) {
			h := math.Min(math.Min(pvpHealPerTick, z.healLeft), pvpHP-p.Health)
			if z.healLeft == pvpHotfixHeal {
				p.Healed++
			}
			p.Health += h
			z.healLeft -= h
		}
		z.ticks--
		return z.ticks > 0
	})
}

// results ranks deposited balances; ties share a rank, all-zero included.
func (r *Room) results() []resultState {
	var out []resultState
	for _, p := range r.players {
		out = append(out, resultState{ID: p.ID, Balance: p.Balance, Kills: p.Kills})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Balance > out[j].Balance })
	for i := range out {
		out[i].Rank = 1
		for _, o := range out {
			if o.Balance > out[i].Balance {
				out[i].Rank++
			}
		}
	}
	return out
}

func (r *Room) partySnapshot() snapshotMsg {
	pt := &r.party
	ms := func(ticks int) int64 { return int64(ticks) * 1000 / tickRate }
	m := snapshotMsg{
		Type: "snapshot", Tick: r.tick, Mode: modeParty,
		Players: []playerState{}, Shots: []shotState{}, Waves: []waveState{}, Items: []itemState{},
		Zones: []zoneState{}, Drops: []dropState{},
		Party: &partyMsg{
			CountdownMs: ms(max(0, partyCountdown-pt.t)),
			TimeLeftMs:  ms(max(0, partyTicks-max(0, pt.t-partyCountdown))),
			Solo:        pt.n == 1,
			SafeX:       r.level.Safe.X, SafeY: r.level.Safe.Y,
			Sprays:  pt.Sprays,
			Bundles: []bundleState{}, ATMs: []atmState{},
		},
	}
	m.ServerTime = serverNow()
	if pt.banner > 0 {
		m.Party.Banner = hrBanner
	}
	if pt.over {
		m.Party.Results = r.results()
	}
	for _, p := range r.players {
		ps := playerState{
			ID: p.ID, X: p.X, Y: p.Y, VX: p.VX, VY: p.VY, Facing: p.Facing, Ground: p.OnGround, Away: p.client == nil,
			Action: p.action, Charge: float64(p.charge) / chargeMaxTicks, Ping: p.ping,
			SkillMs: ms(p.skillCD), SkillOf: ms(skillCooldown[p.Slot]),
			Attacks: p.Attacks, Pulses: p.Pulses, Dashes: p.Dashes, Hurts: p.Hurts, Hits: p.Hits,
			Casts: p.Casts, SkillHits: p.SkillHits, Cancels: p.Cancels, Reflects: p.Reflects, Healed: p.Healed,
			PvP: &pvpMsg{
				HP: p.Health, Wallet: p.Wallet, Balance: p.Balance, Kills: p.Kills, Dead: p.dead, Left: p.left,
				RespawnMs: ms(p.respawn), ProtectMs: ms(p.protect), MarkedBy: 0,
				Pickups: p.Pickups, Drops: p.Drops, Deposits: p.Deposits, AtmStarts: p.AtmStarts,
				AtmCancels: p.AtmCancels, Respawns: p.Respawns, Marks: p.Marks,
			},
		}
		if p.dead {
			ps.Action = actDown
		}
		m.Players = append(m.Players, ps)
	}
	for i, ps := range m.Players { // who carries whose mark, for the stamp icon over the target
		for _, q := range r.players {
			if q.markOn == ps.ID && q.markTicks > 0 {
				m.Players[i].PvP.MarkedBy = q.ID
			}
		}
	}
	for _, b := range pt.bundles {
		m.Party.Bundles = append(m.Party.Bundles, bundleState{ID: b.ID, X: b.X, Y: b.Y, Value: b.Value})
	}
	for _, a := range pt.atms {
		as := atmState{X: a.X, Y: a.Y, User: a.user, Ms: ms(a.ticks), Of: ms(atmTicks)}
		if p := r.byID(a.user); p != nil {
			as.Amount = p.Wallet
		}
		m.Party.ATMs = append(m.Party.ATMs, as)
	}
	for _, s := range pt.shots {
		m.Shots = append(m.Shots, shotState{ID: s.ID, Kind: s.Kind, X: s.X, Y: s.Y, VX: s.VX})
	}
	for _, z := range pt.zones {
		m.Zones = append(m.Zones, zoneState{ID: z.ID, X: z.X, Y: z.Y, Ms: ms(z.ticks)})
	}
	return m
}
