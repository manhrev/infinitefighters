package main

import "testing"

// asciiLevel: '#' solid, '-' one-way, anything else empty. Tiles are 32 px.
func asciiLevel(rows ...string) *Level {
	l := &Level{W: len(rows[0]), H: len(rows), Tile: 32, tiles: make([]tileKind, len(rows)*len(rows[0]))}
	for r, row := range rows {
		for c, ch := range row {
			switch ch {
			case '#':
				l.tiles[r*l.W+c] = tileSolid
			case '-':
				l.tiles[r*l.W+c] = tileOneWay
			}
		}
	}
	return l
}

func run(b *Body, l *Level, ticks int, in Input, jumpFirst bool) {
	for i := range ticks {
		b.step(l, in, jumpFirst && i == 0)
	}
}

// Floor at row 7 (top y=224), one-way platform at row 3 (top y=96), wall at col 9.
var testLevel = asciiLevel(
	"#........#",
	"#........#",
	"#........#",
	"#.----...#",
	"#........#",
	"#........#",
	"#........#",
	"##########",
)

func TestFallsAndLandsOnFloor(t *testing.T) {
	b := &Body{X: 230, Y: 10}
	run(b, testLevel, 120, Input{}, false)
	if b.Y != 224 || !b.OnGround || b.onOneWay {
		t.Fatalf("got Y=%v ground=%v oneWay=%v, want 224 on solid ground", b.Y, b.OnGround, b.onOneWay)
	}
}

func TestJumpUpThroughOneWayAndLand(t *testing.T) {
	b := &Body{X: 100, Y: 224, OnGround: true}
	run(b, testLevel, 120, Input{}, true)
	if b.Y != 96 || !b.onOneWay {
		t.Fatalf("got Y=%v oneWay=%v, want standing on platform at 96", b.Y, b.onOneWay)
	}
}

func TestDropThroughOneWay(t *testing.T) {
	b := &Body{X: 100, Y: 96}
	run(b, testLevel, 1, Input{}, false) // settle: sets onOneWay
	run(b, testLevel, 60, Input{Down: true}, true)
	if b.Y != 224 {
		t.Fatalf("got Y=%v, want 224 after dropping to the floor", b.Y)
	}
}

func TestDownAloneDoesNotDrop(t *testing.T) {
	b := &Body{X: 100, Y: 96}
	run(b, testLevel, 60, Input{Down: true}, false)
	if b.Y != 96 {
		t.Fatalf("got Y=%v, want to stay on platform at 96", b.Y)
	}
}

func TestWallStopsRun(t *testing.T) {
	b := &Body{X: 230, Y: 224}
	run(b, testLevel, 120, Input{Right: true}, false)
	if want := 9*32 - hitW/2; b.X != want {
		t.Fatalf("got X=%v, want %v against the wall", b.X, want)
	}
	run(b, testLevel, 120, Input{Left: true}, false)
	if want := 32 + hitW/2; b.X != want {
		t.Fatalf("got X=%v, want %v against the left wall", b.X, want)
	}
}

func TestJumpClearsFourTileStep(t *testing.T) {
	l := asciiLevel(
		"#........#", "#........#", "#........#", "#........#", "#........#",
		"#........#", "#........#", "#........#", "##########",
	)
	b := &Body{X: 100, Y: 256, OnGround: true}
	apex := b.Y
	for i := range 60 {
		b.step(l, Input{}, i == 0)
		apex = min(apex, b.Y)
	}
	if rise := 256 - apex; rise < 4*32+16 {
		t.Fatalf("jump rises %v px, want at least 144", rise)
	}
}

func TestHeadBumpsSolidCeiling(t *testing.T) {
	l := asciiLevel("##########", "#........#", "#........#", "##########")
	b := &Body{X: 100, Y: 96, OnGround: true}
	for i := range 30 {
		b.step(l, Input{}, i == 0)
		if b.Y < 32+hitH {
			t.Fatalf("tick %d: head at %v went into the ceiling", i, b.Y-hitH)
		}
	}
}
