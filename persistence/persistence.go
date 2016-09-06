package persistence

import (
	"errors"
)

// Mailbox stores an array of encrypted and encoded messages
type Mailbox []string

var errMailboxNotFound = errors.New("No mailbox found at the address specified.")

type Backend int8

const (
	MapBackend = iota
)

// MessageRepository defines the methods required to store and retrieve messages routed through the server
type MessageRepository interface {
	SaveMessages(string, Mailbox) error
	GetMessages(string) (Mailbox, error)
}

// MapRepository is a simple implementation of a Message Repository
type MapRepository map[string]Mailbox

// NewMessageRepository returns a new MessageRepository with the specified backend.
func NewMessageRepository(backend Backend) MessageRepository {
	var repo MessageRepository
	switch backend {
	case MapBackend:
		repo = make(MapRepository)
	}
	return repo
}

// SaveMessages saves an array of messages in the target address' mailbox
func (repo MapRepository) SaveMessages(address string, msgs Mailbox) error {
	if mailbox, ok := repo[address]; ok {
		mailbox = append(mailbox, msgs...)
	} else {
		repo[address] = msgs
	}
	return nil
}

// GetMessages retrieves all of the messages stored in a mailbox at an address and removes the associated address' mailbox
func (repo MapRepository) GetMessages(address string) (Mailbox, error) {
	if box, ok := repo[address]; ok {
		var unopened Mailbox
		copy(unopened, box)
		delete(repo, address)
		return box, nil
	}
	return nil, errMailboxNotFound
}
