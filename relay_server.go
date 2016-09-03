package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"golang.org/x/net/websocket"
)

var controllerConns map[string]chan []byte

// HouseControllerConn manages websocket connections coming from the Raspberry Pis
func HouseControllerConn(ws *websocket.Conn) {
	log.Printf("Connection started from: %s\n", ws.Config().Origin.Host)

	connsKey := ws.Config().Origin.Host

	controllerConns[connsKey] = make(chan []byte)

	for {
		msg := <-controllerConns[connsKey]
		if string(msg) == "END" {
			break
		}

		ws.Write(msg)
	}
}

// UserDeviceAPI receives requests from a user's mobile app and routes them to the correct Raspberry Pi websocket connection
func UserDeviceAPI(w http.ResponseWriter, r *http.Request) {
	controllerID := r.FormValue("controller")

	// reading the body of the request in this way prevents overflow attacks
	msg, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		w.Write([]byte(fmt.Sprintf("Error reading request: %v", err)))
		return
	}

	if conn, ok := controllerConns[controllerID]; ok {
		conn <- msg
		w.Write([]byte(fmt.Sprintf("Sent command to %s.", controllerID)))
	} else {
		w.Write([]byte(fmt.Sprintf("Controller %s not found.", controllerID)))
	}
}

func main() {
	controllerConns = make(map[string]chan []byte)

	http.Handle("/listen", websocket.Handler(HouseControllerConn))
	http.HandleFunc("/command", UserDeviceAPI)

	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
