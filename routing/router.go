package routing

import (
	"github.com/alittlebrighter/switchboard/models"
	uuid "github.com/satori/go.uuid"
)

const channelBufferSize = 5

// Switchboard manages connections and routing messages between connections
type Switchboard map[*uuid.UUID]chan *models.Envelope

// NewSwitchboard returns a new pointer to a switchboard instance
func NewSwitchboard() Switchboard {
	return make(Switchboard)
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (s Switchboard) AddControllerConn(connID *uuid.UUID) chan *models.Envelope {
	connChan := make(chan *models.Envelope, channelBufferSize)
	s[connID] = connChan
	return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (s Switchboard) FetchControllerConn(connID *uuid.UUID) chan *models.Envelope {
	if conn, ok := s[connID]; ok {
		return conn
	}
	return nil
}

// CloseControllerConn closes and deletes the channel linked to a controller websocket connection
func (s Switchboard) CloseControllerConn(connID *uuid.UUID) {
	close(s[connID])
	delete(s, connID)
}

// DeliverEnvelope attempts to deliver the contents of an Envelope if the destination is connected
// it returns a flag denoting whether the envelope was delivered immediately (true) or stored for later pickup (false)
func (s Switchboard) DeliverEnvelope(envelope *models.Envelope) bool {
	conn := s.FetchControllerConn(envelope.To)
	if conn == nil {
		return false
	}
	conn <- envelope
	return true
}
