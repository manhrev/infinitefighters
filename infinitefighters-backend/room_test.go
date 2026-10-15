package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
)

func TestSprintLevelParses(t *testing.T) {
	data, err := levelFS.ReadFile("levels/sprint.json")
	if err != nil {
		t.Fatal(err)
	}
	l, err := parseLevel(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Spawns) != maxPlayers {
		t.Fatalf("got %d spawns, want %d", len(l.Spawns), maxPlayers)
	}
	// Each desk level is reachable from the one below: the step is under the jump apex with margin.
	apex := jumpSpeed * jumpSpeed / (2 * gravity)
	below := l.Spawns[0].Y // floor
	for i := len(l.PlatformTops) - 1; i >= 0; i-- {
		if step := below - l.PlatformTops[i]; step <= 0 || step > apex-16 {
			t.Fatalf("desk level y=%v is %v px above the one below, jump apex %v", l.PlatformTops[i], step, apex)
		}
		below = l.PlatformTops[i]
	}
	for _, s := range l.Spawns { // every spawn stands on solid floor
		if l.at(cell(s.X, 32), cell(s.Y, 32)) != tileSolid {
			t.Fatalf("spawn %v is not on the floor", s)
		}
	}
}

// End-to-end over real WebSockets: join, start, move, drop, reconnect.
func TestRoomFlow(t *testing.T) {
	data, _ := levelFS.ReadFile("levels/sprint.json")
	level, err := parseLevel(data)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	room := newRoom("TEST", level)
	go room.run(ctx)
	srv := httptest.NewServer(newMux(room))
	defer srv.Close()
	url := "ws" + strings.TrimPrefix(srv.URL, "http") + "/ws"

	dial := func() *websocket.Conn {
		c, _, err := websocket.Dial(ctx, url, nil)
		if err != nil {
			t.Fatal(err)
		}
		return c
	}
	send := func(c *websocket.Conn, m clientMsg) {
		if err := c.Write(ctx, websocket.MessageText, encode(m)); err != nil {
			t.Fatal(err)
		}
	}
	// next reads until a message of the given type and decodes it into out.
	next := func(c *websocket.Conn, typ string, out any) {
		t.Helper()
		for {
			_, b, err := c.Read(ctx)
			if err != nil {
				t.Fatalf("waiting for %q: %v", typ, err)
			}
			var head struct{ Type string }
			json.Unmarshal(b, &head)
			if head.Type == typ {
				json.Unmarshal(b, out)
				return
			}
		}
	}
	rejected := func(c *websocket.Conn, reason string) {
		t.Helper()
		var err error
		for err == nil { // skip queued messages until the close arrives
			_, _, err = c.Read(ctx)
		}
		var ce websocket.CloseError
		if !errors.As(err, &ce) || ce.Reason != reason {
			t.Fatalf("got %v, want close %q", err, reason)
		}
	}
	findPlayer := func(s snapshotMsg, id int) playerState {
		for _, p := range s.Players {
			if p.ID == id {
				return p
			}
		}
		t.Fatalf("player %d missing from snapshot", id)
		return playerState{}
	}

	bad := dial()
	send(bad, clientMsg{Type: "join", Name: "X", Code: "NOPE"})
	rejected(bad, "wrong room code")

	a := dial()
	send(a, clientMsg{Type: "join", Name: "  An  ", Code: "test"})
	var wa welcomeMsg
	next(a, "welcome", &wa)

	b := dial()
	send(b, clientMsg{Type: "join", Name: "Bình", Code: "TEST"})
	var wb welcomeMsg
	next(b, "welcome", &wb)
	var lob lobbyMsg
	next(b, "lobby", &lob)
	if len(lob.Players) != 2 || !lob.Players[0].Host || lob.Players[0].Name != "An" || lob.Players[1].Slot != 1 {
		t.Fatalf("lobby = %+v", lob)
	}

	send(b, clientMsg{Type: "start"}) // not host: ignored
	send(a, clientMsg{Type: "start"})
	var st startMsg
	next(a, "start", &st)
	next(b, "start", &st)
	if st.Level != coopLevelURL {
		t.Fatalf("level = %q", st.Level)
	}

	late := dial()
	send(late, clientMsg{Type: "join", Name: "Late", Code: "TEST"})
	rejected(late, "match in progress")

	var snap snapshotMsg
	next(a, "snapshot", &snap)
	x0 := findPlayer(snap, wa.ID).X
	send(a, clientMsg{Type: "input", Input: Input{Right: true}})
	for findPlayer(snap, wa.ID).X < x0+40 {
		next(a, "snapshot", &snap)
	}

	a.CloseNow()
	for !findPlayer(snap, wa.ID).Away {
		next(b, "snapshot", &snap)
	}
	frozenX := findPlayer(snap, wa.ID).X
	next(b, "snapshot", &snap)
	if x := findPlayer(snap, wa.ID).X; x != frozenX {
		t.Fatalf("away player moved from %v to %v", frozenX, x)
	}

	a2 := dial()
	send(a2, clientMsg{Type: "join", Token: wa.Token})
	var wa2 welcomeMsg
	next(a2, "welcome", &wa2)
	if wa2.ID != wa.ID || wa2.Phase != phasePlaying {
		t.Fatalf("reconnect welcome = %+v, want id %d playing", wa2, wa.ID)
	}
	next(a2, "start", &st)
	next(a2, "snapshot", &snap)
	if p := findPlayer(snap, wa.ID); p.Away || p.X < frozenX-1 {
		t.Fatalf("reconnected player = %+v, want back at x≈%v", p, frozenX)
	}

	// Same token in a second tab kicks the first.
	a3 := dial()
	send(a3, clientMsg{Type: "join", Token: wa.Token})
	next(a3, "welcome", &wa2)
	rejected(a2, "opened in another tab")
}
