package appContext

import (
	"github.com/alittlebrighter/switchboard/persistence"
	"github.com/alittlebrighter/switchboard/switchboard"
)

// ServerContext maintains the map of controller IDs and their corresponding channels linked to the active websocket
type ServerContext struct {
	persistence.MessageRepository
	switchboard.Switchboard
}

// NewServerContext returns a pointer to a new instance
func NewServerContext(persistenceBackend persistence.Backend) *ServerContext {
	return &ServerContext{Switchboard: switchboard.NewSwitchboard(),
		MessageRepository: persistence.NewMessageRepository(persistenceBackend)}
}
