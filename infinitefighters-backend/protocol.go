package main

import "encoding/json"

// Wire format: one JSON object per WebSocket text message, with a "type" field.

// clientMsg is any client → server message.
//
//	{"type":"join","name":"An","code":"ABCD"}       new player (lobby only)
//	{"type":"join","token":"..."}                   resume a kept slot
//	{"type":"start"}                                host only, lobby only
//	{"type":"input","left":true,"jump":false,...}   held buttons; send on every change
type clientMsg struct {
	Type  string `json:"type"`
	Name  string `json:"name,omitempty"`
	Code  string `json:"code,omitempty"`
	Token string `json:"token,omitempty"`
	Input
}

// Input is the set of held buttons. Drop-through is down + jump on a one-way platform.
type Input struct {
	Left  bool `json:"left"`
	Right bool `json:"right"`
	Down  bool `json:"down"`
	Jump  bool `json:"jump"`
}

// Server → client messages. Rejections are sent as a WebSocket close
// (status 1008, policy violation) with the reason as the close text.

type welcomeMsg struct {
	Type  string `json:"type"` // "welcome"
	ID    int    `json:"id"`
	Token string `json:"token"` // keep in localStorage to resume after a drop
	Phase string `json:"phase"`
}

type lobbyMsg struct {
	Type    string        `json:"type"` // "lobby"
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
	Players    []playerState `json:"players"`
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
}

func encode(v any) []byte {
	b, _ := json.Marshal(v) // only plain structs above; cannot fail
	return b
}
