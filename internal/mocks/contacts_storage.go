package mocks

import (
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type ContactsStorageMock struct {
	Contacts []replication.Contact
}

func NewContactsStorageMock() *ContactsStorageMock {
	return &ContactsStorageMock{}
}

func (s ContactsStorageMock) GetContacts(peer identity.Public) ([]replication.Contact, error) {
	return s.Contacts, nil
}
