package main

import "math/rand/v2"

// Khabib sandbox (docs/levels/sandbox-khabib.md): the level 1 arena with the boss swapped for
// Khabib, who loops three attacks. No DEADLINE and no loss: downed players come back by
// themselves and the boss refills when beaten, so skills and timings can be tried freely.
// Concept art: assets/concepts/boss-khabib-v1.png.

const (
	levelSandbox       = "sandbox" // client "start" level names; the arena is the level 1 map
	levelSandboxVought = "sandbox-vought"
	sandboxKhabib      = "khabib"
	sandboxVought      = "vought"
	levelSandboxYasuka = "sandbox-yasuka"
	sandboxYasuka      = "yasuka"

	khIdleTicks      = tickRate / 2
	khPressTicks     = 150 // 2.5 s advancing
	khPressSpeed     = 0.6 * runSpeed
	khPressWaveEvery = 40
	khSpamWindup     = 30
	khSpamTicks      = 150
	khSpamEvery      = 15
	khNagSpeed       = 0.8 * cardSpeed
	khNagW, khNagH   = 56.0, 16.0
	khTeleOutTicks   = 30 // dissolving, still where he stood
	khTeleWarnTicks  = 36 // unseen, landing spot marked on the floor
	khSlamTicks      = 12 // landing slam: his body hurts
	khRecoverTicks   = 60 // vulnerable, damage ×2
	khChainOutTicks  = 12 // teleport chain: every hop is faster
	khChainWarnTicks = 24

	sandboxRespawnTicks = 2 * tickRate
)

const (
	khPressWindup = "press_windup"
	khPress       = "press"
	khSpamWind    = "spam_windup"
	khSpam        = "spam"
	khTeleOut     = "teleport_out"
	khTeleWarn    = "teleport_warn"
	khTeleIn      = "teleport_in"
	khRecover     = "recover"
	khChainWind   = "chain_windup"
)

var khCycle = []string{khPressWindup, khSpamWind, khTeleOut, khPressWindup, khSpamWind, khChainWind}

var khWords = []string{"XONG CHUA?", "BAO CAO!", "UPDATE!"}

func (r *Room) stepKhabib() {
	b := &r.boss
	set := func(s string) { b.State, b.ticks = s, 0 }
	r.settleBoss()
	switch b.State {
	case bossIdle:
		r.faceNearest()
		if b.ticks >= r.t(khIdleTicks) && len(r.cards) == 0 && len(r.waves) == 0 {
			set(khCycle[b.cycle%len(khCycle)])
			b.cycle++
		}
	case khPressWindup:
		r.faceNearest()
		if b.ticks >= r.t(windupTicks) {
			b.chargeHit = map[int]bool{}
			set(khPress)
		}
	case khPress: // pressing: walks at the nearest player, pressure waves along the floor
		r.faceNearest()
		b.X = max(r.bossMinX, min(r.bossMaxX, b.X+float64(b.Facing)*khPressSpeed*r.moveMul()*dt)) // walks off a desk edge like a player
		if b.ticks%r.t(khPressWaveEvery) == 1 {
			r.pressWave()
		}
		r.contactHurt()
		if b.ticks >= r.t(khPressTicks) {
			set(khRecover)
		}
	case khSpamWind:
		r.faceNearest()
		if b.ticks >= r.t(khSpamWindup) {
			set(khSpam)
		}
	case khSpam: // talk spam: one word per lane, any lane from the floor up to the highest desk
		r.faceNearest()
		if b.ticks%r.t(khSpamEvery) == 1 {
			r.fireNag()
		}
		if b.ticks >= r.t(khSpamTicks) {
			set(bossIdle)
		}
	case khChainWind: // teleport chain: a few quick hops, one player each (random spots when alone)
		r.faceNearest()
		if b.ticks >= r.t(windupTicks) {
			r.startChain()
			set(khTeleOut)
		}
	case khTeleOut:
		out, _ := r.teleTimes()
		if b.ticks >= r.t(out) {
			b.landX, b.landY = r.nextSpot()
			set(khTeleWarn)
		}
	case khTeleWarn:
		if _, warn := r.teleTimes(); b.ticks >= r.t(warn) {
			b.X, b.Y, b.landX, b.landY = b.landX, b.landY, 0, 0
			if b.chain = max(0, b.chain-1); b.chain == 0 {
				b.queue = b.queue[:0] // a chain cut short must not leak into the next teleport
			}
			r.faceNearest()
			b.chargeHit = map[int]bool{}
			set(khTeleIn)
		}
	case khTeleIn:
		r.contactHurt()
		if b.ticks >= khSlamTicks {
			if b.chain > 0 {
				set(khTeleOut)
			} else {
				set(khRecover)
			}
		}
	case khRecover:
		if b.ticks >= r.t(khRecoverTicks) {
			set(bossIdle)
		}
	}
}

// pressWave is the ground shock: always along the floor, even from a desk, plus one along the desk he stands on.
func (r *Room) pressWave() {
	b := &r.boss
	floor := r.level.Boss.Y
	rows := []float64{floor}
	if b.Y < floor {
		rows = append(rows, b.Y)
	}
	for _, y := range rows {
		r.nextShotID++
		r.waves = append(r.waves, &Wave{ID: r.nextShotID, X: b.X + float64(b.Facing)*(b.W/2+waveW/2), Y: y, Dir: b.Facing, hit: map[int]bool{}})
	}
}

// teleTimes are the dissolve and warning lengths: quicker while a chain still has hops to land.
func (r *Room) teleTimes() (out, warn int) {
	if r.boss.chain > 0 {
		return khChainOutTicks, khChainWarnTicks
	}
	return khTeleOutTicks, khTeleWarnTicks
}

// startChain picks the hops. Two or more players: one hop behind each of them, in random order.
// Alone: two to four hops to random spots.
func (r *Room) startChain() {
	b := &r.boss
	b.queue = b.queue[:0]
	for _, p := range r.players {
		if p.client != nil && p.action != actDown {
			b.queue = append(b.queue, p.ID)
		}
	}
	rand.Shuffle(len(b.queue), func(i, j int) { b.queue[i], b.queue[j] = b.queue[j], b.queue[i] })
	if len(b.queue) >= 2 {
		b.chain = len(b.queue)
		return
	}
	b.queue, b.chain = b.queue[:0], 2+rand.IntN(3)
}

// nextSpot is where the next hop lands: behind the next player in the chain, a random spot when the
// chain has none left, or behind the nearest player for the plain teleport.
func (r *Room) nextSpot() (x, y float64) {
	b := &r.boss
	for len(b.queue) > 0 {
		id := b.queue[0]
		b.queue = b.queue[1:]
		if p := r.byID(id); p != nil && p.client != nil && p.action != actDown {
			return r.behind(p)
		}
	}
	if b.chain > 0 {
		return r.randomSpot()
	}
	if t := r.nearestTarget(); t != nil {
		return r.behind(t)
	}
	return b.X, b.Y
}

// behind is behind the player's back, close enough that standing still gets slammed,
// on whatever the player is above: the floor or a desk.
func (r *Room) behind(t *Player) (x, y float64) {
	return max(r.bossMinX, min(r.bossMaxX, t.X-float64(t.Facing)*r.boss.W/2)), r.groundBelow(t.X, t.Y)
}

// randomSpot is any standing place in the arena: the floor or a desk.
func (r *Room) randomSpot() (x, y float64) {
	floor, ts := r.level.Boss.Y, float64(r.level.Tile)
	tiers := append([]float64{floor}, r.level.PlatformTops...)
	for range 20 {
		y = tiers[rand.IntN(len(tiers))]
		x = r.bossMinX + rand.Float64()*(r.bossMaxX-r.bossMinX)
		if y == floor || r.level.at(cell(x, ts), cell(y, ts)) == tileOneWay {
			return x, y
		}
	}
	return (r.bossMinX + r.bossMaxX) / 2, floor
}

// groundBelow is the top of the first tile at or under feet height y: the desk or floor under a jumping player.
func (r *Room) groundBelow(x, y float64) float64 {
	ts := float64(r.level.Tile)
	row := cell(y, ts)
	for r.level.at(cell(x, ts), row) == tileEmpty {
		row++
	}
	return float64(row) * ts
}

// settleBoss makes him fall like a player when nothing is under his centre (a desk edge, or after
// a teleport above ground).
func (r *Room) settleBoss() {
	b := &r.boss
	ground := r.groundBelow(b.X, b.Y)
	if b.Y >= ground {
		b.VY = 0
		return
	}
	b.VY = min(b.VY+gravity*dt, maxFall)
	if b.Y += b.VY * dt; b.Y >= ground {
		b.Y, b.VY = ground, 0
	}
}

// fireNag throws one word along a random lane: the floor or the top of a desk row.
func (r *Room) fireNag() {
	b := &r.boss
	lanes := append([]float64{r.level.Boss.Y}, r.level.PlatformTops...)
	r.nextShotID++
	r.cards = append(r.cards, &Shot{
		ID: r.nextShotID, Kind: "card", Word: khWords[b.nags%len(khWords)],
		X: b.X + float64(b.Facing)*(b.W/2+khNagW/2), Y: lanes[rand.IntN(len(lanes))] - hitH/2,
		VX: float64(b.Facing) * khNagSpeed * r.attackMul(), W: khNagW, H: khNagH,
	})
	b.nags++
}

// contactHurt hurts each player the moving boss body touches, once per attack (chargeHit).
func (r *Room) contactHurt() {
	b := &r.boss
	for _, p := range r.players {
		if !b.chargeHit[p.ID] && r.canHurt(p) && overlaps(playerBox(p), r.bossBox()) {
			b.chargeHit[p.ID] = true
			r.hurt(p, 1, b.X)
		}
	}
}

// stepSandbox replaces the win / lose rules: the boss refills when beaten, downed players get up.
func (r *Room) stepSandbox() {
	if r.boss.HP <= 0 {
		r.boss.HP, r.boss.State, r.boss.ticks, r.boss.landX, r.boss.chain = r.boss.MaxHP, bossIdle, 0, 0, 0
		r.cards, r.shots, r.waves, r.stamps, r.seals, r.papers = nil, nil, nil, nil, nil, nil
		r.boss.dome = 0
		r.boss.patTarget = nil
		for _, p := range r.players {
			if p.action == actPat {
				p.action, p.actionTicks = "", 0
			}
		}
	}
	for _, p := range r.players {
		if p.action != actDown || p.client == nil {
			p.downTicks = 0
			continue
		}
		if p.downTicks++; p.downTicks >= sandboxRespawnTicks {
			p.HP, p.downTicks = playerHP, 0
			p.action, p.actionTicks, p.invuln = actRevive, reviveTicks, invulnTicks
		}
	}
}
