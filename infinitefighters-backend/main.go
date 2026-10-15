package main

import (
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// web/ holds the built frontend. The Vite build must output inside this module: go:embed cannot reach outside it.
//
//go:embed web
var webFS embed.FS

//go:embed levels
var levelFS embed.FS

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	code := flag.String("code", "", "room code (default: random)")
	flag.Parse()

	data, err := levelFS.ReadFile("levels/sprint.json")
	if err != nil {
		log.Fatal(err)
	}
	level, err := parseLevel(data)
	if err != nil {
		log.Fatal(err)
	}
	if *code == "" {
		*code = newRoomCode()
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	room := newRoom(*code, level)
	go room.run(ctx)

	srv := &http.Server{Addr: *addr, Handler: newMux(room)}
	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	fmt.Println("TASK VÔ TẬN server")
	for _, ip := range lanIPs() {
		fmt.Printf("  open  http://%s:%d\n", ip, port)
	}
	fmt.Printf("  local http://localhost:%d\n", port)
	fmt.Printf("  room code: %s\n", *code)

	go func() {
		<-ctx.Done()
		sctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		srv.Shutdown(sctx)
	}()
	if err := srv.Serve(ln); !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}

func newMux(room *Room) *http.ServeMux {
	web, _ := fs.Sub(webFS, "web")
	mux := http.NewServeMux()
	mux.Handle("GET /ws", serveWS(room))
	mux.Handle("GET /levels/", http.FileServerFS(levelFS))
	mux.Handle("GET /", http.FileServerFS(web))
	return mux
}

// lanIPs lists IPv4 addresses of up, non-loopback interfaces. Works offline, unlike dialing a public IP.
func lanIPs() []string {
	var ips []string
	ifaces, _ := net.Interfaces()
	for _, ifc := range ifaces {
		if ifc.Flags&net.FlagUp == 0 || ifc.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := ifc.Addrs()
		for _, a := range addrs {
			if n, ok := a.(*net.IPNet); ok && n.IP.To4() != nil {
				ips = append(ips, n.IP.String())
			}
		}
	}
	return ips
}
