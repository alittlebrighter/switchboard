package routing

import (
	logger "github.com/Sirupsen/logrus"
	uuid "github.com/satori/go.uuid"

	"github.com/alittlebrighter/switchboard/models"
)

const channelBufferSize = 5

// Switchboard manages connections and routing messages between connections
type Switchboard map[uuid.UUID]chan *models.Envelope

// NewSwitchboard returns a new pointer to a switchboard instance
func NewSwitchboard() Switchboard {
	return make(Switchboard)
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (s Switchboard) AddControllerConn(connID *uuid.UUID) chan *models.Envelope {
	logger.WithField("connID", connID.String()).Debugln("Adding connection to user.")
	connChan := make(chan *models.Envelope, channelBufferSize)
	s[*connID] = connChan
	return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (s Switchboard) FetchControllerConn(connID *uuid.UUID) chan *models.Envelope {
	log := logger.WithField("connID", connID.String())
	log.Debugln("Fetching connection to user.")
	if conn, ok := s[*connID]; ok {
		log.Debugln("Found user connection.")
		return conn
	}
	log.Debugln("Could not find user connection.")
	return nil
}

// CloseControllerConn closes and deletes the channel linked to a controller websocket connection
func (s Switchboard) CloseControllerConn(connID *uuid.UUID) {
	logger.WithField("connID", connID.String()).Debugln("Closing connection to user.")
	close(s[*connID])
	delete(s, *connID)
}

// DeliverEnvelope attempts to deliver the contents of an Envelope if the destination is connected
// it returns a flag denoting whether the envelope was delivered immediately (true) or stored for later pickup (false)
func (s Switchboard) DeliverEnvelope(envelope *models.Envelope) bool {
	log := logger.WithField("to", envelope.To.String())
	conn := s.FetchControllerConn(envelope.To)
	if conn == nil {
		log.Debugln("Cannot deliver envelope right now.")
		return false
	}
	log.Debugln("Delivering envelope now.")
	conn <- envelope
	return true
}
