package appContext

import (
	"github.com/alittlebrighter/switchboard/persistence"
	"github.com/alittlebrighter/switchboard/routing"
)

// ServerContext maintains the map of controller IDs and their corresponding channels linked to the active websocket
type ServerContext struct {
	persistence.UserRepository
	routing.Switchboard
}

// NewServerContext returns a pointer to a new instance
func NewServerContext(persistenceBackend persistence.Backend) *ServerContext {
	return &ServerContext{Switchboard: routing.NewSwitchboard(),
		UserRepository: persistence.NewUserRepository(persistenceBackend)}
}
