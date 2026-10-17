package main

import (
	"math"
	"testing"

	"github.com/jakecoffman/cp/v2"
)

func run(w *World, n int) {
	for range n {
		w.Step()
	}
}

func feetY(p *Player) float64 { return p.body.Position().Y + p.h/2 }

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %.1f, want %.1f ± %.1f", name, got, want, tol)
	}
}

func TestPlayerLandsAndJumps(t *testing.T) {
	w := NewWorld()
	p := w.AddPlayer()
	run(w, 120)
	near(t, "feet on floor", feetY(p), 680, 3)

	p.in.Jump = true
	run(w, 1)
	if p.body.Velocity().Y >= 0 {
		t.Fatalf("jump did not launch: vy = %.1f", p.body.Velocity().Y)
	}
	p.in.Jump = false
	run(w, 120)
	near(t, "feet back on floor", feetY(p), 680, 3)
}

func TestOneWayPassUpLandOnTopDropDown(t *testing.T) {
	w := NewWorld()
	for _, it := range w.items { // seeded items rest on the platform and would block the jump
		w.removeItem(it)
	}
	w.items = nil
	p := w.AddPlayer()
	p.body.SetPosition(cp.Vector{X: 800, Y: 680 - p.h/2}) // under the one-way at y=540
	run(w, 30)
	p.in.Jump = true
	run(w, 1)
	p.in.Jump = false
	run(w, 90)
	near(t, "feet on one-way", feetY(p), 540, 3)

	p.in.Down = true
	run(w, 60)
	near(t, "feet after drop", feetY(p), 680, 3)
}

func TestMovingPlatformCarriesPlayer(t *testing.T) {
	w := NewWorld()
	p := w.AddPlayer()
	m := w.mover.Position()
	p.body.SetPosition(cp.Vector{X: m.X, Y: m.Y - w.moverH/2 - p.h/2 - 1})
	run(w, 10)
	startX := p.body.Position().X - w.mover.Position().X
	run(w, 30) // short: later the mover carries the rider into the fan, which knocks it off
	near(t, "offset from mover", p.body.Position().X-w.mover.Position().X, startX, 6)
}

func TestItemsSettleOnFloor(t *testing.T) {
	w := NewWorld()
	w.Reset()
	run(w, 600)
	for _, it := range w.items {
		if y := it.body.Position().Y; y > 680 || y < 0 {
			t.Errorf("%s %d out of bounds: y = %.1f", it.kind, it.id, y)
		}
	}
}

func TestRunSprintDashAndAirControl(t *testing.T) {
	w := NewWorld()
	p := w.AddPlayer()
	p.body.SetPosition(cp.Vector{X: 150, Y: 680 - p.h/2})
	run(w, 30)
	x0 := p.body.Position().X
	p.in.Right = true
	run(w, 30)
	near(t, "run distance in 0.5s", p.body.Position().X-x0, runSpeed*0.5, 25)

	p.in.Sprint = true
	run(w, 15)
	near(t, "sprint speed", p.body.Velocity().X, sprintSpeed, 5)

	p.in.Dash = true
	run(w, 2)
	near(t, "dash speed", p.body.Velocity().X, dashSpeed, 1)
	near(t, "dash keeps height", p.body.Velocity().Y, 0, 1)
}

func TestFanSpinsAndMoverPassesThrough(t *testing.T) {
	w := NewWorld()
	a0 := w.fan.Angle()
	minX, maxX := math.Inf(1), math.Inf(-1)
	for range 360 {
		w.Step()
		x := w.mover.Position().X
		minX, maxX = math.Min(minX, x), math.Max(maxX, x)
	}
	near(t, "fan turned in 6s", w.fan.Angle()-a0, fanOmega*6, 0.3)
	if maxX-minX < 250 {
		t.Errorf("mover stuck by the fan: x range %.0f..%.0f", minX, maxX)
	}
	near(t, "fan stays at centre", w.fan.Position().X, fanX, 1)
}

func TestCycleCharactersKeepsFeetAndHitbox(t *testing.T) {
	w := NewWorld()
	p := w.AddPlayer()
	for range len(chars) {
		w.setChar(p, p.char+1)
		run(w, 90)
		c := chars[p.char]
		if p.w != c.W || p.h != c.H {
			t.Errorf("%s hitbox = %.0fx%.0f, want %.0fx%.0f", c.Skin, p.w, p.h, c.W, c.H)
		}
		near(t, c.Skin+" feet on floor", feetY(p), 680, 3)
	}
	if p.char != startChar {
		t.Errorf("full cycle ended on %d, want %d", p.char, startChar)
	}
}
