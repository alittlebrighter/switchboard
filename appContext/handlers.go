package appContext

import (
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/gegillam/pi-webserver/persistence"
	"github.com/gegillam/pi-webserver/switchboard"
	"github.com/gegillam/pi-webserver/util"
)

// WebsocketConn manages websocket connections coming from the Raspberry Pis and user devices
func (sCtx *ServerContext) WebsocketConn(ws *websocket.Conn) {
	connKey := ws.Config().Origin.Host
	log.Printf("Connection started from: %s", connKey)

	// register our connection
	connChan := sCtx.AddControllerConn(connKey)

	// read messages received on the websocket and route them
	go func() {
		for {
			var msg = make([]byte, 1024)
			n, err := ws.Read(msg)
			if err != nil {
				log.Println("Error reading incoming message: " + err.Error())
				break
			}

			env := new(switchboard.Envelope)
			err = util.Unmarshal(msg[:n], env)
			if err != nil {
				log.Println("Error parsing incoming message: " + err.Error())
				continue
			}
			if sCtx.DeliverEnvelope(env) {
				sCtx.SaveMessages(env.Destination, persistence.Mailbox{env.Contents})
			}
		}
	}()

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() { 
	// read and send messages delivered to our address
		for msg := range connChan {
			if _, err := ws.Write([]byte(msg)); err != nil {
				sCtx.SaveMessages(connKey, persistence.Mailbox{msg})
			}
		}
		wg.Done()
	}()

	// fetch persisted messages and send them now that there is a connection
	newMsgs, err := sCtx.GetMessages(connKey)
	if err == nil {
		for _, unopened := range newMsgs {
			connChan <- unopened
		}
	} else {
		log.Println("Error getting messages: " + err.Error())
	}

	wg.Wait()

	sCtx.CloseControllerConn(connKey)

	log.Printf("Closing connection to: %s", connKey)
}

// HTTPConn receives requests via http and routes them to the correct Raspberry Pi websocket connection
func (sCtx *ServerContext) HTTPConn(w http.ResponseWriter, r *http.Request) {
	switch strings.ToLower(r.Method) {
	case "get":
		storedMessages, err := sCtx.GetMessages(r.FormValue("destination"))
		if err != nil {
			w.Write([]byte("Error retrieving messages: " + err.Error()))
			return
		}

		marshalled, err := util.Marshal(r, storedMessages)
		if err != nil {
			w.Write([]byte("Error marshalling messages: " + err.Error()))
			return
		}
		w.Write(marshalled)
	case "post":
		defer r.Body.Close()

		env := new(switchboard.Envelope)
		if err := util.UnmarshalRequest(r, env); err != nil {
			w.Write([]byte("Error parsing request body: " + err.Error()))
		}

		if sCtx.DeliverEnvelope(env) {
			w.Write([]byte("Message queued."))
		} else {
			w.Write([]byte("Message received."))
		}
	}
	return
}
