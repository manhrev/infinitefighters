package main

import (
	"math"
	"testing"
)

func sandboxRoom(t *testing.T) *Room {
	t.Helper()
	r := testRoom(t, 1, 0)
	r.startLevel(levelSandbox)
	settle(r)
	return r
}

func TestSandboxNeverEndsAndRecovers(t *testing.T) {
	r := sandboxRoom(t)
	p := r.players[0]
	p.action = actDown
	steps(r, sandboxRespawnTicks+2)
	if p.action == actDown || p.HP != playerHP {
		t.Fatalf("downed player did not get up: action %q HP %d", p.action, p.HP)
	}
	r.boss.HP = 0
	steps(r, 2)
	if r.boss.HP != r.boss.MaxHP || r.result != "" {
		t.Fatalf("boss HP %v, result %q after beaten", r.boss.HP, r.result)
	}
	steps(r, r.deadlineTicks+1)
	if r.result != "" {
		t.Fatalf("sandbox lost on the DEADLINE: %q", r.result)
	}
}

func TestKhabibLoopsAllThreeAttacks(t *testing.T) {
	r := sandboxRoom(t)
	p := r.players[0]
	seen := map[string]bool{}
	for range 30 * tickRate {
		p.HP, p.invuln = playerHP, 0 // keep the target alive
		r.stepMatch()
		seen[r.boss.State] = true
	}
	for _, s := range []string{khPress, khSpam, khTeleOut, khTeleWarn, khTeleIn, khRecover} {
		if !seen[s] {
			t.Errorf("boss never entered %q", s)
		}
	}
}

func TestTeleportLandsBehindPlayerAndSlams(t *testing.T) {
	r := sandboxRoom(t)
	p := r.players[0]
	p.X, p.Facing = r.boss.X-300, 1 // facing right, the boss ahead
	r.boss.State, r.boss.ticks = khTeleOut, 0
	for r.boss.State != khTeleWarn {
		r.stepMatch()
	}
	if r.boss.landX >= p.X {
		t.Fatalf("lands at %v, want behind (left of) the player at %v", r.boss.landX, p.X)
	}
	if overlaps(shotBox(&Shot{X: p.X, Y: p.Y - hitH/2, W: 1, H: 1}), r.bossBox()) {
		t.Fatal("boss is hittable while out of sight")
	}
	hp := p.HP
	for r.boss.State != khTeleIn {
		r.stepMatch()
	}
	r.stepMatch() // the slam lands on the first tick of teleport_in
	if p.HP != hp-1 {
		t.Fatalf("standing still: HP %d, want %d", p.HP, hp-1)
	}
}

func TestNagCardsCarryWords(t *testing.T) {
	r := sandboxRoom(t)
	r.boss.State, r.boss.ticks = khSpam, 0
	r.stepMatch()
	if len(r.cards) != 1 || r.cards[0].Word != khWords[0] {
		t.Fatalf("cards %+v", r.cards)
	}
}

func TestEndOnlyFromSandboxHost(t *testing.T) {
	r := testRoom(t, 1, 0) // normal match
	c := r.players[0].client
	r.handle(c, clientMsg{Type: "end"})
	if r.phase != phasePlaying {
		t.Fatal("end worked outside the sandbox")
	}
	r.startLevel(levelSandbox)
	r.handle(c, clientMsg{Type: "end"})
	if r.phase != phaseLobby || r.sandbox != "" {
		t.Fatalf("phase %q sandbox %q after end", r.phase, r.sandbox)
	}
}

func TestKhabibWalksOffDeskAndFalls(t *testing.T) {
	r := sandboxRoom(t)
	p := r.players[0]
	ts := float64(r.level.Tile)
	top := r.level.PlatformTops[0]
	row := int(top / ts)
	col := 0
	for r.level.at(col, row) != tileOneWay {
		col++
	}
	end := col
	for r.level.at(end+1, row) == tileOneWay {
		end++
	}
	// Boss near the right end of a desk, the player on the floor to the right of it.
	r.boss.X, r.boss.Y, r.boss.State, r.boss.ticks, r.boss.chargeHit = float64(end)*ts, top, khPressWindup, 0, map[int]bool{}
	fell := false
	for range 6 * tickRate {
		p.X, p.Y, p.VY, p.HP, p.invuln = float64(end+3)*ts, r.level.Boss.Y, 0, playerHP, 0
		r.stepMatch()
		if r.boss.Y > top {
			fell = true
		}
	}
	if !fell || r.boss.Y != r.level.Boss.Y {
		t.Fatalf("boss y %v (desk %v, floor %v), fell %v: want it to walk off and land on the floor", r.boss.Y, top, r.level.Boss.Y, fell)
	}
}

func TestPressWaveHitsFloorFromADesk(t *testing.T) {
	r := sandboxRoom(t)
	top := r.level.PlatformTops[0]
	r.boss.Y = top
	r.pressWave()
	floor := r.level.Boss.Y
	if len(r.waves) != 2 || r.waves[0].Y != floor || r.waves[1].Y != top {
		t.Fatalf("waves %+v, want one on the floor and one on the desk", r.waves)
	}
	r.boss.Y, r.waves = floor, nil
	r.pressWave()
	if len(r.waves) != 1 {
		t.Fatalf("on the floor: %d waves, want 1", len(r.waves))
	}
}

// landings runs a chain and returns the boss position at each landing.
func landings(r *Room, keepAlive func()) (xs, ys []float64) {
	was := r.boss.State
	for range 20 * tickRate {
		keepAlive()
		r.stepMatch()
		if r.boss.State == khTeleIn && was != khTeleIn {
			xs, ys = append(xs, r.boss.X), append(ys, r.boss.Y)
		}
		if r.boss.State == khRecover {
			break
		}
		was = r.boss.State
	}
	return
}

func TestChainHitsEachPlayerOnceInOrder(t *testing.T) {
	r := testRoom(t, 3, 0)
	r.startLevel(levelSandbox)
	settle(r)
	floor := r.level.Boss.Y
	for i, p := range r.players {
		p.X, p.Y, p.Facing = r.bossMinX+float64(i)*200, floor, 1
	}
	keep := func() {
		for i, p := range r.players {
			p.X, p.Y, p.VY, p.HP, p.invuln = r.bossMinX+float64(i)*200, floor, 0, playerHP, 0
		}
	}
	r.boss.State, r.boss.ticks = khChainWind, 0
	for r.boss.State == khChainWind {
		r.stepMatch()
	}
	order := append([]int{}, r.boss.queue...)
	if len(order) != 3 || r.boss.chain != 3 {
		t.Fatalf("queue %v chain %d, want all 3 players", order, r.boss.chain)
	}
	xs, _ := landings(r, keep)
	if len(xs) != 3 {
		t.Fatalf("%d landings, want 3", len(xs))
	}
	for i, id := range order {
		want, _ := r.behind(r.byID(id))
		if math.Abs(xs[i]-want) > 10 { // players get knocked back a few px by earlier slams
			t.Errorf("hop %d landed at %v, want behind P%d at %v", i, xs[i], id, want)
		}
	}
}

func TestChainAloneHopsToRandomSpots(t *testing.T) {
	r := sandboxRoom(t)
	keep := func() { r.players[0].HP, r.players[0].invuln = playerHP, 0 }
	r.boss.State, r.boss.ticks = khChainWind, 0
	for r.boss.State == khChainWind {
		r.stepMatch()
	}
	if r.boss.chain < 2 || r.boss.chain > 4 || len(r.boss.queue) != 0 {
		t.Fatalf("chain %d queue %v, want 2-4 hops and no queue", r.boss.chain, r.boss.queue)
	}
	n := r.boss.chain
	xs, ys := landings(r, keep)
	if len(xs) != n {
		t.Fatalf("%d landings, want %d", len(xs), n)
	}
	for i := range xs {
		if xs[i] < r.bossMinX || xs[i] > r.bossMaxX {
			t.Errorf("hop %d landed at %v outside the arena", i, xs[i])
		}
		if ys[i] > r.level.Boss.Y {
			t.Errorf("hop %d below the floor: %v", i, ys[i])
		}
	}
}
