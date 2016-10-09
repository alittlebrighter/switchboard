package appContext

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"

	logger "github.com/Sirupsen/logrus"
	"github.com/satori/go.uuid"
	"golang.org/x/net/websocket"

	"github.com/alittlebrighter/switchboard/models"
	"github.com/alittlebrighter/switchboard/util"
)

const byteChunkSize = 256

func (sCtx *ServerContext) processMessage(data []byte) {
	log := logger.WithField("func", "processMessage")

	env := new(models.Envelope)
	if err := util.Unmarshal(data, env); err != nil {
		log.WithError(err).Errorln("Could not parse envelope data.")
	}

	if !sCtx.DeliverEnvelope(env) {
		user, err := sCtx.GetUser(env.To)
		if err == nil {
			user.SaveMessage(env)
			sCtx.SaveUser(user)
		} else {
			log.WithError(err).Warnln("Could not find user mailbox .")
		}
	}
}

// WebsocketConn manages websocket connections coming from the Raspberry Pis and user devices
func (sCtx *ServerContext) WebsocketConn(ws *websocket.Conn) {
	log := logger.WithField("func", "WebsocketConn")

	connKey := ws.Config().Origin.Host
	log.WithField("connKey", connKey).Debugln("Websocket connection started.")

	id, err := uuid.FromString(connKey)
	if err != nil {
		ws.Close()
		return
	}
	user, err := sCtx.GetUser(&id)
	if err != nil {
		user = models.NewUser(&id)
		sCtx.SaveUser(user)
	}

	// register our connection
	connChan := sCtx.AddControllerConn(&id)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	// read messages received on the websocket and route them
	go func() {
		util.ReadFromWebSocket(ws, sCtx.processMessage)
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		// read and send messages delivered to our address
		for msg := range connChan {
			data, err := json.Marshal(msg)
			if err != nil {
				log.WithError(err).Errorln("Could not marshal message.")
				continue
			}
			if _, err = ws.Write(data); err != nil {
				log.Debugln("Websocket closed.  Saving message.")
				ws.Close()
				user.SaveMessage(msg)
				sCtx.SaveUser(user)
			}
		}
		wg.Done()
	}()

	// fetch persisted messages and send them now that there is a connection
	for _, unopened := range user.FlushMessages() {
		data, err := json.Marshal(unopened)
		if _, err = ws.Write(data); err != nil {
			log.Debugln("Websocket closed.  Saving message.")
			ws.Close()
			user.SaveMessage(unopened)
		}
	}

	wg.Wait()

	sCtx.CloseControllerConn(&id)

	log.WithField("connKey", connKey).Debugln("Closing connection.")
}

// HTTPConn receives requests via http and routes them to the correct Raspberry Pi websocket connection
func (sCtx *ServerContext) HTTPConn(w http.ResponseWriter, r *http.Request) {
	log := logger.WithField("func", "HTTPConn")

	log.WithField("httpMethod", r.Method).Debugln("HTTP request received.")
	switch strings.ToLower(r.Method) {
	case "get":
		id, err := uuid.FromString(r.FormValue("to"))
		if err != nil {
			w.Write([]byte("Error parsing ID: " + err.Error()))
			return
		}
		user, err := sCtx.GetUser(&id)
		if err != nil {
			log.WithField("to", r.FormValue("to")).WithError(err).Errorln("Could not retrieve messages.")
			w.Write([]byte("Error retrieving messages: " + err.Error()))
			return
		}
		storedMessages := user.FlushMessages()
		sCtx.SaveUser(user)

		marshalled, err := util.MarshalResponse(r, storedMessages)
		if err != nil {
			log.WithError(err).Errorln("Could not marshal messages for user.")
			w.Write([]byte("Error marshalling messages: " + err.Error()))
			return
		}
		w.Write(marshalled)
	case "post":
		defer r.Body.Close()

		env := new(models.Envelope)
		if err := util.UnmarshalRequest(r, env); err != nil {
			log.WithError(err).Errorln("Could not unmarshal envelope.")
			w.Write([]byte("Error parsing request body: " + err.Error()))
		}

		if !sCtx.DeliverEnvelope(env) {
			user, err := sCtx.GetUser(env.To)
			if err == nil {
				log.WithError(err).Infoln("Could not deliver envelope immediately.")
				user.SaveMessage(env)
				sCtx.SaveUser(user)
			}
		}
	}
	return
}
