package main

import (
	"encoding/json"
	"flag"
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
	host := flag.String("host", "localhost:12345", "The relay host to connect to.")
	flag.Parse()

	serverCtx := NewServerContext(json.Marshal, json.Unmarshal)
	serverCtx.msgRepo = persistence.NewMapRepository()

	http.Handle("/socket", websocket.Handler(serverCtx.WebsocketConn))
	http.HandleFunc("/messages", serverCtx.HTTPConn)

	err := http.ListenAndServe(*host, nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

// WebsocketConn manages websocket connections coming from the Raspberry Pis and user devices
func (sCtx *ServerContext) WebsocketConn(ws *websocket.Conn) {
	connKey := ws.Config().Origin.Host
	log.Printf("Connection started from: %s", connKey)

	go func() {
		for {
			var msg = make([]byte, 1024)
			n, err := ws.Read(msg)
			if err != nil {
				log.Printf("Error reading incoming message: %s", err.Error())
				break
			}

			if _, err := sCtx.deliverEnvelope(msg[:n]); err != nil {
				log.Printf("Error delivering message: " + err.Error())
				break
			}
		}
	}()

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
		if _, err := ws.Write([]byte(msg)); err != nil {
			sCtx.deliverMessage(connKey, -1, msg)
		}
	}

	sCtx.CloseControllerConn(connKey)

	log.Printf("Closing connection to: %s", connKey)
}

// HTTPConn receives requests via http and routes them to the correct Raspberry Pi websocket connection
func (sCtx *ServerContext) HTTPConn(w http.ResponseWriter, r *http.Request) {
	switch strings.ToLower(r.Method) {
	case "get":
		response, err := sCtx.msgRepo.GetMessages(r.FormValue("destination"))
		if err != nil {
			w.Write([]byte("Error retrieving messages: " + err.Error()))
			return
		}

		marshalled, err := sCtx.marshaller(response)
		if err != nil {
			w.Write([]byte("Error marshalling messages: " + err.Error()))
			return
		}
		w.Write(marshalled)
	case "post":
		// reading the body of the request in this way prevents overflow attacks
		msg, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
		if err != nil {
			w.Write([]byte(fmt.Sprintf("Error reading request: %v", err)))
			return
		}

		queued, err := sCtx.deliverEnvelope(msg)
		if err != nil {
			w.Write([]byte("Error delivering message: " + err.Error()))
			return
		}

		if queued {
			w.Write([]byte("Message queued"))
		} else {
			w.Write([]byte("Message received"))
		}
	}
	return
}

// ServerContext maintains the map of controller IDs and their corresponding channels linked to the active websocket
type ServerContext struct {
	msgRepo         persistence.MessageRepository
	controllerConns map[string]chan string
	marshaller      func(interface{}) ([]byte, error)
	unmarshaller    func(data []byte, v interface{}) error
}

// NewServerContext returns a pointer to a new instance
func NewServerContext(marshaller func(interface{}) ([]byte, error), unmarshaller func(data []byte, v interface{}) error) *ServerContext {
	return &ServerContext{controllerConns: make(map[string]chan string),
		marshaller:   marshaller,
		unmarshaller: unmarshaller}
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (sCtx *ServerContext) AddControllerConn(connID string) chan string {
	connChan := make(chan string, 10)
	sCtx.controllerConns[connID] = connChan
	return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (sCtx *ServerContext) FetchControllerConn(connID string) chan string {
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

func (sCtx *ServerContext) deliverEnvelope(toDeliver []byte) (queued bool, err error) {
	envelope := new(Envelope)
	err = sCtx.unmarshaller(toDeliver, envelope)
	if err != nil {
		return
	}

	queued, err = sCtx.deliverMessage(envelope.Destination, envelope.TTL, envelope.Contents)
	return
}

func (sCtx *ServerContext) deliverMessage(destination string, ttl int64, toDeliver string) (queued bool, err error) {
	conn := sCtx.FetchControllerConn(destination)
	if conn == nil {
		err = sCtx.msgRepo.SaveMessages(destination, persistence.Mailbox{toDeliver})
		queued = true
		if err != nil {
			log.Println("Error saving message: " + err.Error())
		}
		return
	}

	conn <- toDeliver
	queued = false

	return
}

type Envelope struct {
	Destination string
	TTL         int64
	Contents    string
}
