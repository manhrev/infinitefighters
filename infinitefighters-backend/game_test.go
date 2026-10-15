package main

import "testing"

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

func steps(r *Room, n int) {
	for range n {
		r.stepMatch()
	}
}

// settle lets everyone land on the floor.
func settle(r *Room) { steps(r, 30) }

func TestBossHPScalesWithAllSlots(t *testing.T) {
	r := testRoom(t, 2, 1) // away players count too
	if want := bossBaseHP + 2*bossHPPerPlayer; r.boss.MaxHP != want {
		t.Fatalf("boss HP %d, want %d", r.boss.MaxHP, want)
	}
	r.players[0].action = actDown
	r.players[1].action = actDown
	steps(r, lostTicks+2) // lost, then restart
	if r.result != "" || r.boss.MaxHP != bossBaseHP+2*bossHPPerPlayer {
		t.Fatalf("after restart: result %q, boss HP %d", r.result, r.boss.MaxHP)
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
		t.Fatalf("boss HP %d, want %d after one swing", r.boss.HP, hp-1)
	}

	p.Facing = -1 // facing away
	p.pressed.attack = true
	r.stepMatch()
	p.Facing = -1
	steps(r, attackTicks+2)
	if r.boss.HP != hp-1 {
		t.Fatalf("swing facing away hit the boss: HP %d", r.boss.HP)
	}

	p.X, p.Facing = r.boss.X, 1 // standing inside the boss is allowed and still hits
	p.pressed.attack = true
	steps(r, attackTicks+2)
	if r.boss.HP != hp-2 || p.HP != playerHP {
		t.Fatalf("inside the boss: boss HP %d (want %d), player HP %d", r.boss.HP, hp-2, p.HP)
	}
}

func TestVolleyMissesFloor(t *testing.T) {
	r := testRoom(t, 5, 0) // 5 players: both sides, two salvos
	settle(r)
	for i, p := range r.players { // spread across the floor on both sides of the boss
		p.X = []float64{100, 600, 1100, 1450, 1800}[i]
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
	xs := []float64{300, 1500, 300, 1500} // over the two lowest platforms in sprint.json
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
		t.Fatalf("not reset: result %q boss %d/%d deadline %d", r.result, r.boss.HP, r.boss.MaxHP, r.deadline)
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
	r.damageBoss(1)
	r.stepMatch()
	if r.result != resultWon || r.boss.State != bossDefeat {
		t.Fatalf("result %q state %q", r.result, r.boss.State)
	}
	steps(r, wonTicks)
	if r.phase != phaseLobby || len(r.players) != 2 {
		t.Fatalf("phase %q with %d players, want lobby with 2", r.phase, len(r.players))
	}
	r.start()
	if r.phase != phasePlaying || r.boss.HP != bossBaseHP+bossHPPerPlayer {
		t.Fatalf("second start: phase %q boss HP %d", r.phase, r.boss.HP)
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
