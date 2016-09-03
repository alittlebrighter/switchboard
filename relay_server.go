package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

func main() {
	serverCtx := NewServerContext()

	http.Handle("/listen", websocket.Handler(serverCtx.HouseControllerConn))
	http.HandleFunc("/command", serverCtx.UserDeviceAPI)

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

// HouseControllerConn manages websocket connections coming from the Raspberry Pis
func (sCtx *ServerContext) HouseControllerConn(ws *websocket.Conn) {
	connKey := ws.Config().Origin.Host
	log.Printf("Connection started from: %s", connKey)

	connChan := sCtx.AddControllerConn(connKey)
	for {
		msg := <-connChan
		if string(msg) == "END" {
			break
		}
		log.Printf("Sending %s to %s", msg, connKey)
		ws.Write(msg)
	}
	sCtx.CloseControllerConn(connKey)
	log.Printf("Closing connection to: %s", connKey)
}

// UserDeviceAPI receives requests from a user's mobile app and routes them to the correct Raspberry Pi websocket connection
func (sCtx *ServerContext) UserDeviceAPI(w http.ResponseWriter, r *http.Request) {
	// should be query parameter since we're sending the body of the request as the message
	controllerID := r.FormValue("controller")

	// reading the body of the request in this way prevents overflow attacks
	msg, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error reading request: %v", err)))
		return
	}

	conn, err := sCtx.FetchControllerConn(controllerID)
	if err != nil {
		w.Write([]byte(err.Error()))
		return
	}

	conn <- msg
	w.Write([]byte(fmt.Sprintf("Sent command to %s.", controllerID)))
}

// ConnectionNotFound is thrown when a controller connection cannot be found in the ServerContext
type ConnectionNotFound struct{}

func (cErr *ConnectionNotFound) Error() string {
	return "Connection not found"
}

// ServerContext maintains the map of controller IDs and their corresponding channels linked to the active websocket
type ServerContext struct {
	controllerConns map[string]chan []byte
}

// NewServerContext returns a pointer to a new instance
func NewServerContext() *ServerContext {
	return &ServerContext{controllerConns: make(map[string]chan []byte)}
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (sCtx *ServerContext) AddControllerConn(connID string) chan []byte {
	connChan := make(chan []byte)
	sCtx.controllerConns[connID] = connChan
	return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (sCtx *ServerContext) FetchControllerConn(connID string) (chan []byte, error) {
	if conn, ok := sCtx.controllerConns[connID]; ok {
		return conn, nil
	}
	return nil, new(ConnectionNotFound)
}

// CloseControllerConn deletes the channel linked to a controller websocket connection
func (sCtx *ServerContext) CloseControllerConn(connID string) {
	delete(sCtx.controllerConns, connID)
}
