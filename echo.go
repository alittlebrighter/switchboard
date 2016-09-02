package main

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

// EchoServer echos the data received on the WebSocket.
func EchoServer(ws *websocket.Conn) {
	var msg = make([]byte, 512)
	n, err := ws.Read(msg)
	if err != nil {
		log.Fatal(err)
	}

	response := fmt.Sprintf("Server received: %s\n", msg[:n])
	log.Printf(response)

	ws.Write([]byte(response))
}

func main() {
	http.Handle("/echo", websocket.Handler(EchoServer))

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
