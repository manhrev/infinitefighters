package main

import "math"

// "Em đang ở trên thì tiện nhận luôn nhé!": file stacks slam down on players standing on desks.
//
// Rules this file guarantees:
//   - Only after the boss drops below 60% HP, at most one turn every 12 s, and only in place of a
//     volley turn. With nobody on a desk, the boss throws the volley as before.
//   - Up to 3 columns, at most one per desk, so every desk keeps at least 2 tiles of free landing room
//     (desks are at least 4 tiles wide, a column is 2).
//   - Columns are locked at the start of the 1 s telegraph and never follow anyone.
//   - A column runs from just under the desk above it (or the ceiling) down to its target desk, so it
//     never hits anyone on the desk above or on a lower level.
//   - Each player is hit at most once per turn; stacks vanish on impact, leaving no lasting hazard.
//   - The paper jam afterwards is a 1.2 s window where the boss takes double damage.

const (
	bossDropWindup = "drop_windup"
	bossDrop       = "drop"
	bossDropJam    = "drop_jam"

	dropWindupTicks = tickRate // 1 s telegraph
	dropStagger     = 15       // 0.25 s between stacks
	dropJamTicks    = 72       // 1.2 s
	dropCooldown    = 12 * tickRate
	dropHPFrac      = 0.6
	dropMaxTargets  = 3
	dropW, dropH    = 64.0, 48.0
	dropSpeed       = 900.0
	deskUnderside   = 30.0 // desk top plus its apron: a stack under a desk starts below this
)

type Drop struct {
	ID          int
	X           float64 // centre of the column
	Top, Bottom float64 // the column: start of the fall, target desk top
	Y           float64 // bottom of the stack
	delay       int     // ticks after the telegraph before it falls
}

// lockDrops picks the columns at the start of the telegraph. It reports false when the attack
// is not available or nobody is standing on a desk.
func (r *Room) lockDrops() bool {
	if r.boss.HP >= dropHPFrac*r.boss.MaxHP || r.dropCD > 0 {
		return false
	}
	ts := float64(r.level.Tile)
	var drops []*Drop
	used := map[[2]int]bool{} // desk run: row, first column
	for _, p := range r.players {
		if len(drops) == dropMaxTargets {
			break
		}
		if p.client == nil || p.action == actDown {
			continue
		}
		row, c0, c1, ok := r.deskUnder(p)
		if !ok || used[[2]int{row, c0}] {
			continue
		}
		used[[2]int{row, c0}] = true
		x := math.Max(float64(c0)*ts+dropW/2, math.Min(float64(c1+1)*ts-dropW/2, p.X))
		top := r.coverAbove(x, row)
		r.nextShotID++
		drops = append(drops, &Drop{
			ID: r.nextShotID, X: x, Top: top, Bottom: float64(row) * ts, Y: top + dropH, delay: len(drops) * dropStagger,
		})
	}
	if len(drops) == 0 {
		return false
	}
	r.drops, r.dropHit, r.dropCD = drops, map[int]bool{}, dropCooldown
	return true
}

// deskUnder finds the one-way desk below a player's feet (standing or in the air above it).
// It reports false over the floor or a gap down to the floor.
func (r *Room) deskUnder(p *Player) (row, c0, c1 int, ok bool) {
	l, ts := r.level, float64(r.level.Tile)
	col := cell(p.X, ts)
	for row = cell(p.Y-eps, ts); row < l.H; row++ {
		switch l.at(col, row) {
		case tileSolid:
			return 0, 0, 0, false
		case tileOneWay:
			c0, c1 = col, col
			for l.at(c0-1, row) == tileOneWay {
				c0--
			}
			for l.at(c1+1, row) == tileOneWay {
				c1++
			}
			return row, c0, c1, true
		}
	}
	return 0, 0, 0, false
}

// coverAbove is where a column starts: under the lowest desk above it, or the top of the map.
func (r *Room) coverAbove(x float64, row int) float64 {
	l, ts := r.level, float64(r.level.Tile)
	top := 0.0
	for c := cell(x-dropW/2, ts); c <= cell(x+dropW/2-eps, ts); c++ {
		for rr := row - 1; rr >= 0; rr-- {
			if l.at(c, rr) != tileEmpty {
				top = math.Max(top, float64(rr)*ts+deskUnderside)
				break
			}
		}
	}
	return top
}

// stepDrops moves the falling stacks. It reports true when all of them have landed.
func (r *Room) stepDrops() bool {
	r.drops = keep(r.drops, func(d *Drop) bool {
		if d.delay > 0 {
			d.delay--
			return true
		}
		d.Y = math.Min(d.Bottom, d.Y+dropSpeed*dt)
		box := rect{d.X - dropW/2, d.Y - dropH, d.X + dropW/2, d.Y}
		for _, p := range r.players {
			if !r.dropHit[p.ID] && r.canHurt(p) && overlaps(box, playerBox(p)) {
				r.dropHit[p.ID] = true
				r.hurt(p, 1, d.X)
			}
		}
		return d.Y < d.Bottom // breaks on the target desk
	})
	return len(r.drops) == 0
}
