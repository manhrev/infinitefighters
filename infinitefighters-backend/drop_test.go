package main

import "testing"

// onDesk puts a player standing on the desk whose top is y, at x.
func onDesk(p *Player, x, y float64) { p.X, p.Y, p.VY, p.OnGround = x, y, 0, true }

// dropReady is a match with the boss under 60% HP, passive until a test starts a turn.
func dropReady(t *testing.T, n int) *Room {
	r := testRoom(t, n, 0)
	settle(r)
	r.boss.HP = r.boss.MaxHP * 0.5
	r.boss.State = passive
	return r
}

// runDrop plays a whole drop turn from the telegraph to the end of the paper jam.
func runDrop(t *testing.T, r *Room) {
	t.Helper()
	if !r.lockDrops() {
		t.Fatal("drop not available")
	}
	r.boss.State, r.boss.ticks = bossDropWindup, 0
	for i := 0; r.boss.State != bossIdle; i++ {
		r.stepMatch()
		if i > 10*tickRate {
			t.Fatal("drop turn never ended")
		}
	}
}

func TestDropOnlyBelow60PercentAndWithSomeoneOnADesk(t *testing.T) {
	r := dropReady(t, 1)
	p := r.players[0]
	if r.lockDrops() {
		t.Fatal("dropped on a player standing on the floor")
	}
	onDesk(p, 300, 480)
	r.boss.HP = r.boss.MaxHP * 0.7
	if r.lockDrops() {
		t.Fatal("dropped above 60% HP")
	}
	r.boss.HP = r.boss.MaxHP * 0.5
	if !r.lockDrops() || len(r.drops) != 1 {
		t.Fatal("no drop on a player on a desk")
	}
	r.drops = nil
	if r.lockDrops() {
		t.Fatal("dropped again within 12 s")
	}
}

func TestDropReplacesVolleyTurnElseVolley(t *testing.T) {
	r := dropReady(t, 1)
	onDesk(r.players[0], 300, 480)
	r.boss.State, r.boss.ticks, r.boss.cycle = bossWalk, walkMaxTicks+1, 0 // next turn: volley
	r.stepMatch()
	if r.boss.State != bossDropWindup {
		t.Fatalf("state %q, want drop_windup", r.boss.State)
	}

	r = dropReady(t, 1) // nobody on a desk: the volley as before
	r.boss.State, r.boss.ticks, r.boss.cycle = bossWalk, walkMaxTicks+1, 0
	r.stepMatch()
	if r.boss.State != bossVolleyWindup {
		t.Fatalf("state %q, want volley_windup", r.boss.State)
	}
}

func TestDropHitsTargetOnceNotLowerOrUpperLevels(t *testing.T) {
	r := dropReady(t, 3)
	target, below, above := r.players[0], r.players[1], r.players[2]
	onDesk(target, 900, 480) // right low desk
	onDesk(below, 900, 608)  // floor right under the column
	onDesk(above, 900, 224)  // right high desk, over the column
	if !r.lockDrops() {
		t.Fatal("no drop")
	}
	if len(r.drops) != 2 { // the floor player is not a target; the high desk gets its own column
		t.Fatalf("%d columns, want 2", len(r.drops))
	}
	r.drops = r.drops[:1] // keep only the low desk column to check its reach
	d := r.drops[0]
	if d.Bottom != 480 || d.Top <= 224 {
		t.Fatalf("column %v→%v: must start under the high desk and stop on the low one", d.Top, d.Bottom)
	}
	r.boss.State, r.boss.ticks = bossDropWindup, 0
	for r.boss.State != bossIdle {
		r.stepMatch()
		target.invuln = 0 // a second overlap would hurt again if not tracked per turn
		onDesk(target, 900, 480)
	}
	if target.Hurts != 1 || below.Hurts != 0 || above.Hurts != 0 {
		t.Fatalf("hurts: target %d (want 1), below %d, above %d (want 0)", target.Hurts, below.Hurts, above.Hurts)
	}
}

func TestDropLeavesLandingRoomOneColumnPerDesk(t *testing.T) {
	r := dropReady(t, 3)
	for _, p := range r.players { // everyone crowds the 4-tile ledge at the left wall
		onDesk(p, 60, 352)
	}
	r.players[0].X = 40 // at the very edge
	if !r.lockDrops() || len(r.drops) != 1 {
		t.Fatalf("%d columns on one desk, want 1", len(r.drops))
	}
	d := r.drops[0]
	ts := float64(r.level.Tile)
	ledge0, ledge1 := 1*ts, 5*ts
	if d.X-dropW/2 < ledge0 || d.X+dropW/2 > ledge1 {
		t.Fatalf("column %v±32 hangs over the ledge edge", d.X)
	}
	if free := (ledge1 - ledge0) - dropW; free < 2*ts {
		t.Fatalf("only %v px of landing room left", free)
	}
}

func TestDropDodgeBySteppingOut(t *testing.T) {
	r := dropReady(t, 1)
	p := r.players[0]
	onDesk(p, 300, 480)
	if !r.lockDrops() {
		t.Fatal("no drop")
	}
	r.boss.State, r.boss.ticks = bossDropWindup, 0
	for range dropWindupTicks / 2 { // react halfway through the telegraph: run out of the column
		p.in = Input{Right: true}
		r.stepMatch()
	}
	p.in = Input{}
	for r.boss.State != bossIdle {
		r.stepMatch()
	}
	if p.Hurts != 0 {
		t.Fatalf("hurt after stepping %v px out of a column at %v", p.X-300, 300)
	}
}

func TestDropSkipsDownedAndAwayAndJamDoublesDamage(t *testing.T) {
	r := testRoom(t, 2, 1)
	settle(r)
	r.boss.HP, r.boss.State = r.boss.MaxHP*0.5, passive
	onDesk(r.players[0], 300, 480)
	r.hurt(r.players[0], playerHP, 0) // downed on a desk
	onDesk(r.players[2], 950, 480)    // away on a desk
	if r.lockDrops() {
		t.Fatal("targeted a downed or away player")
	}
	onDesk(r.players[1], 640, 352)
	runDrop(t, r)
	hp := r.boss.HP
	r.boss.State = bossDropJam
	r.damageBoss(1, false)
	if !near(hp-r.boss.HP, 2) {
		t.Fatalf("paper jam damage %v, want 2", hp-r.boss.HP)
	}
}
