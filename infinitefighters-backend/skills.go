package main

// The five slot skills ("Năm chiêu riêng", dev/game-scenario.md). One button, its effect depends on
// the slot. Shared limits: a missed skill still uses its cooldown; skills never hurt or push
// teammates, never pause the DEADLINE, cancel a boss telegraph or stun the boss.

const (
	slotTeal = iota
	slotViolet
	slotAmber
	slotMint
	slotOrange
)

const (
	actCharge   = "charge"   // P1 holding
	actSmash    = "smash"    // P1 release
	actShield   = "shield"   // P2
	actStamp    = "stamp"    // P3 throw
	actHotfix   = "hotfix"   // P4 deploy
	actTimecard = "timecard" // P5 lunge
)

// Tuning (starting values from the design doc).
const (
	chargeMaxTicks = tickRate // P1 auto-releases at full charge
	chargeSlow     = 0.4      // walking speed while charging
	smashMin       = 3.0      // damage for an instant release
	smashMax       = 10.0     // damage at full charge
	smashReach     = 48.0

	shieldTicks    = 36 // 0.6 s
	maxReflects    = 3
	returnedDamage = 3.0

	stampSpeed  = 500.0
	stampLife   = 36 // ≈ 300 px
	stampDamage = 2.0
	markTicks   = 4 * tickRate
	markBonus   = 1.15

	hotfixTicks  = 3 * tickRate
	hotfixW      = 128.0
	hotfixH      = 48.0
	hotfixDamage = 3.0

	timecardTicks  = 12 // 3 tiles at timecardSpeed
	timecardSpeed  = 480.0
	timecardInvuln = 9 // 0.15 s
	timecardDamage = 4.0
)

// skillCooldown is per slot, in ticks, counted from the cast (P1: from release or cancel).
var skillCooldown = [5]int{6 * tickRate, 8 * tickRate, 10 * tickRate, 14 * tickRate, 7 * tickRate}

// Zone is a P4 hotfix area at someone's feet.
type Zone struct {
	ID     int
	X, Y   float64 // bottom-centre
	ticks  int
	healed map[int]bool // at most one heart per player per cast
}

func (r *Room) startSkill(p *Player) {
	switch p.Slot {
	case slotTeal:
		p.action, p.actionTicks, p.charge = actCharge, 0, 0
		return // the cast is counted on release
	case slotViolet:
		p.action, p.actionTicks, p.reflectsLeft = actShield, shieldTicks, maxReflects
	case slotAmber:
		p.action, p.actionTicks = actStamp, castTicks
		r.addShot(p, "stamp", 32, 32, stampSpeed, stampDamage, stampLife)
	case slotMint:
		p.action, p.actionTicks = actHotfix, castTicks
		r.nextShotID++
		z := &Zone{ID: r.nextShotID, X: p.X, Y: p.Y, ticks: hotfixTicks, healed: map[int]bool{}}
		r.zones = append(r.zones, z)
		if r.boss.State != bossDefeat && overlaps(zoneBox(z), r.bossBox()) {
			p.SkillHits++
			r.damageBoss(hotfixDamage, false)
		}
	case slotOrange:
		p.action, p.actionTicks, p.castHit = actTimecard, timecardTicks, false
		p.Body.dashTicks, p.Body.dashSpeed = timecardTicks, timecardSpeed
		p.dashCD = skillCooldown[slotOrange] // one shared cooldown with the normal dash
	}
	p.skillCD = skillCooldown[p.Slot]
	p.Casts++
}

// stepCharge runs before the normal action handling while P1 holds the skill.
func (r *Room) stepCharge(p *Player, in Input, pr presses) {
	switch {
	case pr.dash:
		r.cancelCharge(p) // the dash below then starts as usual
	case !in.Special || p.charge >= chargeMaxTicks:
		power := float64(p.charge) / chargeMaxTicks
		p.action, p.actionTicks, p.castHit = actSmash, castTicks, false
		p.chargeDmg = smashMin + (smashMax-smashMin)*power
		p.skillCD = skillCooldown[slotTeal]
		p.Casts++
	default:
		p.charge++
	}
}

func (r *Room) cancelCharge(p *Player) {
	p.action, p.charge = "", 0
	p.skillCD = skillCooldown[slotTeal]
	p.Cancels++
}

// skillHits checks the melee parts of skills against the boss, once per cast.
func (r *Room) skillHits(p *Player) {
	switch {
	case p.action == actSmash && !p.castHit && overlaps(reachBox(p, smashReach), r.bossBox()):
		p.castHit = true
		p.Hits++
		p.SkillHits++
		r.damageBoss(p.chargeDmg, true)
	case p.action == actTimecard && !p.castHit && overlaps(playerBox(p), r.bossBox()):
		p.castHit = true
		p.SkillHits++
		r.damageBoss(timecardDamage, false)
	}
}

func (r *Room) endTimecard(p *Player) {
	p.action, p.actionTicks = "", 0
}

// reflect bounces a task card off a P2 shield back toward the boss. It reports whether it did.
func (r *Room) reflect(p *Player, s *Shot) bool {
	if p.action != actShield || p.reflectsLeft == 0 || p.client == nil ||
		s.VX*float64(p.Facing) >= 0 || !overlaps(shotBox(s), shieldBox(p)) {
		return false
	}
	p.reflectsLeft--
	p.Reflects++
	dir := 1.0
	if r.boss.X < s.X {
		dir = -1
	}
	r.nextShotID++
	r.shots = append(r.shots, &Shot{
		ID: r.nextShotID, Kind: "returned", X: s.X, Y: s.Y, VX: dir * cardSpeed,
		W: cardW, H: cardH, Dmg: returnedDamage, life: 4 * tickRate,
	})
	return true
}

func shieldBox(p *Player) rect {
	near, far := p.X+float64(p.Facing)*hitW/2, p.X+float64(p.Facing)*(hitW/2+24)
	return rect{min(near, far), p.Y - 60, max(near, far), p.Y}
}

func zoneBox(z *Zone) rect { return rect{z.X - hotfixW/2, z.Y - hotfixH, z.X + hotfixW/2, z.Y} }

func (r *Room) stepZones() {
	r.zones = keep(r.zones, func(z *Zone) bool {
		for _, p := range r.players {
			if !z.healed[p.ID] && p.client != nil && p.action != actDown && p.HP < playerHP &&
				overlaps(zoneBox(z), playerBox(p)) {
				z.healed[p.ID] = true
				p.HP++
				p.Healed++
			}
		}
		z.ticks--
		return z.ticks > 0
	})
}
