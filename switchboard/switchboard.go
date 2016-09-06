package switchboard

const channelBufferSize = 5

// Envelope wraps the encrypted contents of a message with minimal metadata to help the relay server know
// what to do with the message
type Envelope struct {
    Destination string
    TTL         int64
    Contents    string
}

// Switchboard manages connections and routing messages between connections
type Switchboard struct {
    conns map[string]chan string
}

// NewSwitchboard returns a new pointer to a switchboard instance
func NewSwitchboard() *Switchboard {
    return &Switchboard{conns: make(map[string]chan string)}
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (s *Switchboard) AddControllerConn(connID string) chan string {
    connChan := make(chan string, channelBufferSize)
    s.conns[connID] = connChan
    return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (s *Switchboard) FetchControllerConn(connID string) chan string {
    if conn, ok := s.conns[connID]; ok {
        return conn
    }
    return nil
}

// CloseControllerConn closes and deletes the channel linked to a controller websocket connection
func (s *Switchboard) CloseControllerConn(connID string) {
    close(s.conns[connID])
    delete(s.conns, connID)
}

// DeliverEnvelope attempts to deliver the contents of an Envelope if the destination is connected
// otherwise it returns true to indicate to the caller that the contents should be persisted
func (s *Switchboard) DeliverEnvelope(envelope *Envelope) bool {
    if conn := s.FetchControllerConn(envelope.Destination); conn != nil {
        conn <- envelope.Contents
        return false
    }
    return true
}
