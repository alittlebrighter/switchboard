package main

import ()

// ConnectionNotFound is thrown when a controller connection cannot be found in the ServerContext
type ConnectionNotFound struct{}

func (cErr *ConnectionNotFound) Error() string {
	return "Connection not found"
}

// ServerContext maintains the map of controller IDs and their corresponding channels linked to the active websocket
type ServerContext struct {
	controllerConns map[string]chan []byte
}

// NewServerContext returns a pointer to a new instance
func NewServerContext() *ServerContext {
	return &ServerContext{controllerConns: make(map[string]chan []byte)}
}

// AddControllerConn adds or overwrites the channel linked to a controller websocket connection
func (sCtx *ServerContext) AddControllerConn(connID string) chan [string][]byte {
	connChan := make(chan []byte)
	sCtx.controllerConns[connID] = connChan
	return connChan
}

// FetchControllerConn retrieves the channel linked to a controller websocket connection
func (sCtx *ServerContext) FetchControllerConn(connID string) (chan []byte, error) {
	if conn, ok := controllerConns[connID]; ok {
		return conn, nil
	}
	return nil, new(ConnectionNotFound)
}

// CloseControllerConn deletes the channel linked to a controller websocket connection
func (sCtx *ServerContext) CloseControllerConn(connID string) {
	delete(sCtx.controllerConns, connID)
}
