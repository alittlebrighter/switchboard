package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/websocket"

	"github.com/gegillam/pi-webserver/persistence"
)

func main() {
	serverCtx := NewServerContext()
	serverCtx.msgRepo = persistence.NewMapRepository()

	http.Handle("/socket", websocket.Handler(serverCtx.WebsocketConn))
	http.HandleFunc("/messages", serverCtx.HTTPConn)

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

// WebsocketConn manages websocket connections coming from the Raspberry Pis
func (sCtx *ServerContext) WebsocketConn(ws *websocket.Conn) {
	connKey := ws.Config().Origin.Host
	log.Printf("Connection started from: %s", connKey)

	connChan := sCtx.AddControllerConn(connKey)

	newMsgs, err := sCtx.msgRepo.GetMessages(connKey)
	if err == nil {
		for _, unopened := range newMsgs {
			connChan <- unopened
		}
	} else {
		log.Println(err.Error())
	}

	for msg := range connChan {
		if _, err := ws.Write(msg); err != nil {
			sCtx.msgRepo.SaveMessages(connKey, persistence.Mailbox{msg})
			break
		}
	}
	toBeDelivered := persistence.Mailbox(make([][]byte, 0))
	for queued := range connChan {
		toBeDelivered = append(toBeDelivered, queued)
	}
	sCtx.msgRepo.SaveMessages(connKey, toBeDelivered)
	sCtx.CloseControllerConn(connKey)

	log.Printf("Closing connection to: %s", connKey)
}

// HTTPConn receives requests via http and routes them to the correct Raspberry Pi websocket connection
func (sCtx *ServerContext) HTTPConn(w http.ResponseWriter, r *http.Request) {
	switch strings.ToLower(r.Method) {
	case "post":
		// should be query parameter since we're sending the body of the request as the message
		destination := r.FormValue("destination")

		// reading the body of the request in this way prevents overflow attacks
		msg, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error reading request: %v", err)))
			return
		}

		conn := sCtx.FetchControllerConn(destination)
		if conn == nil {
			sCtx.msgRepo.SaveMessages(destination, persistence.Mailbox{msg})
			w.Write([]byte("Message delivered to " + destination))
			return
		}

		conn <- msg
		w.Write([]byte("Message received by %s" + destination))
	}
}

// ServerContext maintains the map of controller IDs and their corresponding channels linked to the active websocket
type ServerContext struct {
	msgRepo         persistence.MessageRepository
	controllerConns map[string]chan []byte
}

// NewServerContext returns a pointer to a new instance
func NewServerContext() *ServerContext {
	return &ServerContext{controllerConns: make(map[string]chan []byte)}
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (sCtx *ServerContext) AddControllerConn(connID string) chan []byte {
	connChan := make(chan []byte, 10)
	sCtx.controllerConns[connID] = connChan
	return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (sCtx *ServerContext) FetchControllerConn(connID string) chan []byte {
	if conn, ok := sCtx.controllerConns[connID]; ok {
		return conn
	}
	return nil
}

// CloseControllerConn deletes the channel linked to a controller websocket connection
func (sCtx *ServerContext) CloseControllerConn(connID string) {
	close(sCtx.controllerConns[connID])
	delete(sCtx.controllerConns, connID)
}
