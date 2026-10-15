package main

import "testing"

func voughtRoom(t *testing.T, n int) *Room {
	t.Helper()
	r := testRoom(t, n, 0)
	r.startLevel(levelSandboxVought)
	settle(r)
	return r
}

func TestVoughtHitboxAndLoop(t *testing.T) {
	r := voughtRoom(t, 1)
	if r.boss.W != 36 || r.boss.H != 95 {
		t.Fatalf("hitbox %vx%v, want 36x95", r.boss.W, r.boss.H)
	}
	p := r.players[0]
	seen := map[string]bool{}
	for range 40 * tickRate {
		p.HP, p.invuln = playerHP, 0
		r.stepMatch()
		seen[r.boss.State] = true
		if p.action == actPat { // let him finish the pat
			p.patPresses = 0
		}
	}
	for _, s := range []string{vDeflect, vPatOut, vPatWarn, vPat, vStamp, vReflect, khRecover} {
		if !seen[s] {
			t.Errorf("boss never entered %q", s)
		}
	}
}

func TestLobsArriveAtThePlayer(t *testing.T) {
	r := voughtRoom(t, 1)
	p := r.players[0]
	p.X = r.boss.X - 300
	r.boss.State, r.boss.ticks, r.boss.nags = vDeflect, 0, 1 // the middle of the three spots: straight at the player
	r.stepMatch()
	if len(r.cards) != 1 || r.cards[0].Word == "" || r.cards[0].G == 0 {
		t.Fatalf("cards %+v, want one lobbed task", r.cards)
	}
	hp := p.HP
	for range 2 * tickRate {
		p.X, p.Y, p.VY = r.boss.X-300, r.boss.Y, 0
		r.stepMatch()
		if p.HP < hp {
			return
		}
	}
	t.Fatal("the lobbed task never hit the player")
}

func patRoom(t *testing.T, n int) (*Room, *Player) {
	r := voughtRoom(t, n)
	p := r.players[0]
	p.X, p.Facing = r.boss.X-250, 1
	r.boss.State, r.boss.ticks, r.boss.patTarget = vPatOut, 0, nil
	for r.boss.State != vPat {
		r.stepMatch()
	}
	return r, p
}

func TestPatHoldsThenTasksTheListener(t *testing.T) {
	r, p := patRoom(t, 1)
	if p.action != actPat {
		t.Fatalf("action %q, want held", p.action)
	}
	x := p.X
	steps(r, vPatHoldTicks+2)
	if p.action == actPat || p.dashCD < vBurdenTicks/2 || p.skillCD < vBurdenTicks/2 {
		t.Fatalf("after the pat: action %q dashCD %d skillCD %d, want a task on their back", p.action, p.dashCD, p.skillCD)
	}
	if p.X != x {
		t.Fatalf("held player moved from %v to %v", x, p.X)
	}
}

func TestPatBreaksWhenAloneByMashing(t *testing.T) {
	r, p := patRoom(t, 1)
	for range vPatEscape {
		p.pressed.attack = true
		r.stepMatch()
	}
	steps(r, 2)
	if p.action == actPat || p.dashCD > 0 {
		t.Fatalf("action %q dashCD %d, want free and not tasked", p.action, p.dashCD)
	}
}

func TestPatInterruptedByTeammateHit(t *testing.T) {
	r, p := patRoom(t, 2)
	r.damageBoss(1, false) // a teammate lands a hit
	steps(r, 2)
	if p.action == actPat || p.dashCD > 0 || r.boss.State != khRecover {
		t.Fatalf("action %q dashCD %d boss %q, want freed, not tasked, boss staggered", p.action, p.dashCD, r.boss.State)
	}
}

func TestStampBreaksUnderSwingsAndOtherwiseHits(t *testing.T) {
	r := voughtRoom(t, 2)
	a, b := r.players[0], r.players[1]
	a.X, b.X, a.Facing, b.Facing = r.boss.X-300, r.boss.X-300+30, 1, -1
	r.boss.State, r.boss.ticks = vStamp, 0
	r.stepMatch()
	if len(r.stamps) != 1 {
		t.Fatalf("%d stamps, want 1", len(r.stamps))
	}
	s := r.stamps[0]
	victim, other := s.target, a
	if victim == a {
		other = b
	}
	hp := victim.HP
	for range vStampHP { // the teammate swings at it
		other.pressed.attack = true
		steps(r, attackTicks+2)
	}
	if len(r.stamps) != 0 || victim.HP != hp {
		t.Fatalf("stamps %d, victim HP %d (was %d): want it broken before it lands", len(r.stamps), victim.HP, hp)
	}

	r.boss.State, r.boss.ticks = vStamp, 0 // ignored this time
	r.stepMatch()
	if len(r.stamps) != 1 {
		t.Fatal("no second stamp")
	}
	r.boss.State = passive
	victim = r.stamps[0].target
	hp = victim.HP
	steps(r, vStampTicks+2)
	if victim.HP != hp-1 || victim.dashCD == 0 {
		t.Fatalf("unbroken stamp: HP %d (was %d), dashCD %d, want 1 damage and a burden", victim.HP, hp, victim.dashCD)
	}
}

func TestShieldSoaksHitsAndReturnsShots(t *testing.T) {
	r := voughtRoom(t, 1)
	p := r.players[0]
	p.X, p.Facing = r.boss.X-60, 1
	r.boss.State, r.boss.ticks = vReflect, 0
	hp, hits := r.boss.HP, r.boss.Hits
	r.damageBoss(5, false)
	if r.boss.HP != hp || r.boss.Hits != hits || len(r.cards) != 1 {
		t.Fatalf("melee: HP %v hits %d cards %d, want no damage and one counter task", r.boss.HP, r.boss.Hits, len(r.cards))
	}
	r.cards = nil
	r.boss.counterCD = 0
	r.addShot(p, "pulse", 32, 16, pulseSpeed, pulseDamage, pulseLife)
	for len(r.cards) == 0 && r.boss.HP == hp {
		r.stepMatch()
	}
	if r.boss.HP != hp {
		t.Fatalf("pulse hurt the shielded boss: HP %v", r.boss.HP)
	}
	back := false
	for _, c := range r.cards {
		back = back || c.VX < 0
	}
	if !back {
		t.Fatalf("cards %+v, want the pulse sent back toward the player", r.cards)
	}
}
