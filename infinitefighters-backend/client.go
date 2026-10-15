package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
)

const (
	sendBuffer   = 64 // ~2 s of snapshots
	writeTimeout = 5 * time.Second
	maxMsgBytes  = 4096
)

type client struct {
	conn   *websocket.Conn
	send   chan []byte
	cancel context.CancelFunc
	player *Player // owned by the room goroutine
	closing bool   // owned by the room goroutine
}

// push never blocks the room: a client whose buffer is full is dropped.
func (c *client) push(b []byte) {
	select {
	case c.send <- b:
	default:
		c.cancel()
	}
}

// reject closes the socket with a reason the browser sees as CloseEvent.reason.
func (c *client) reject(reason string) {
	c.closing = true
	go c.conn.Close(websocket.StatusPolicyViolation, reason)
}

func (c *client) writeLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case b := <-c.send:
			wctx, cancel := context.WithTimeout(ctx, writeTimeout)
			err := c.conn.Write(wctx, websocket.MessageText, b)
			cancel()
			if err != nil {
				c.cancel()
				return
			}
		}
	}
}

func serveWS(room *Room) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			// Same-origin is always allowed; this adds the Vite dev server.
			OriginPatterns: []string{"localhost:*", "127.0.0.1:*"},
		})
		if err != nil {
			return
		}
		conn.SetReadLimit(maxMsgBytes)
		ctx, cancel := context.WithCancel(r.Context())
		c := &client{conn: conn, send: make(chan []byte, sendBuffer), cancel: cancel}
		go c.writeLoop(ctx)
		defer func() {
			cancel()
			conn.CloseNow()
			room.inbox <- func() { room.disconnect(c) }
		}()

		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var m clientMsg
			if err := json.Unmarshal(data, &m); err != nil {
				log.Printf("bad message: %v", err)
				continue
			}
			room.inbox <- func() { room.handle(c, m) }
		}
	}
}
