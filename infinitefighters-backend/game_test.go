package main

import (
	"math"
	"testing"
)

// testRoom starts a match on the real level with n connected players and `away` disconnected ones.
func testRoom(t *testing.T, n, away int) *Room {
	t.Helper()
	data, _ := levelFS.ReadFile("levels/sprint.json")
	level, err := parseLevel(data)
	if err != nil {
		t.Fatal(err)
	}
	r := newRoom("TEST", level)
	for i := range n + away {
		p := &Player{ID: i + 1, Slot: i}
		if i < n {
			p.client = &client{send: make(chan []byte, 1<<12), cancel: func() {}, player: p}
		}
		r.players = append(r.players, p)
	}
	r.hostID = 1
	r.start()
	return r
}

// passive is not a boss state: stepBoss does nothing, for tests that are not about the boss.
const passive = "passive"

func steps(r *Room, n int) {
	for range n {
		r.stepMatch()
	}
}

// settle lets everyone land on the floor.
func settle(r *Room) { steps(r, 30) }

func TestBossHPScalesWithAllSlots(t *testing.T) {
	r := testRoom(t, 2, 1) // away players count too
	if want := float64(bossBaseHP + 2*bossHPPerPlayer); r.boss.MaxHP != want {
		t.Fatalf("boss HP %v, want %v", r.boss.MaxHP, want)
	}
	r.players[0].action = actDown
	r.players[1].action = actDown
	steps(r, lostTicks+2) // lost, then restart
	if r.result != "" || r.boss.MaxHP != float64(bossBaseHP+2*bossHPPerPlayer) {
		t.Fatalf("after restart: result %q, boss HP %v", r.result, r.boss.MaxHP)
	}
}

func TestSwingHitsBossOncePerSwing(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	p := r.players[0]
	p.X, p.Y, p.Facing = r.boss.X-bossW/2-hitW/2-10, r.boss.Y, 1
	hp := r.boss.HP
	p.pressed.attack = true
	steps(r, attackTicks+2)
	if r.boss.HP != hp-1 {
		t.Fatalf("boss HP %v, want %v after one swing", r.boss.HP, hp-1)
	}

	p.Facing = -1 // facing away
	p.pressed.attack = true
	r.stepMatch()
	p.Facing = -1
	steps(r, attackTicks+2)
	if r.boss.HP != hp-1 {
		t.Fatalf("swing facing away hit the boss: HP %v", r.boss.HP)
	}

	p.X, p.Facing = r.boss.X, 1 // standing inside the boss is allowed and still hits
	p.pressed.attack = true
	steps(r, attackTicks+2)
	if r.boss.HP != hp-2 || p.HP != playerHP {
		t.Fatalf("inside the boss: boss HP %v (want %v), player HP %v", r.boss.HP, hp-2, p.HP)
	}
}

func TestVolleyMissesFloor(t *testing.T) {
	r := testRoom(t, 5, 0) // 5 players: both sides, two salvos
	settle(r)
	for i, p := range r.players { // spread across the floor on both sides of the boss
		p.X = []float64{100, 400, 700, 950, 1200}[i]
	}
	r.boss.State, r.boss.ticks = bossVolley, 0
	fired := 0
	for range 10 * tickRate {
		r.stepMatch()
		fired = max(fired, len(r.cards))
		if r.boss.State != bossVolley && len(r.cards) == 0 {
			break
		}
		for _, p := range r.players {
			if p.HP != playerHP {
				t.Fatalf("player at x=%v standing on the floor was hit by the volley", p.X)
			}
			p.in = Input{} // stand still
		}
		if r.boss.State == bossIdle {
			r.boss.ticks = 0 // no next attack
		}
	}
	if want := 2 * 2 * len(r.level.PlatformTops); fired != want {
		t.Fatalf("max %d cards in flight, want %d (2 sides × 2 salvos × platform rows)", fired, want)
	}
	if r.boss.Hits != 0 || len(r.cards) != 0 {
		t.Fatalf("volley did not finish cleanly")
	}
}

func TestWaveMissesLowestPlatform(t *testing.T) {
	r := testRoom(t, 4, 0)
	settle(r)
	low := r.level.PlatformTops[len(r.level.PlatformTops)-1]
	xs := []float64{300, 950, 300, 950} // over the two lowest desks in sprint.json
	for i, p := range r.players {
		p.X, p.Y, p.VY = xs[i], low, 0
	}
	steps(r, 2)
	r.boss.State, r.boss.ticks = bossWave, 0
	for range 10 * tickRate {
		r.stepMatch()
		for _, p := range r.players {
			if p.HP != playerHP {
				t.Fatalf("player on platform at x=%v (y=%v) was hit by a wave", p.X, p.Y)
			}
		}
		if r.boss.State == bossIdle {
			r.boss.ticks = 0
			if len(r.waves) == 0 {
				return
			}
		}
	}
	t.Fatal("waves never cleared")
}

func TestWaveHitsFloorAndSkipsAwayAndDown(t *testing.T) {
	r := testRoom(t, 2, 1)
	settle(r)
	r.players[1].action = actDown
	r.boss.State, r.boss.ticks = bossWave, 0
	for range 4 * tickRate {
		r.stepMatch()
	}
	if r.players[0].HP != playerHP-1 {
		t.Fatalf("standing player HP %d, want one hit", r.players[0].HP)
	}
	if r.players[1].Hurts != 0 || r.players[2].Hurts != 0 {
		t.Fatal("downed or away player was hit")
	}
}

func TestAllConnectedDownLosesAndResets(t *testing.T) {
	r := testRoom(t, 2, 1)
	settle(r)
	r.players[0].X += 200
	r.boss.HP -= 10
	for _, p := range r.players[:2] {
		r.hurt(p, playerHP, 0)
	}
	r.stepMatch()
	if r.result != resultLost {
		t.Fatalf("result %q, want lost", r.result)
	}
	steps(r, lostTicks)
	if r.result != "" || r.boss.HP != r.boss.MaxHP || r.deadline != r.deadlineTicks {
		t.Fatalf("not reset: result %q boss %v/%v deadline %v", r.result, r.boss.HP, r.boss.MaxHP, r.deadline)
	}
	for i, p := range r.players {
		if p.HP != playerHP || p.action != "" || p.X != r.level.Spawns[i].X {
			t.Fatalf("player %d not reset: %+v", i, p.fighter)
		}
	}
}

func TestDeadlineLoses(t *testing.T) {
	r := testRoom(t, 1, 0)
	r.deadline = 3
	r.boss.State, r.boss.cycle = bossWind, 0 // keep the boss passive
	steps(r, 3)
	if r.result != resultLost {
		t.Fatalf("result %q, want lost at deadline", r.result)
	}
}

func TestHeldInteractRevives(t *testing.T) {
	r := testRoom(t, 2, 0)
	settle(r)
	r.boss.State = bossWind // passive
	d, h := r.players[0], r.players[1]
	r.hurt(d, playerHP, 0)
	h.X = d.X + 20
	h.in.Interact = true
	steps(r, reviveHoldTicks-1)
	if d.action != actDown {
		t.Fatal("revived too early")
	}
	r.stepMatch()
	if d.action != actRevive || d.HP != reviveHP {
		t.Fatalf("after hold: action %q HP %d", d.action, d.HP)
	}
}

func TestWinReturnsToLobbyAndPurgesAway(t *testing.T) {
	r := testRoom(t, 2, 1)
	r.boss.HP = 1
	r.damageBoss(1, false)
	r.stepMatch()
	if r.result != resultWon || r.boss.State != bossDefeat {
		t.Fatalf("result %q state %q", r.result, r.boss.State)
	}
	steps(r, wonTicks)
	if r.phase != phaseLobby || len(r.players) != 2 {
		t.Fatalf("phase %q with %d players, want lobby with 2", r.phase, len(r.players))
	}
	r.start()
	if r.phase != phasePlaying || r.boss.HP != float64(bossBaseHP+bossHPPerPlayer) {
		t.Fatalf("second start: phase %q boss HP %v", r.phase, r.boss.HP)
	}
}

func TestCoffeeHealsHurtPlayerAndRespawns(t *testing.T) {
	r := testRoom(t, 2, 0)
	r.boss.State = passive
	c := r.level.Coffees[0]
	full, hurt := r.players[0], r.players[1]
	full.X, full.Y = c.X, c.Y
	r.stepMatch()
	if r.coffeeIn[0] != 0 || full.HP != playerHP {
		t.Fatal("full-HP player took the coffee")
	}
	hurt.HP, hurt.X, hurt.Y = 2, c.X, c.Y
	r.stepMatch()
	if hurt.HP != 3 || r.coffeeIn[0] != coffeeRespawn {
		t.Fatalf("hurt player HP %d, coffee timer %d", hurt.HP, r.coffeeIn[0])
	}
	steps(r, coffeeRespawn)
	if r.coffeeIn[0] != 0 {
		t.Fatal("coffee did not come back")
	}
}

func TestDashIsInvulnerable(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	p := r.players[0]
	p.pressed.dash = true
	r.stepMatch()
	if p.action != actDash || r.canHurt(p) {
		t.Fatalf("dash: action %q canHurt %v", p.action, r.canHurt(p))
	}
}

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestBossWalksTowardTargetWithinLimits(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	p := r.players[0]
	ts := float64(r.level.Tile)
	start := r.boss.X
	r.boss.State, r.boss.ticks = bossIdle, bossIdleTicks
	for i := 0; r.boss.State == bossIdle || r.boss.State == bossWalk; i++ {
		r.stepMatch()
		if i > 5*tickRate {
			t.Fatal("walk never ended")
		}
	}
	walked := start - r.boss.X
	if walked <= 0 || walked > walkMaxTiles*ts+1 {
		t.Fatalf("walked %v px toward a player at x=%v, want 0 < d ≤ %v", walked, p.X, walkMaxTiles*ts)
	}
	if r.boss.X < r.bossMinX || r.boss.X > r.bossMaxX {
		t.Fatalf("boss at %v outside [%v, %v]", r.boss.X, r.bossMinX, r.bossMaxX)
	}
	if lo := r.bossMinX - bossW/2; lo < 32+wallGapTiles*ts {
		t.Fatalf("wall gap too small: body edge can reach %v", lo)
	}
}

func TestChargeHitsFloorOnceAndMissesPlatform(t *testing.T) {
	r := testRoom(t, 2, 0)
	settle(r)
	floor, high := r.players[0], r.players[1]
	ts := float64(r.level.Tile)
	r.boss.X = 1100
	floor.X = 900
	low := r.level.PlatformTops[len(r.level.PlatformTops)-1]
	high.X, high.Y, high.VY = 950, low, 0 // on the right low desk, above the corridor
	steps(r, 2)
	if !r.lockCharge() {
		t.Fatal("expected room to charge left")
	}
	if r.boss.chargeTo >= r.boss.X || r.boss.X-r.boss.chargeTo > chargeTiles*ts+1e-9 {
		t.Fatalf("charge %v → %v", r.boss.chargeFrom, r.boss.chargeTo)
	}
	r.boss.State, r.boss.ticks = bossChargeWindup, 0
	steps(r, windupTicks)
	if floor.Hurts != 0 {
		t.Fatal("hurt during the windup")
	}
	for range chargeTicks + recoverTicks {
		r.stepMatch()
		floor.invuln = 0 // a second overlap would hurt again if not tracked per charge
	}
	if floor.Hurts != 1 || high.Hurts != 0 {
		t.Fatalf("floor hurts %d (want 1), platform hurts %d (want 0)", floor.Hurts, high.Hurts)
	}
}

func TestJumpClearsCharge(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	p := r.players[0]
	r.boss.X, p.X = 1100, 900
	if !r.lockCharge() {
		t.Fatal("no room")
	}
	r.boss.State, r.boss.ticks = bossChargeWindup, 0
	steps(r, windupTicks)
	for i := range chargeTicks {
		if r.boss.X-bossW/2 < p.X+hitW/2+90 && p.OnGround { // jump as the body closes in
			p.pressed.jump = true
		}
		r.stepMatch()
		_ = i
	}
	if p.Hurts != 0 {
		t.Fatal("a well-timed jump should clear the charging body")
	}
}

func TestChargeNearWallGoesTheOtherWay(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	r.players[0].X = 1230 // target behind the boss, toward the near wall
	r.boss.X = r.bossMaxX - 10
	r.boss.State, r.boss.ticks, r.boss.cycle = bossWalk, walkMaxTicks+1, 2 // next: charge
	r.stepMatch()
	if r.boss.State != bossChargeWindup || r.boss.chargeTo >= r.boss.chargeFrom || r.boss.Facing != -1 {
		t.Fatalf("state %q %v→%v, want a charge away from the wall", r.boss.State, r.boss.chargeFrom, r.boss.chargeTo)
	}
	if r.boss.chargeTo < r.bossMinX {
		t.Fatalf("charge ends at %v, inside the wall gap", r.boss.chargeTo)
	}
}

func TestEnragedAddsSecondWaveSolo(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	r.boss.HP = r.boss.MaxHP * 0.4
	r.boss.State, r.boss.ticks = bossWave, 0
	most := 0
	for range waveTicks {
		r.stepMatch()
		most = max(most, len(r.waves))
	}
	if most != 4 {
		t.Fatalf("%d waves in flight, want 4 (two pairs) when enraged", most)
	}
}

func TestSmashChargeScalesAndDashCancels(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	r.boss.State = passive
	p := r.players[0]
	p.X, p.Facing = r.boss.X-bossW/2-hitW/2-20, 1
	hp := r.boss.HP
	p.pressed.special, p.in.Special = true, true
	steps(r, chargeMaxTicks+castTicks+2) // held to the cap: auto-release
	if !near(r.boss.HP, hp-smashMax) || p.Casts != 1 {
		t.Fatalf("full smash: boss HP %v want %v, casts %d", r.boss.HP, hp-smashMax, p.Casts)
	}
	if p.skillCD == 0 {
		t.Fatal("no cooldown after release")
	}

	p.skillCD = 0
	p.pressed.special = true
	r.stepMatch() // charging
	p.pressed.dash = true
	r.stepMatch()
	if p.Cancels != 1 || p.action != actDash || p.skillCD == 0 {
		t.Fatalf("dash: cancels %d action %q cd %d", p.Cancels, p.action, p.skillCD)
	}
}

func TestShieldReflectsUpToThree(t *testing.T) {
	r := testRoom(t, 2, 0)
	settle(r)
	r.boss.State = passive
	p := r.players[1]      // P2
	p.X, p.Facing = 600, 1 // boss is to the right
	p.pressed.special = true
	r.stepMatch()
	for i := range 4 {
		r.cards = append(r.cards, &Shot{ID: 100 + i, Kind: "card", X: p.X + 40, Y: p.Y - 30, VX: -cardSpeed, W: cardW, H: cardH})
	}
	steps(r, 4)
	if p.Reflects != maxReflects || p.Hurts != 1 {
		t.Fatalf("reflects %d (want %d), hurts %d (want 1)", p.Reflects, maxReflects, p.Hurts)
	}
	returned := 0
	for _, s := range r.shots {
		if s.Kind == "returned" && s.VX > 0 {
			returned++
		}
	}
	if returned != maxReflects {
		t.Fatalf("%d returned cards flying at the boss", returned)
	}
}

func TestStampMarksAndBoostsKeyboard(t *testing.T) {
	r := testRoom(t, 3, 0)
	settle(r)
	r.boss.State = passive
	amber := r.players[2]
	amber.X, amber.Facing = r.boss.X-120, 1
	amber.pressed.special = true
	steps(r, 20)
	if r.boss.markTicks == 0 || r.boss.Marks != 1 {
		t.Fatal("stamp did not mark the boss")
	}
	hp := r.boss.HP
	r.damageBoss(1, true)
	r.damageBoss(4, false) // timecard: not boosted
	if !near(hp-r.boss.HP, markBonus+4) {
		t.Fatalf("damage %v, want %v", hp-r.boss.HP, markBonus+4)
	}
	steps(r, markTicks)
	if r.boss.markTicks != 0 {
		t.Fatal("mark did not expire")
	}
}

func TestHotfixHealsOncePerPlayerNotDowned(t *testing.T) {
	r := testRoom(t, 4, 0)
	settle(r)
	r.boss.State = passive
	mint := r.players[3]
	for _, p := range r.players {
		p.X = 400
	}
	mint.HP, r.players[0].HP = 2, 1
	r.hurt(r.players[1], playerHP, 0) // downed: must not be revived
	mint.pressed.special = true
	steps(r, hotfixTicks+5)
	if mint.HP != 3 || r.players[0].HP != 2 || r.players[1].action != actDown || r.players[2].Healed != 0 {
		t.Fatalf("HP mint %d, p1 %d, downed %q, full-HP healed %d", mint.HP, r.players[0].HP, r.players[1].action, r.players[2].Healed)
	}
	if len(r.zones) != 0 {
		t.Fatal("zone did not expire")
	}
}

func TestTimecardHitsOnceStopsAtWallSharesCooldown(t *testing.T) {
	r := testRoom(t, 5, 0)
	settle(r)
	r.boss.State = passive
	orange := r.players[4]
	orange.X, orange.Facing = r.boss.X-60, 1
	hp := r.boss.HP
	orange.pressed.special = true
	r.stepMatch()
	if r.canHurt(orange) {
		t.Fatal("not invulnerable at the start of the lunge")
	}
	steps(r, timecardTicks+2)
	if !near(hp-r.boss.HP, timecardDamage) || orange.SkillHits != 1 {
		t.Fatalf("boss lost %v, want %v once", hp-r.boss.HP, timecardDamage)
	}
	if orange.dashCD == 0 {
		t.Fatal("normal dash not on the shared cooldown")
	}

	orange.skillCD, orange.dashCD = 0, 0
	orange.X = r.level.Spawns[0].X // near the left wall
	orange.Facing = -1
	orange.pressed.special = true
	steps(r, timecardTicks-2) // 54 px to the wall: stops before the full 3 tiles
	if orange.action == actTimecard {
		t.Fatalf("lunge did not stop at the wall (x=%v)", orange.X)
	}
	if orange.X != 32+hitW/2 {
		t.Fatalf("went through the wall: x=%v", orange.X)
	}
}

func TestHardModeNoInvulnerabilityAfterHit(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	r.hard, r.boss.State = true, passive
	p := r.players[0]
	for i := range 2 { // two cards back to back
		r.cards = append(r.cards, &Shot{ID: 900 + i, Kind: "card", X: p.X + 30, Y: p.Y - 22, VX: -cardSpeed, W: cardW, H: cardH})
		steps(r, 3)
	}
	if p.Hurts != 2 || p.invuln != 0 {
		t.Fatalf("hurts %d invuln %d, want 2 hits and no grace", p.Hurts, p.invuln)
	}
	// a wave still hits only once while it passes through the player
	r.waves = append(r.waves, &Wave{ID: 950, X: p.X - 40, Y: r.boss.Y, Dir: 1, hit: map[int]bool{}})
	steps(r, 20)
	if p.Hurts != 3 {
		t.Fatalf("hurts %d, want 3: one more from the wave", p.Hurts)
	}
}

func TestPhase2ChargeStillJumpable(t *testing.T) {
	r := testRoom(t, 1, 0)
	settle(r)
	r.boss.State = passive
	r.boss.HP = r.boss.MaxHP * 0.4
	p := r.players[0]
	r.boss.X, p.X = 900, 700 // open floor: the desk above is too high to land on, so only a jump over works
	if !r.lockCharge() {
		t.Fatal("no room")
	}
	r.boss.State, r.boss.ticks = bossChargeWindup, 0
	for i := range r.t(windupTicks) + r.chargeDur() + 5 {
		if i == r.t(windupTicks)-6 { // jump just before the body launches
			p.pressed.jump = true
		}
		r.stepMatch()
	}
	if p.Hurts != 0 {
		t.Fatal("a well-timed jump should still clear the twice-as-fast phase 2 charge")
	}
}
