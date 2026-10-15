package main

import "testing"

func yasukaRoom(t *testing.T, n int) *Room {
	t.Helper()
	r := testRoom(t, n, 0)
	r.startLevel(levelSandboxYasuka)
	settle(r)
	return r
}

func TestYasukaHitboxAndLoop(t *testing.T) {
	r := yasukaRoom(t, 1)
	if r.boss.W != 52 || r.boss.H != 100 {
		t.Fatalf("hitbox %vx%v, want 52x100", r.boss.W, r.boss.H)
	}
	p := r.players[0]
	seen := map[string]bool{}
	for range 40 * tickRate {
		p.HP, p.invuln = playerHP, 0
		if r.boss.State == ySilence {
			r.boss.dome = 0 // someone breaks the dome
		}
		r.stepMatch()
		seen[r.boss.State] = true
	}
	for _, s := range []string{yDodgeS, yProc, ySilence, yFlurryS, khRecover} {
		if !seen[s] {
			t.Errorf("boss never entered %q", s)
		}
	}
}

func TestDodgeSlidesOffTheFirstHitThenOpens(t *testing.T) {
	r := yasukaRoom(t, 1)
	p := r.players[0]
	p.X = r.boss.X - 100
	r.boss.State, r.boss.ticks = yDodgeS, 0
	hp, x := r.boss.HP, r.boss.X
	r.damageBoss(5, false)
	if r.boss.HP != hp || r.boss.State != khRecover || r.boss.X <= x {
		t.Fatalf("HP %v (was %v) state %q x %v (was %v): want no damage, open, and slid away from the player", r.boss.HP, hp, r.boss.State, r.boss.X, x)
	}
	r.damageBoss(2, false)
	if r.boss.HP != hp-4 {
		t.Fatalf("open window: HP %v, want %v (x2 damage)", r.boss.HP, hp-4)
	}
}

func TestDomeSoaksUntilItBreaksThenBursts(t *testing.T) {
	r := yasukaRoom(t, 1)
	r.boss.State, r.boss.ticks, r.boss.dome = ySilence, 0, yDomeHP
	hp := r.boss.HP
	for range int(yDomeHP) {
		r.damageBoss(1, false)
	}
	if r.boss.HP != hp {
		t.Fatalf("boss took damage through the dome: %v", r.boss.HP)
	}
	r.stepMatch()
	if r.boss.State != yFlurryWindS {
		t.Fatalf("state %q after the dome broke, want the paper windup", r.boss.State)
	}
	for r.boss.State != yFlurryS {
		r.stepMatch()
	}
	r.stepMatch()
	if len(r.cards) < 2*len(r.level.PlatformTops) {
		t.Fatalf("%d cards, want paper along every desk height both ways", len(r.cards))
	}
}

func TestDomeTimesOutWithoutBursting(t *testing.T) {
	r := yasukaRoom(t, 1)
	r.boss.State, r.boss.ticks, r.boss.dome = ySilence, 0, yDomeHP
	steps(r, yDomeTicks+2)
	if r.boss.State != khRecover || len(r.cards) != 0 {
		t.Fatalf("state %q cards %d, want an unbroken dome to just drop", r.boss.State, len(r.cards))
	}
}

func TestSealSendsPlayerBackToStartOnceAwake(t *testing.T) {
	r := yasukaRoom(t, 1)
	p := r.players[0]
	p.X, p.Y = r.boss.X-400, r.boss.Y
	r.boss.State, r.boss.ticks = yProc, 0
	r.boss.Facing = -1
	r.stepMatch() // lays the seals; one is under the player
	start := r.level.Spawns[p.Slot]
	steps(r, ySealWarn-5) // still asleep
	if p.X == start.X {
		t.Fatal("seal bit while it was still warning")
	}
	p.X = r.boss.X - 400
	steps(r, 10)
	if p.X != start.X || p.Rejects != 1 {
		t.Fatalf("player at %v rejects %d, want sent back to %v", p.X, p.Rejects, start.X)
	}
}

func TestDashPassesOverASeal(t *testing.T) {
	r := yasukaRoom(t, 1)
	p := r.players[0]
	p.X = r.boss.X - 400
	r.boss.State, r.boss.ticks = passive, 0
	r.seals = []*Seal{{ID: 1, X: p.X, Y: p.Y}}
	p.action, p.actionTicks = actDash, dashTicks
	r.stepMatch()
	if p.Rejects != 0 {
		t.Fatal("a dashing player was rejected")
	}
}

func TestFormsOnePaperPerPlayerAndTheyHome(t *testing.T) {
	r := yasukaRoom(t, 3)
	floor := r.level.Boss.Y
	spot := func() {
		for i, p := range r.players {
			p.X, p.Y, p.VY, p.HP, p.invuln = r.boss.X-200-float64(i)*120, floor, 0, playerHP, 0
		}
	}
	spot()
	r.boss.State, r.boss.ticks = yFormsS, 0
	r.stepMatch()
	if len(r.papers) != 3 {
		t.Fatalf("%d papers, want one per player", len(r.papers))
	}
	targets := map[*Player]bool{}
	for _, pp := range r.papers {
		targets[pp.target] = true
	}
	if len(targets) != 3 {
		t.Fatalf("papers target %d different players, want 3", len(targets))
	}
	r.boss.State = passive
	for range 4 * tickRate {
		spot() // everyone stands still
		r.stepMatch()
	}
	var hurts uint32 // a paper hits whoever it touches first, so a teammate in the way can take it
	for _, p := range r.players {
		hurts += p.Hurts
	}
	if hurts != 3 {
		t.Errorf("%d hits, want all three papers to land", hurts)
	}
	if len(r.papers) != 0 {
		t.Fatalf("%d papers left after they all landed", len(r.papers))
	}
}

func TestPaperCanBeSwattedShotDownAndDashed(t *testing.T) {
	r := yasukaRoom(t, 1)
	p := r.players[0]
	p.X, p.Y, p.Facing = r.boss.X-300, r.boss.Y, 1
	r.boss.State = passive
	r.papers = []*Paper{{ID: 1, X: p.X + 25, Y: p.Y - hitH/2, target: p, life: 100}}
	p.pressed.attack = true
	steps(r, attackTicks)
	if len(r.papers) != 0 || p.Hurts != 0 {
		t.Fatalf("swing: %d papers, %d hurts, want the paper swatted", len(r.papers), p.Hurts)
	}

	r.papers = []*Paper{{ID: 2, X: p.X + 60, Y: p.Y - hitH/2, target: p, life: 100}}
	r.addShot(p, "pulse", 32, 16, pulseSpeed, pulseDamage, pulseLife)
	steps(r, 6)
	if len(r.papers) != 0 || p.Hurts != 0 {
		t.Fatalf("pulse: %d papers, %d hurts, want the paper shot down", len(r.papers), p.Hurts)
	}

	r.papers = []*Paper{{ID: 3, X: p.X + 15, Y: p.Y - hitH/2, target: p, life: 100}}
	p.action, p.actionTicks = actDash, dashTicks
	r.stepMatch()
	if p.Hurts != 0 {
		t.Fatal("a dashing player was hit by the paper")
	}
}

func TestPaperExpires(t *testing.T) {
	r := yasukaRoom(t, 1)
	p := r.players[0]
	r.boss.State = passive
	r.papers = []*Paper{{ID: 1, X: p.X + 500, Y: p.Y - 400, target: p, life: 10}}
	steps(r, 12)
	if len(r.papers) != 0 {
		t.Fatal("paper outlived its life")
	}
}
