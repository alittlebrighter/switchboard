package main

import (
	"flag"
	"net/http"

	"golang.org/x/net/websocket"

	"github.com/gegillam/pi-webserver/appContext"
	"github.com/gegillam/pi-webserver/persistence"
)

func main() {
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	flag.Parse()

	serverCtx := appContext.NewServerContext(persistence.MapBackend)

	http.Handle("/socket", websocket.Handler(serverCtx.WebsocketConn))
	http.HandleFunc("/messages", serverCtx.HTTPConn)

	err := http.ListenAndServe(*host, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
