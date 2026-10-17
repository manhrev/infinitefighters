// Physics sandbox: cp/v2 (Chipmunk2D) world behind a websocket, drawn by web/index.html.
// Authoritative fixed-step server like infinitefighters-backend, but isolated from the game.
package main

import (
	"context"
	"embed"
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
)

//go:embed web
var webFS embed.FS

type clientMsg struct {
	In    *Input
	Spawn string // "circle" | "box" | "boss"
	X, Y  float64
	Reset bool
	Cycle bool // next character
}

type client struct {
	p    *Player
	send chan []byte
}

type hub struct {
	mu      sync.Mutex
	w       *World
	clients map[*client]bool
}

func (h *hub) tick() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.w.Step()
	b, _ := json.Marshal(map[string]any{"type": "snap", "tick": h.w.tick, "bodies": h.w.Snapshot()})
	for c := range h.clients {
		select {
		case c.send <- b:
		default: // slow client: drop the frame
		}
	}
}

func (h *hub) serveWS(rw http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(rw, r, nil)
	if err != nil {
		return
	}
	defer conn.CloseNow()
	ctx := r.Context()

	h.mu.Lock()
	c := &client{p: h.w.AddPlayer(), send: make(chan []byte, 4)}
	h.clients[c] = true
	init, _ := json.Marshal(map[string]any{"type": "init", "id": c.p.id, "w": worldW, "h": worldH, "statics": h.w.Statics})
	h.mu.Unlock()
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		h.w.RemovePlayer(c.p)
		h.mu.Unlock()
	}()

	if conn.Write(ctx, websocket.MessageText, init) != nil {
		return
	}
	go func() {
		for b := range c.send {
			wctx, cancel := context.WithTimeout(ctx, time.Second)
			err := conn.Write(wctx, websocket.MessageText, b)
			cancel()
			if err != nil {
				conn.CloseNow()
				return
			}
		}
	}()
	for {
		_, data, err := conn.Read(ctx)
		if err != nil {
			return
		}
		var m clientMsg
		if json.Unmarshal(data, &m) != nil {
			continue
		}
		h.mu.Lock()
		switch {
		case m.In != nil:
			c.p.in = *m.In
		case m.Cycle:
			h.w.setChar(c.p, c.p.char+1)
		case m.Reset:
			h.w.Reset()
		case m.Spawn != "":
			h.w.SpawnItem(m.Spawn, m.X, m.Y)
		}
		h.mu.Unlock()
	}
}

func main() {
	addr := flag.String("addr", ":8090", "listen address")
	sprites := flag.String("sprites", "../assets/sprites", "sprite dir served at /sprites/ (run from engine-sandbox/)")
	flag.Parse()

	h := &hub{w: NewWorld(), clients: map[*client]bool{}}
	go func() {
		for range time.Tick(time.Second / tickRate) {
			h.tick()
		}
	}()

	web, _ := fs.Sub(webFS, "web")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServerFS(web))
	mux.Handle("/sprites/", http.StripPrefix("/sprites/", http.FileServer(http.Dir(*sprites))))
	mux.HandleFunc("/ws", h.serveWS)
	log.Printf("engine-sandbox on http://localhost%s", *addr)
	log.Fatal(http.ListenAndServe(*addr, mux))
}
