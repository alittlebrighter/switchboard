package persistence

import (
	"errors"
)

type Mailbox []string

type MessageRepository interface {
	SaveMessages(string, Mailbox) error
	GetMessages(string) (Mailbox, error)
}

var ErrMailboxNotFound = errors.New("No mailbox found at address specified.")

type MapRepository map[string]Mailbox

func NewMapRepository() MapRepository {
	return MapRepository(make(map[string]Mailbox))
}

func (repo MapRepository) SaveMessages(address string, msgs Mailbox) error {
	if mailbox, ok := repo[address]; ok {
		mailbox = append(mailbox, msgs...)
	} else {
		repo[address] = msgs
	}
	return nil
}

func (repo MapRepository) GetMessages(address string) (Mailbox, error) {
	if box, ok := repo[address]; ok {
		var unopened Mailbox
		copy(unopened, box)
		delete(repo, address)
		return box, nil
	}
	return nil, ErrMailboxNotFound
}
