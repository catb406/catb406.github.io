package messenger

import (
	"SB/catb406.github.io/service/repository/persistence"
	"time"
)

type (
	messenger struct {
		persistent persistence.Persistent
	}

	Messenger interface {
		AddMessage(msg persistence.PersistentObject) error
		GetDialogs(idUser int64) ([]persistence.PersistentObject, error)
		AddRequest(req persistence.PersistentObject) error
		UpdateRequest(req persistence.PersistentObject) (persistence.PersistentObject, error)
		GetMessages(idUsers []int64, t time.Time) ([]persistence.PersistentObject, error)
	}
)

func NewMessenger(persistent persistence.Persistent) Messenger {
	return &messenger{
		persistent: persistent,
	}
}

func (messenger *messenger) AddMessage(msg persistence.PersistentObject) error {
	return messenger.persistent.AddMessage(msg)
}

func (messenger *messenger) GetDialogs(idUser int64) ([]persistence.PersistentObject, error) {
	return messenger.persistent.GetDialogs(idUser)
}

func (messenger *messenger) AddRequest(req persistence.PersistentObject) error {
	return messenger.persistent.AddRequest(req)
}

func (messenger *messenger) UpdateRequest(req persistence.PersistentObject) (persistence.PersistentObject, error) {
	return messenger.persistent.UpdateRequest(req)
}

func (messenger *messenger) GetMessages(idUsers []int64, t time.Time) ([]persistence.PersistentObject, error) {
	nullTime := time.Time{}
	if t == nullTime {
		t = time.Now()
	}
	return messenger.persistent.GetMessages(idUsers, t)
}
