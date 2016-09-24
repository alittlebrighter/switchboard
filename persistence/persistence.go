package persistence

import (
	"errors"

	"github.com/satori/go.uuid"

	"github.com/alittlebrighter/switchboard/models"
)

var errUserNotFound = errors.New("No user found with that ID.")

type Backend string

const (
	MapBackend  = "map"
	BoltBackend = "bolt"
)

// UserRepository defines the methods required to store and retrieve messages routed through the server
type UserRepository interface {
	SaveUser(*models.User) error
	GetUser(*uuid.UUID) (*models.User, error)
}

// NewMessageRepository returns a new MessageRepository with the specified backend.
func NewUserRepository(backend Backend) UserRepository {
	var repo UserRepository
	switch backend {
	case MapBackend:
		repo = make(MapRepository)
	}
	return repo
}

// MapRepository is a simple implementation of a Message Repository
type MapRepository map[uuid.UUID]*models.User

// SaveUser saves an array of messages in the target address' models.Mailbox
func (repo MapRepository) SaveUser(user *models.User) error {
	repo[user.ID] = user
	return nil
}

// GetUser retrieves all of the messages stored in a models.Mailbox at an address and removes the associated address' models.Mailbox
func (repo MapRepository) GetUser(id *uuid.UUID) (*models.User, error) {
	if user, ok := repo[*id]; ok {
		return user, nil
	}
	return nil, errUserNotFound
}
