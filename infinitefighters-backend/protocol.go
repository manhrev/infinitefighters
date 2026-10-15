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
// Presses (jump, attack, skill, dash, ping) act on the up→down edge; interact acts while held.
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
	Action string  `json:"action"` // "", attack, pulse, dash, hurt, down, revive
	Revive float64 `json:"revive"` // 0..1 progress while a teammate revives this downed player
	Ping   int     `json:"ping"`
	// Counters only grow; play the one-shot animation when one changes.
	Attacks uint32 `json:"attacks"`
	Pulses  uint32 `json:"pulses"`
	Dashes  uint32 `json:"dashes"`
	Hurts   uint32 `json:"hurts"`
}

type bossState struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"` // feet, hitbox 70×110
	Facing int     `json:"facing"`
	HP     int     `json:"hp"`
	MaxHP  int     `json:"maxHp"`
	State  string  `json:"state"` // idle, volley_windup, volley, wave_windup, wave, wind_clock, defeat
	Hits   uint32  `json:"hits"`  // counter: flash and spark when it changes
}

type shotState struct {
	ID   int     `json:"id"`
	Kind string  `json:"kind"` // "card" (boss task card) or "pulse" (player code pulse)
	X    float64 `json:"x"`    // centre
	Y    float64 `json:"y"`
	VX   float64 `json:"vx"`
}

type waveState struct {
	ID  int     `json:"id"`
	X   float64 `json:"x"` // centre; the wave is 32×32 standing on the floor
	Dir int     `json:"dir"`
}

func encode(v any) []byte {
	b, _ := json.Marshal(v) // only plain structs above; cannot fail
	return b
}
