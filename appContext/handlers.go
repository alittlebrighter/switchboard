package appContext

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"github.com/alittlebrighter/switchboard/persistence"
	"github.com/alittlebrighter/switchboard/routing"
	"github.com/alittlebrighter/switchboard/util"
)

const byteChunkSize = 256

func (sCtx *ServerContext) processMessage(data []byte) {
	env := new(routing.Envelope)
	if err := util.Unmarshal(data, env); err != nil {
		fmt.Printf("Error parsing data: %s\n", err.Error())
	}

	if sCtx.DeliverEnvelope(env) {
		sCtx.SaveMessages(env.To, persistence.Mailbox{env})
	}
}

// WebsocketConn manages websocket connections coming from the Raspberry Pis and user devices
func (sCtx *ServerContext) WebsocketConn(ws *websocket.Conn) {
	connKey := ws.Config().Origin.Host
	log.Printf("Connection started from: %s", connKey)

	// register our connection
	connChan := sCtx.AddControllerConn(connKey)

	// read messages received on the websocket and route them
	go util.ReadFromWebSocket(ws, sCtx.processMessage)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	go func() {
		// read and send messages delivered to our address
		for msg := range connChan {
			data, err := json.Marshal(msg)
			if err != nil {
				log.Printf("Error marshalling message: %s\n", err.Error())
				continue
			}
			if _, err = ws.Write(data); err != nil {
				ws.Close()
				sCtx.SaveMessages(connKey, persistence.Mailbox{msg})
			}
		}
		wg.Done()
	}()

	// fetch persisted messages and send them now that there is a connection
	newMsgs, err := sCtx.GetMessages(connKey)
	if err == nil {
		for _, unopened := range newMsgs {
			data, err := util.MarshalToMimeType(unopened, "encoding/json")
			if _, err = ws.Write(data); err != nil {
				ws.Close()
				sCtx.SaveMessages(connKey, persistence.Mailbox{unopened})
			}
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
		storedMessages, err := sCtx.GetMessages(r.FormValue("to"))
		if err != nil {
			w.Write([]byte("Error retrieving messages: " + err.Error()))
			return
		}

		marshalled, err := util.MarshalResponse(r, storedMessages)
		if err != nil {
			w.Write([]byte("Error marshalling messages: " + err.Error()))
			return
		}
		w.Write(marshalled)
	case "post":
		defer r.Body.Close()

		env := new(routing.Envelope)
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
