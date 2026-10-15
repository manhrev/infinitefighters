package main

import (
	"math/rand/v2"
	"testing"
)

func partyLevelForTest(t *testing.T) *Level {
	t.Helper()
	data, err := levelFS.ReadFile("levels/party.json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := parseLevel(data)
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// partyRoom starts the party with n connected players, past the countdown when live.
func partyRoom(t *testing.T, n int, live bool) *Room {
	t.Helper()
	r := testRoom(t, n, 0)
	r.partyLevel = partyLevelForTest(t)
	r.startParty()
	if live {
		for range partyCountdown {
			r.stepParty()
		}
	}
	return r
}

func partySteps(r *Room, n int) {
	for range n {
		r.stepParty()
	}
}

func coinsInPlay(r *Room) int {
	total := 0
	for _, b := range r.party.bundles {
		total += b.Value
	}
	for _, p := range r.players {
		total += p.Wallet + p.Balance
	}
	return total
}

// stand puts a player on the floor at x, idle.
func stand(r *Room, p *Player, x float64) {
	p.X, p.Y, p.VY, p.OnGround = x, r.level.Spawns[0].Y, 0, true
	p.in, p.action, p.actionTicks = Input{}, "", 0
}

func TestPartyMapParses(t *testing.T) {
	l := partyLevelForTest(t)
	if !l.HasSafe || l.HasBoss || len(l.ATMs) != 2 || len(l.Spawns) < maxPlayers || len(l.Respawns) < 2 {
		t.Fatalf("safe %v boss %v atms %d spawns %d respawns %d", l.HasSafe, l.HasBoss, len(l.ATMs), len(l.Spawns), len(l.Respawns))
	}
	apex := jumpSpeed * jumpSpeed / (2 * gravity)
	below := l.Spawns[0].Y
	for i := len(l.PlatformTops) - 1; i >= 0; i-- {
		if step := below - l.PlatformTops[i]; step <= 0 || step > apex-16 {
			t.Fatalf("desk level y=%v is %v px above the one below, jump apex %v", l.PlatformTops[i], step, apex)
		}
		below = l.PlatformTops[i]
	}
}

func TestPartyStartAndCountdown(t *testing.T) {
	r := partyRoom(t, 2, false)
	a, b := r.players[0], r.players[1]
	if r.mode != modeParty || a.Health != pvpHP || a.Wallet != 0 || a.Balance != 0 {
		t.Fatalf("mode %q hp %v wallet %d", r.mode, a.Health, a.Wallet)
	}
	stand(r, a, 600)
	stand(r, b, 620)
	a.pressed.attack = true
	partySteps(r, attackTicks+2)
	if b.Health != pvpHP || len(r.party.bundles) != 0 {
		t.Fatal("hit or spray during the countdown")
	}
	partySteps(r, partyCountdown)
	if r.party.sprayed != sprayPerPlayer*2 {
		t.Fatalf("sprayed %d at the start, want %d", r.party.sprayed, sprayPerPlayer*2)
	}
}

func TestPartySprayScheduleAndEnd(t *testing.T) {
	r := partyRoom(t, 3, true)
	for _, p := range r.players { // keep everyone out of the way
		stand(r, p, 40)
		p.client = nil
		p.offline = -1 << 30
	}
	partySteps(r, partyTicks)
	want := 10*sprayPerPlayer*3 + finalPerPlayer*3
	if r.party.sprayed != want || !r.party.over {
		t.Fatalf("sprayed %d (want %d), over %v", r.party.sprayed, want, r.party.over)
	}
	if got := coinsInPlay(r); got != want {
		t.Fatalf("%d coins in play, sprayed %d", got, want)
	}
	partySteps(r, partyResultTicks)
	if r.phase != phaseLobby || r.mode != modeCoop || r.level != r.coopLevel {
		t.Fatal("did not return to the lobby with the boss map")
	}
}

func TestPartyHitDropsTenPercentKillDropsHalf(t *testing.T) {
	r := partyRoom(t, 2, true)
	a, b := r.players[0], r.players[1]
	r.party.bundles = nil
	b.Wallet = 95
	if !r.pvpDamage(a, b, 10, false) || b.Wallet != 95-10 { // ceil(9.5) = 10
		t.Fatalf("wallet %d after a hit, want 85", b.Wallet)
	}
	if r.pvpDamage(a, b, 10, false) || b.Wallet != 85 {
		t.Fatal("a hit during invulnerability dropped coins")
	}
	b.invuln, b.Health = 0, 5
	if !r.pvpDamage(a, b, 10, false) || b.Wallet != 85-43 || !b.dead || a.Kills != 1 { // ceil(42.5) = 43
		t.Fatalf("after the kill: wallet %d dead %v kills %d", b.Wallet, b.dead, a.Kills)
	}
	if got := coinsInPlay(r); got != 95 {
		t.Fatalf("%d coins in play, want 95", got)
	}
}

func TestPartyPreviousOwnerWaitsOneSecond(t *testing.T) {
	r := partyRoom(t, 2, true)
	a, b := r.players[0], r.players[1]
	r.party.bundles = nil
	stand(r, a, 120) // away from the safe's spray
	stand(r, b, 1150)
	a.Wallet = 50
	r.dropMoney(a, 20)
	bn := r.party.bundles[0]
	bn.vx, bn.vy, bn.X, bn.Y = 0, 0, a.X, a.Y
	partySteps(r, 10)
	if a.Wallet != 30 {
		t.Fatal("previous owner picked their coins back up at once")
	}
	partySteps(r, bundleLock)
	if a.Wallet != 50 {
		t.Fatalf("wallet %d, want 50 after the lock", a.Wallet)
	}
}

func TestPartyATMDeposit(t *testing.T) {
	r := partyRoom(t, 2, true)
	a, b := r.players[0], r.players[1]
	r.party.bundles = nil
	atm := r.party.atms[0]
	stand(r, a, atm.X)
	stand(r, b, atm.X+10)
	a.Wallet, b.Wallet = 70, 40
	a.in.Interact, b.in.Interact = true, true
	partySteps(r, 1)
	if atm.user != a.ID || b.atm != 0 {
		t.Fatal("ATM must serve one player at a time")
	}
	partySteps(r, atmTicks/2)
	a.in.Right = true // moving cancels, nothing lost
	partySteps(r, 1)
	if a.atm != 0 || a.Wallet != 70 || a.Balance != 0 || a.AtmCancels != 1 {
		t.Fatalf("after moving: atm %d wallet %d balance %d", a.atm, a.Wallet, a.Balance)
	}
	stand(r, a, atm.X)
	b.in.Interact = false
	a.in.Interact = true
	partySteps(r, atmTicks+2)
	if a.Wallet != 0 || a.Balance != 70 || a.Deposits != 1 {
		t.Fatalf("wallet %d balance %d deposits %d", a.Wallet, a.Balance, a.Deposits)
	}
}

func TestPartyDepositOnTheLastTickCounts(t *testing.T) {
	r := partyRoom(t, 1, true)
	a := r.players[0]
	r.party.bundles = nil
	atm := r.party.atms[1]
	partySteps(r, partyTicks-atmTicks-1)
	r.party.bundles = nil
	stand(r, a, atm.X)
	a.Wallet, a.in.Interact = 30, true
	partySteps(r, atmTicks+1)
	if !r.party.over || a.Balance != 30 {
		t.Fatalf("over %v balance %d: a deposit finished at 180 s must count", r.party.over, a.Balance)
	}
}

func TestPartyRespawnFarthestAndProtection(t *testing.T) {
	r := partyRoom(t, 2, true)
	a, b := r.players[0], r.players[1]
	r.party.bundles = nil
	stand(r, a, 160) // near the left floor respawn point
	stand(r, b, 300)
	b.Health = 1
	r.pvpDamage(a, b, 10, false)
	partySteps(r, pvpRespawnTicks)
	if b.dead || b.Health != pvpHP || b.protect == 0 || b.X < 640 {
		t.Fatalf("respawn: dead %v hp %v protect %d x %v (want far from the killer)", b.dead, b.Health, b.protect, b.X)
	}
	if r.pvpDamage(a, b, 10, false) {
		t.Fatal("hit during spawn protection")
	}
	b.pressed.attack = true // attacking ends protection
	partySteps(r, 1)
	if b.protect != 0 {
		t.Fatal("attacking did not end spawn protection")
	}
}

func TestPartyOfflinePlayerLeavesAfterFiveSeconds(t *testing.T) {
	r := partyRoom(t, 2, true)
	a, b := r.players[0], r.players[1]
	r.party.bundles = nil
	stand(r, a, 300)
	stand(r, b, 320)
	b.Wallet, b.Balance = 60, 25
	c := b.client
	b.client = nil
	partySteps(r, 30)
	if !r.pvpDamage(a, b, 10, false) || b.Wallet != 54 {
		t.Fatal("an offline player must stay hittable for 5 s")
	}
	partySteps(r, offlineTicks)
	if !b.left || b.Wallet != 0 || b.Balance != 25 || coinsInPlay(r) != 85+r.party.sprayed {
		t.Fatalf("left %v wallet %d balance %d coins %d", b.left, b.Wallet, b.Balance, coinsInPlay(r))
	}
	b.client = c // back after leaving: respawn, empty wallet, balance kept
	partySteps(r, 1)
	if b.left || b.Health != pvpHP || b.Balance != 25 {
		t.Fatalf("after coming back: left %v hp %v balance %d", b.left, b.Health, b.Balance)
	}
}

func TestPartyP5GateIsAsymmetric(t *testing.T) {
	r := partyRoom(t, 5, true)
	o := r.players[4]
	stand(r, o, 600)
	o.pressed.special = true
	partySteps(r, 1)
	if o.dashCD < pvpCardDashLock-2 {
		t.Fatalf("dash locked %d ticks after the skill, want ~%d", o.dashCD, pvpCardDashLock)
	}
	o.skillCD, o.dashCD = 0, 0
	partySteps(r, 20)
	o.pressed.dash = true
	partySteps(r, 1)
	if o.skillCD < pvpDashCardLock-2 || o.skillCD > pvpDashCardLock {
		t.Fatalf("skill locked %d ticks after a dash, want ~%d", o.skillCD, pvpDashCardLock)
	}
}

func TestPartySwingHitsEachOnceSmashHitsOne(t *testing.T) {
	r := partyRoom(t, 3, true)
	a, b, c := r.players[0], r.players[1], r.players[2]
	r.party.bundles = nil
	stand(r, a, 600)
	stand(r, b, 620)
	stand(r, c, 625)
	a.Facing = 1
	a.pressed.attack = true
	partySteps(r, attackTicks+2)
	if b.Health != pvpHP-pvpSwing || c.Health != pvpHP-pvpSwing {
		t.Fatalf("swing: b %v c %v, want one hit each", b.Health, c.Health)
	}
	b.invuln, c.invuln = 0, 0
	stand(r, b, 620)
	stand(r, c, 640)
	a.Facing = 1
	a.pressed.special, a.in.Special = true, true // P1: hold to the cap
	partySteps(r, chargeMaxTicks+castTicks+2)
	if b.Health != pvpHP-pvpSwing-pvpSmashMax || c.Health != pvpHP-pvpSwing {
		t.Fatalf("smash: b %v c %v, want only the nearest hit for %v", b.Health, c.Health, pvpSmashMax)
	}
}

func TestPartyShieldReflectsPulseToShooter(t *testing.T) {
	r := partyRoom(t, 2, true)
	a, v := r.players[0], r.players[1] // P1 shoots, P2 shields
	r.party.bundles = nil
	stand(r, a, 400)
	stand(r, v, 520)
	a.Facing, v.Facing = 1, -1
	v.pressed.special = true
	partySteps(r, 1)
	a.pressed.skill = true
	partySteps(r, 40)
	if v.Reflects != 1 || v.Health != pvpHP || a.Health != pvpHP-pvpReflected {
		t.Fatalf("reflects %d, shield hp %v, shooter hp %v", v.Reflects, v.Health, a.Health)
	}
}

func TestPartyMarkHelpsOnlyItsOwner(t *testing.T) {
	r := partyRoom(t, 3, true)
	a, b, amber := r.players[0], r.players[1], r.players[2]
	amber.markOn, amber.markTicks = b.ID, markTicks
	r.pvpDamage(amber, b, 10, true)
	if !near(pvpHP-b.Health, 10*markBonus) {
		t.Fatalf("P3 on its mark: %v, want %v", pvpHP-b.Health, 10*markBonus)
	}
	b.invuln = 0
	before := b.Health
	r.pvpDamage(a, b, 10, true)
	if !near(before-b.Health, 10) {
		t.Fatalf("someone else on the mark: %v, want 10", before-b.Health)
	}
}

func TestPartyHotfixHealsOnlyP4UpTo20(t *testing.T) {
	r := partyRoom(t, 4, true)
	mint, b := r.players[3], r.players[1]
	r.party.bundles = nil
	for _, p := range r.players {
		stand(r, p, 40+float64(p.ID)*2) // everyone else out of the zone except b
	}
	stand(r, mint, 600)
	stand(r, b, 620)
	mint.Health, b.Health = 50, 50
	mint.pressed.special = true
	partySteps(r, hotfixTicks+5)
	if !near(mint.Health, 70) || !near(b.Health, 50-pvpHotfix) {
		t.Fatalf("P4 %v (want 70), rival %v (want %v)", mint.Health, b.Health, 50-pvpHotfix)
	}
}

func TestPartyCoinsConservedInARandomBrawl(t *testing.T) {
	r := partyRoom(t, 5, true)
	rng := rand.New(rand.NewPCG(1, 2))
	for tick := range partyTicks {
		for _, p := range r.players {
			p.in = Input{Left: rng.IntN(4) == 0, Right: rng.IntN(4) == 0, Interact: rng.IntN(3) == 0, Special: rng.IntN(2) == 0}
			p.pressed = presses{jump: rng.IntN(20) == 0, attack: rng.IntN(6) == 0, skill: rng.IntN(15) == 0,
				dash: rng.IntN(30) == 0, special: rng.IntN(40) == 0}
		}
		if tick%900 == 300 { // someone drops out for a while
			r.players[tick/900%5].client = nil
		}
		if tick%900 == 700 {
			for _, p := range r.players {
				if p.client == nil {
					p.client = &client{send: make(chan []byte, 1<<12), cancel: func() {}, player: p}
				}
			}
		}
		r.stepParty()
		if got := coinsInPlay(r); got != r.party.sprayed {
			t.Fatalf("tick %d: %d coins in play, %d sprayed", tick, got, r.party.sprayed)
		}
		if r.party.over {
			break
		}
	}
	if !r.party.over {
		t.Fatal("party never ended")
	}
}

func TestBossWinLeadsToPartyWithConnectedPlayers(t *testing.T) {
	r := testRoom(t, 2, 1)
	r.partyLevel = partyLevelForTest(t)
	r.boss.HP = 1
	r.damageBoss(1, false)
	r.stepMatch()
	steps(r, wonTicks)
	if r.mode != modeParty || len(r.players) != 2 || r.level != r.partyLevel || r.party.n != 2 {
		t.Fatalf("mode %q players %d", r.mode, len(r.players))
	}
	if st := r.startMsg(); st.Level != partyLevelURL || st.Mode != modeParty {
		t.Fatalf("reconnect start message %+v", st)
	}
}

func TestPartyResultsShareRanks(t *testing.T) {
	r := partyRoom(t, 3, true)
	r.players[0].Balance, r.players[1].Balance, r.players[2].Balance = 80, 80, 0
	res := r.results()
	if res[0].Rank != 1 || res[1].Rank != 1 || res[2].Rank != 3 {
		t.Fatalf("ranks %+v", res)
	}
}

func TestHostStartsPartyFromLobby(t *testing.T) {
	r := testRoom(t, 0, 0)
	r.partyLevel = partyLevelForTest(t)
	r.phase = phaseLobby
	p := &Player{ID: 1, Name: "An"}
	c := &client{send: make(chan []byte, 64), cancel: func() {}, player: p}
	p.client = c
	r.players, r.hostID = []*Player{p}, 1
	r.handle(c, clientMsg{Type: "start", Level: modeParty})
	if r.mode != modeParty || r.phase != phasePlaying || r.party.n != 1 {
		t.Fatalf("mode %q phase %q", r.mode, r.phase)
	}
}
