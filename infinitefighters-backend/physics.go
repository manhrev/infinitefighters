package main

import "math"

// Tuning. Speeds in px/s, +Y is down. Fixed step, never wall-clock deltas.
const (
	dt        = 1.0 / tickRate
	gravity   = 2400.0
	jumpSpeed = 880.0 // apex = jumpSpeed²/(2·gravity) ≈ 161 px, clears a 4-tile (128 px) step
	runSpeed  = 260.0
	maxFall   = 900.0 // 15 px per tick, under one tile, so bodies cannot tunnel through floors
	dropTicks = 10    // one-way platforms are ignored this long after a drop-through
	hitW      = 20.0
	hitH      = 44.0
	eps       = 1e-6
)

// Body is a player's physics state. (X, Y) is the feet: bottom-centre of the hitbox.
type Body struct {
	X, Y      float64
	VX, VY    float64
	Facing    int
	OnGround  bool
	onOneWay  bool // standing only on one-way tiles, so drop-through is allowed
	dropTicks int
	dashTicks int     // > 0: dashing forward, no gravity
	dashSpeed float64 // dash and P5 timecard differ
	slow      bool    // P1 charging: walks at chargeSlow
	pushVX    float64 // knockback overrides input while pushTicks > 0
	pushTicks int
}

// step advances one tick. jump is the latched press edge, not the held button.
func (b *Body) step(l *Level, in Input, jump bool) {
	if b.dashTicks > 0 {
		b.dashTicks--
		b.VX, b.VY = float64(b.Facing)*b.dashSpeed, 0
		b.moveX(l, b.VX*dt)
		if b.VX == 0 {
			b.dashTicks = 0 // hit a wall
		}
		return
	}
	dir := 0
	if in.Left {
		dir--
	}
	if in.Right {
		dir++
	}
	b.VX = float64(dir) * runSpeed
	if b.slow {
		b.VX *= chargeSlow
	}
	if dir != 0 {
		b.Facing = dir
	}
	if b.pushTicks > 0 {
		b.pushTicks--
		b.VX = b.pushVX
	}
	if jump && b.OnGround {
		if in.Down && b.onOneWay {
			b.dropTicks = dropTicks
		} else {
			b.VY = -jumpSpeed
		}
	}
	if b.dropTicks > 0 {
		b.dropTicks--
	}
	b.VY = min(b.VY+gravity*dt, maxFall)
	b.moveX(l, b.VX*dt)
	b.moveY(l, b.VY*dt)
}

// moveX checks only the leading edge column: one tick moves less than one tile.
func (b *Body) moveX(l *Level, dx float64) {
	if dx == 0 {
		return
	}
	b.X += dx
	ts := float64(l.Tile)
	top, bottom := cell(b.Y-hitH, ts), cell(b.Y-eps, ts)
	if dx > 0 {
		col := cell(b.X+hitW/2-eps, ts)
		for row := top; row <= bottom; row++ {
			if l.at(col, row) == tileSolid {
				b.X, b.VX = float64(col)*ts-hitW/2, 0
				return
			}
		}
	} else {
		col := cell(b.X-hitW/2, ts)
		for row := top; row <= bottom; row++ {
			if l.at(col, row) == tileSolid {
				b.X, b.VX = float64(col+1)*ts+hitW/2, 0
				return
			}
		}
	}
}

func (b *Body) moveY(l *Level, dy float64) {
	prev := b.Y
	b.Y += dy
	b.OnGround, b.onOneWay = false, false
	ts := float64(l.Tile)
	left, right := cell(b.X-hitW/2, ts), cell(b.X+hitW/2-eps, ts)
	if dy > 0 {
		row := cell(b.Y-eps, ts)
		top := float64(row) * ts
		solid, oneWay := false, false
		for col := left; col <= right; col++ {
			switch l.at(col, row) {
			case tileSolid:
				solid = true
			case tileOneWay:
				// Land only when the feet were above the platform top last tick.
				oneWay = oneWay || (b.dropTicks == 0 && prev <= top+eps)
			}
		}
		if solid || oneWay {
			b.Y, b.VY, b.OnGround, b.onOneWay = top, 0, true, !solid
		}
		return
	}
	row := cell(b.Y-hitH, ts)
	for col := left; col <= right; col++ {
		if l.at(col, row) == tileSolid { // one-way tiles never block from below
			b.Y, b.VY = float64(row+1)*ts+hitH, 0
			return
		}
	}
}

func cell(px, ts float64) int { return int(math.Floor(px / ts)) }
