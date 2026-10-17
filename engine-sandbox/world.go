package main

import (
	"math"

	"github.com/jakecoffman/cp/v2"
)

// Everything is a cp body. Tuning copied from infinitefighters-backend/physics.go so the feel matches. Speeds in px/s, +Y is down.
const (
	tickRate    = 60
	dt          = 1.0 / tickRate
	gravity     = 2400.0
	jumpSpeed   = 880.0
	runSpeed    = 260.0
	sprintSpeed = 400.0
	dashSpeed   = 700.0
	dashTicks   = 10
	dropTicks   = 10 // one-way platforms are ignored this long after pressing down
	maxFall     = 900.0
	worldW      = 1280.0
	worldH      = 720.0
	maxItems    = 80
	moverMass   = 40.0
	moverOmega  = 14.0   // mover drive: critically-damped spring, rad/s
	accelGround = 5000.0 // player: friction against the surface velocity gives this acceleration
	accelAir    = 2500.0
	fanX        = worldW / 2
	fanY        = worldH / 2
	fanLen      = 200.0 // tip to tip; two crossed bars make four blades
	fanThick    = 14.0
	fanOmega    = 3.0 // rad/s
)

const fanGroup = 1 // shapes in the same non-zero group never collide with each other

const (
	typePlayer cp.CollisionType = iota + 1
	typeItem
	typeOneWay
)

type Input struct {
	Left, Right, Down, Jump, Sprint, Dash bool
}

// Static is a rectangle of level geometry (top-left origin), sent once to clients.
type Static struct {
	X, Y, W, H float64
	OneWay     bool
}

type BodyState struct {
	ID     int
	Kind   string // player, circle, box, mover
	X, Y   float64
	A      float64
	W, H   float64 // box-like kinds
	R      float64 // circle
	Facing float64 // player only
	State  string  // player only: idle, run, jump, fall, dash
	Skin   string  // player only: sprite dir under assets/sprites, e.g. boss/khabib
}

// Char is a controllable sprite set with the hitbox its manifest entry specifies (anchor: bottom-center).
type Char struct {
	Skin string
	W, H float64
	Mass float64
}

// chars is the Q cycle order. Player sprites are 64px frames with a 20x44 hitbox, bosses 160px with 70x110.
var chars = []Char{
	{"player/teal", 20, 44, 3}, {"player/violet", 20, 44, 3}, {"player/amber", 20, 44, 3},
	{"player/mint", 20, 44, 3}, {"player/orange", 20, 44, 3},
	{"boss/doc-cong", 70, 110, 8}, {"boss/khabib", 70, 110, 8}, {"boss/vought", 70, 110, 8}, {"boss/yasuka", 70, 110, 8},
}

const startChar = 6 // boss/khabib

type Player struct {
	id       int
	char     int
	w, h     float64 // current hitbox
	body     *cp.Body
	shape    *cp.Shape
	in       Input
	prevJump bool
	prevDash bool
	facing   float64
	dash     int
	dashing  bool // this tick's dash state, read by the velocity update func
	drop     int
	grounded bool
	target   float64 // wanted horizontal speed
}

type Item struct {
	id   int
	kind string
	body *cp.Body
	w, h float64 // box size
	r    float64 // circle radius
}

type World struct {
	space   *cp.Space
	players map[int]*Player
	items   []*Item
	mover   *cp.Body
	fan     *cp.Body
	moverW  float64
	moverH  float64
	tick    int
	nextID  int
	Statics []Static
}

func NewWorld() *World {
	w := &World{space: cp.NewSpace(), players: map[int]*Player{}, nextID: 1}
	w.space.SetGravity(cp.Vector{Y: gravity})

	for _, s := range []Static{
		{0, 680, worldW, 40, false},        // floor
		{-40, -40, 40, worldH + 80, false}, // left wall
		{worldW, -40, 40, worldH + 80, false},
		{0, -40, worldW, 40, false},  // ceiling
		{60, 540, 200, 24, false},    // solid ledge
		{500, 600, 120, 80, false},   // solid step
		{700, 540, 220, 12, true},    // one-way
		{980, 420, 200, 12, true},    // one-way, higher
		{1100, 560, 120, 120, false}, // solid block, push items against it
	} {
		w.Statics = append(w.Statics, s)
		sh := cp.NewBox2(w.space.StaticBody, cp.BB{L: s.X, B: s.Y, R: s.X + s.W, T: s.Y + s.H}, 0)
		sh.SetFriction(1)
		if s.OneWay {
			sh.SetCollisionType(typeOneWay)
		}
		w.space.AddShape(sh)
	}

	// Moving platform: dynamic, locked to a horizontal groove and pulled along a path by a spring-damper force.
	w.moverW, w.moverH = 140, 20
	w.mover = w.space.AddBody(cp.NewBody(moverMass, cp.INFINITY))
	w.mover.SetVelocityUpdateFunc(func(b *cp.Body, _ cp.Vector, damping, dt float64) {
		cp.BodyUpdateVelocity(b, cp.Vector{}, damping, dt) // gravity off: the groove only locks Y
	})
	w.space.AddConstraint(cp.NewGrooveJoint(w.space.StaticBody, w.mover, cp.Vector{X: 0, Y: 420}, cp.Vector{X: worldW, Y: 420}, cp.Vector{}))
	w.mover.SetPosition(cp.Vector{X: 420, Y: 420})
	sh := w.space.AddShape(cp.NewBox(w.mover, w.moverW, w.moverH, 0))
	sh.SetFriction(1)
	sh.SetFilter(cp.NewShapeFilter(fanGroup, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))

	// Four-blade fan: a dynamic body on a pivot, spun by a motor strong enough that pushing it barely slows it.
	// Same filter group as the mover, so the two pass through each other while everything else collides.
	const fanMass = 30.0
	w.fan = w.space.AddBody(cp.NewBody(fanMass, 2*cp.MomentForBox(fanMass/2, fanLen, fanThick)))
	w.fan.SetPosition(cp.Vector{X: fanX, Y: fanY})
	w.fan.SetVelocityUpdateFunc(func(b *cp.Body, _ cp.Vector, damping, dt float64) {
		cp.BodyUpdateVelocity(b, cp.Vector{}, damping, dt) // the pivot already holds it; gravity only adds jitter
	})
	for _, size := range [][2]float64{{fanLen, fanThick}, {fanThick, fanLen}} {
		blade := w.space.AddShape(cp.NewBox(w.fan, size[0], size[1], 0))
		blade.SetFriction(0.8)
		blade.SetFilter(cp.NewShapeFilter(fanGroup, cp.ALL_CATEGORIES, cp.ALL_CATEGORIES))
	}
	w.space.AddConstraint(cp.NewPivotJoint2(w.space.StaticBody, w.fan, cp.Vector{X: fanX, Y: fanY}, cp.Vector{}))
	motor := w.space.AddConstraint(cp.NewSimpleMotor(w.space.StaticBody, w.fan, -fanOmega))
	motor.SetMaxForce(1e8)

	// One-way: only collide when the mobile shape is above the platform (normal points down in +Y-down).
	w.space.NewCollisionHandler(typePlayer, typeOneWay).PreSolveFunc = func(arb *cp.Arbiter, _ *cp.Space, _ any) bool {
		sh, _ := arb.Shapes()
		p := sh.UserData.(*Player)
		if p.drop > 0 || arb.Normal().Y < 0.5 {
			return arb.Ignore()
		}
		return true
	}
	w.space.NewCollisionHandler(typeItem, typeOneWay).PreSolveFunc = func(arb *cp.Arbiter, _ *cp.Space, _ any) bool {
		if arb.Normal().Y < 0.5 {
			return arb.Ignore()
		}
		return true
	}

	w.Reset()
	return w
}

// Reset clears items, respawns players and seeds a few items.
func (w *World) Reset() {
	for _, it := range w.items {
		w.removeItem(it)
	}
	w.items = nil
	for _, p := range w.players {
		w.spawnPlayerAt(p)
	}
	for i := range 4 {
		w.SpawnItem("circle", 760+float64(i)*40, 100)
		w.SpawnItem("box", 780+float64(i)*40, 40)
	}
}

func (w *World) AddPlayer() *Player {
	p := &Player{id: w.nextID, facing: 1}
	w.nextID++
	p.body = w.space.AddBody(cp.NewBody(3, cp.INFINITY)) // infinite moment: never rotates
	// Chipmunk platformer pattern: friction against the shape's surface velocity drives the player on the
	// ground (and carries it on moving platforms); in the air the velocity func steers toward the target.
	p.body.SetVelocityUpdateFunc(func(b *cp.Body, g cp.Vector, damping, dt float64) {
		if p.dashing {
			cp.BodyUpdateVelocity(b, cp.Vector{}, damping, dt)
			b.SetVelocity(p.facing*dashSpeed, 0)
			return
		}
		cp.BodyUpdateVelocity(b, g, damping, dt)
		v := b.Velocity()
		if !p.grounded {
			v.X = cp.LerpConst(v.X, p.target, accelAir*dt)
		}
		v.Y = math.Min(v.Y, maxFall)
		b.SetVelocityVector(v)
	})
	w.setChar(p, startChar)
	w.spawnPlayerAt(p)
	w.players[p.id] = p
	return p
}

// setChar swaps the player's hitbox and mass to chars[i], keeping the feet where they were.
func (w *World) setChar(p *Player, i int) {
	p.char = i % len(chars)
	c := chars[p.char]
	if p.shape != nil {
		pos := p.body.Position()
		w.space.RemoveShape(p.shape)
		p.body.SetPosition(cp.Vector{X: pos.X, Y: pos.Y + p.h/2 - c.H/2 - 1}) // -1: start just clear of the floor
	}
	p.w, p.h = c.W, c.H
	p.body.SetMass(c.Mass)
	p.shape = w.space.AddShape(cp.NewBox(p.body, c.W-4, c.H-4, 2)) // rounded corners slide over tile seams
	p.shape.SetCollisionType(typePlayer)
	p.shape.UserData = p
}

func (w *World) spawnPlayerAt(p *Player) {
	p.body.SetPosition(cp.Vector{X: 200 + float64(p.id%5)*30, Y: 600})
	p.body.SetVelocity(0, 0)
}

func (w *World) RemovePlayer(p *Player) {
	delete(w.players, p.id)
	p.body.EachShape(func(s *cp.Shape) { w.space.RemoveShape(s) })
	w.space.RemoveBody(p.body)
}

func (w *World) SpawnItem(kind string, x, y float64) {
	if len(w.items) >= maxItems {
		w.removeItem(w.items[0])
		w.items = w.items[1:]
	}
	it := &Item{id: w.nextID, kind: kind}
	w.nextID++
	var sh *cp.Shape
	if kind == "circle" {
		it.r = 16
		it.body = w.space.AddBody(cp.NewBody(1, cp.MomentForCircle(1, 0, it.r, cp.Vector{})))
		sh = cp.NewCircle(it.body, it.r, cp.Vector{})
		sh.SetElasticity(0.5)
		sh.SetFriction(0.6)
	} else {
		it.kind, it.w, it.h = "box", 32, 32
		mass := 2.0
		if kind == "boss" { // boss hitbox from assets/sprites/manifest.json, heavy so the player can barely push it
			it.kind, it.w, it.h, mass = "boss", 70, 110, 8
		}
		it.body = w.space.AddBody(cp.NewBody(mass, cp.MomentForBox(mass, it.w, it.h)))
		sh = cp.NewBox(it.body, it.w, it.h, 0)
		sh.SetElasticity(0.1)
		sh.SetFriction(0.8)
	}
	it.body.SetPosition(cp.Vector{X: x, Y: y})
	sh.SetCollisionType(typeItem)
	w.space.AddShape(sh)
	w.items = append(w.items, it)
}

func (w *World) removeItem(it *Item) {
	it.body.EachShape(func(s *cp.Shape) { w.space.RemoveShape(s) })
	w.space.RemoveBody(it.body)
}

// ground reports whether the player stands on something, and whether that is a one-way platform.
func (p *Player) ground() (ok, oneWay bool) {
	p.body.EachArbiter(func(a *cp.Arbiter) {
		if a.Count() == 0 || a.Normal().Y < 0.7 { // normal points from the player to the other shape
			return
		}
		_, other := a.Shapes()
		ok = true
		oneWay = other.CollisionType() == typeOneWay
	})
	return
}

func (w *World) stepPlayer(p *Player) {
	in := p.in
	ground, oneWay := p.ground()
	if p.drop > 0 {
		p.drop--
		ground = false // contacts with the ignored platform are not ground
	}
	if in.Down && ground && oneWay {
		p.drop = dropTicks
		ground = false
	}

	dir := 0.0
	if in.Left {
		dir--
	}
	if in.Right {
		dir++
	}
	if dir != 0 {
		p.facing = dir
	}
	speed := runSpeed
	if in.Sprint {
		speed = sprintSpeed
	}
	p.target = dir * speed
	p.grounded = ground

	if in.Dash && !p.prevDash && p.dash == 0 {
		p.dash = dashTicks
	}
	p.prevDash = in.Dash
	p.dashing = p.dash > 0
	if p.dashing {
		p.dash--
	}

	if ground && !p.dashing {
		p.shape.SetFriction(accelGround / gravity) // friction force / weight = acceleration / gravity
		p.shape.SetSurfaceV(cp.Vector{X: -p.target})
	} else {
		p.shape.SetFriction(0)
		p.shape.SetSurfaceV(cp.Vector{})
	}

	if in.Jump && !p.prevJump && ground && !p.dashing {
		p.body.ApplyImpulseAtLocalPoint(cp.Vector{Y: p.body.Mass() * (-jumpSpeed - p.body.Velocity().Y)}, cp.Vector{})
	}
	p.prevJump = in.Jump
}

func (p *Player) state() string {
	v := p.body.Velocity()
	switch {
	case p.dashing:
		return "dash"
	case !p.grounded && v.Y < 0:
		return "jump"
	case !p.grounded:
		return "fall"
	case math.Abs(v.X) > 20:
		return "run"
	}
	return "idle"
}

func (w *World) Step() {
	w.tick++
	targetX := 420 + 150*math.Sin(2*math.Pi*float64(w.tick)/(6*tickRate))
	pos, v := w.mover.Position(), w.mover.Velocity()
	w.mover.SetForce(cp.Vector{X: moverMass * (moverOmega*moverOmega*(targetX-pos.X) - 2*moverOmega*v.X)})
	for _, p := range w.players {
		w.stepPlayer(p)
	}
	w.space.Step(dt)
}

func r1(v float64) float64 { return math.Round(v*10) / 10 }

func (w *World) Snapshot() []BodyState {
	out := make([]BodyState, 0, len(w.players)+len(w.items)+1)
	pos := w.mover.Position()
	out = append(out, BodyState{Kind: "mover", X: r1(pos.X), Y: r1(pos.Y), W: w.moverW, H: w.moverH})
	fan := w.fan.Position()
	out = append(out, BodyState{Kind: "fan", X: r1(fan.X), Y: r1(fan.Y), A: math.Round(w.fan.Angle()*100) / 100, W: fanLen, H: fanThick})
	for _, p := range w.players {
		pos := p.body.Position()
		out = append(out, BodyState{ID: p.id, Kind: "player", X: r1(pos.X), Y: r1(pos.Y), W: p.w, H: p.h, Facing: p.facing, State: p.state(), Skin: chars[p.char].Skin})
	}
	for _, it := range w.items {
		pos := it.body.Position()
		out = append(out, BodyState{ID: it.id, Kind: it.kind, X: r1(pos.X), Y: r1(pos.Y), A: math.Round(it.body.Angle()*100) / 100, W: it.w, H: it.h, R: it.r})
	}
	return out
}
