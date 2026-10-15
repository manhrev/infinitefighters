package main

import "encoding/json"

// Wire format: one JSON object per WebSocket text message, with a "type" field.

// clientMsg is any client → server message.
//
//	{"type":"join","name":"An","code":"ABCD"}       new player (lobby only)
//	{"type":"join","token":"..."}                   resume a kept slot
//	{"type":"start"}                                host only, lobby only
//	{"type":"input","left":true,"jump":false,...}   held buttons; send on every change
//
// Presses (jump, attack, skill, special, dash, ping) act on the up→down edge; interact acts while held,
// and P1's special charges while held.
type clientMsg struct {
	Type  string `json:"type"`
	Name  string `json:"name,omitempty"`
	Code  string `json:"code,omitempty"`
	Token string `json:"token,omitempty"`
	Input
}

// Input is the set of held buttons. Drop-through is down + jump on a one-way platform.
type Input struct {
	Left     bool `json:"left"`
	Right    bool `json:"right"`
	Down     bool `json:"down"`
	Jump     bool `json:"jump"`
	Attack   bool `json:"attack"`   // keyboard swing
	Skill    bool `json:"skill"`    // code pulse
	Dash     bool `json:"dash"`     // invulnerable while dashing
	Special  bool `json:"special"`  // the slot's own skill; P1 charges while held
	Interact bool `json:"interact"` // hold next to a downed teammate to revive
	Ping     int  `json:"ping"`     // 0 none, 1 Tôi nhận, 2 Cần hỗ trợ, 3 Chặn scope
}

// Server → client messages. Rejections are sent as a WebSocket close
// (status 1008, policy violation) with the reason as the close text.

type welcomeMsg struct {
	Type  string `json:"type"` // "welcome"
	ID    int    `json:"id"`
	Token string `json:"token"` // keep in localStorage to resume after a drop
	Phase string `json:"phase"`
}

// lobbyMsg is sent on every roster change. Phase "lobby" after a match means the match ended.
type lobbyMsg struct {
	Type    string        `json:"type"` // "lobby"
	Phase   string        `json:"phase"`
	Players []lobbyPlayer `json:"players"`
}

type lobbyPlayer struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Slot      int    `json:"slot"` // 0..4 = P1..P5 colour
	Host      bool   `json:"host"`
	Connected bool   `json:"connected"`
}

type startMsg struct {
	Type  string `json:"type"`  // "start"
	Level string `json:"level"` // URL of the Tiled map
}

type snapshotMsg struct {
	Type       string        `json:"type"` // "snapshot"
	Tick       uint64        `json:"tick"` // 60 per second of match time
	ServerTime int64         `json:"serverTime"`
	DeadlineMs int64         `json:"deadlineMs"` // time left on the DEADLINE bar
	DeadlineOf int64         `json:"deadlineOf"` // full DEADLINE, ms
	Result     string        `json:"result"`     // "", "won", "lost"; frozen a few seconds, then restart or lobby
	Players    []playerState `json:"players"`
	Boss       bossState     `json:"boss"`
	Shots      []shotState   `json:"shots"`
	Waves      []waveState   `json:"waves"` // deadline waves along the floor
	Items      []itemState   `json:"items"` // pickups currently available
	Zones      []zoneState   `json:"zones"` // P4 hotfix zones
	Drops      []dropState   `json:"drops"` // falling file stacks, shown from the start of drop_windup
}

// playerState X/Y are the feet (bottom-centre of the 20×44 hitbox) in level pixels, +Y down.
type playerState struct {
	ID     int     `json:"id"`
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	VX     float64 `json:"vx"`
	VY     float64 `json:"vy"`
	Facing int     `json:"facing"` // 1 right, -1 left
	Ground bool    `json:"ground"`
	Away   bool    `json:"away"` // disconnected: frozen and invulnerable
	HP     int     `json:"hp"`
	// "", attack, pulse, dash, hurt, down, revive, or a slot skill: charge, smash, shield, stamp, hotfix, timecard
	Action  string  `json:"action"`
	Revive  float64 `json:"revive"` // 0..1 progress while a teammate revives this downed player
	Ping    int     `json:"ping"`
	SkillMs int64   `json:"skillMs"` // slot skill cooldown left
	SkillOf int64   `json:"skillOf"` // full cooldown
	Charge  float64 `json:"charge"`  // P1: 0..1 while charging
	// Counters only grow, across restarts too; play the one-shot animation or sound when one changes.
	Attacks   uint32 `json:"attacks"`
	Pulses    uint32 `json:"pulses"`
	Dashes    uint32 `json:"dashes"`
	Hurts     uint32 `json:"hurts"`
	Hits      uint32 `json:"hits"`      // keyboard swing or P1 smash landed
	Casts     uint32 `json:"casts"`     // slot skill used (P1: released)
	SkillHits uint32 `json:"skillHits"` // P1 smash, P4 deploy pulse or P5 timecard hit the boss
	Cancels   uint32 `json:"cancels"`   // P1 charge cancelled
	Reflects  uint32 `json:"reflects"`  // P2 bounced a task card
	Healed    uint32 `json:"healed"`    // this player got a heart from a hotfix zone
	Coffees   uint32 `json:"coffees"`
}

type bossState struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"` // feet, hitbox 70×110
	Facing int     `json:"facing"`
	HP     float64 `json:"hp"`
	MaxHP  float64 `json:"maxHp"`
	// idle, walk, volley_windup, volley, wave_windup, wave, charge_windup, charge, charge_recover,
	// drop_windup, drop, drop_jam, wind_clock, defeat
	State string `json:"state"`
	// Counters reset with each attempt: treat a decrease as a reset, not an event.
	Hits   uint32 `json:"hits"`  // flash and spark when it grows
	Marks  uint32 `json:"marks"` // P3 mark applied
	MarkMs int64  `json:"markMs"`
	// Locked charge corridor (floor x range), set during charge_windup and charge.
	ChargeFrom float64 `json:"chargeFrom,omitempty"`
	ChargeTo   float64 `json:"chargeTo,omitempty"`
}

type shotState struct {
	ID   int     `json:"id"`
	Kind string  `json:"kind"` // card (boss task), pulse (code pulse), stamp (P3), returned (card bounced by P2)
	X    float64 `json:"x"`    // centre
	Y    float64 `json:"y"`
	VX   float64 `json:"vx"`
}

type waveState struct {
	ID  int     `json:"id"`
	X   float64 `json:"x"` // centre; the wave is 32×32 standing on the floor
	Dir int     `json:"dir"`
}

// dropState is one locked column of the drop attack: warn from Top to Bottom until the stack falls.
type dropState struct {
	ID     int     `json:"id"`
	X      float64 `json:"x"`      // centre; the column is 64 px wide
	Top    float64 `json:"top"`    // the fall starts here (under the desk above, or the ceiling)
	Bottom float64 `json:"bottom"` // target desk top: the stack breaks here
	Y      float64 `json:"y"`      // bottom of the 64×48 stack
	Ms     int64   `json:"ms"`     // time until it starts falling; 0 while falling
}

type zoneState struct {
	ID int     `json:"id"`
	X  float64 `json:"x"` // bottom-centre of a 128×48 area
	Y  float64 `json:"y"`
	Ms int64   `json:"ms"` // time left
}

type itemState struct {
	ID   int     `json:"id"`
	Kind string  `json:"kind"` // "coffee": heals 1 HP
	X    float64 `json:"x"`    // bottom-centre
	Y    float64 `json:"y"`
}

func encode(v any) []byte {
	b, _ := json.Marshal(v) // only plain structs above; cannot fail
	return b
}
