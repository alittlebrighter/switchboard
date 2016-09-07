package switchboard

import (
    "time"
)

const channelBufferSize = 5

// Envelope wraps the encrypted contents of a message with minimal metadata to help the relay server know
// what to do with the message
type Envelope struct {
    Destination string
    TTL         time.Duration
    Contents    string
}

// Switchboard manages connections and routing messages between connections
type Switchboard map[string]chan *Envelope

// NewSwitchboard returns a new pointer to a switchboard instance
func NewSwitchboard() Switchboard {
    return make(Switchboard)
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (s Switchboard) AddControllerConn(connID string) chan *Envelope {
    connChan := make(chan *Envelope, channelBufferSize)
    s[connID] = connChan
    return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (s Switchboard) FetchControllerConn(connID string) chan *Envelope {
    if conn, ok := s[connID]; ok {
        return conn
    }
    return nil
}

// CloseControllerConn closes and deletes the channel linked to a controller websocket connection
func (s Switchboard) CloseControllerConn(connID string) {
    close(s[connID])
    delete(s, connID)
}

// DeliverEnvelope attempts to deliver the contents of an Envelope if the destination is connected
// otherwise it returns true to indicate to the caller that the contents should be persisted
func (s Switchboard) DeliverEnvelope(envelope *Envelope) bool {
    conn := s.FetchControllerConn(envelope.Destination)
    switch {
    case conn == nil && envelope.TTL > 0:
        return true
    case conn == nil && envelope.TTL == 0:
        return false
    default:
        conn <- envelope
        return false
    }
}
