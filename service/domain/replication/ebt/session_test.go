package ebt_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/stretchr/testify/require"
)

func TestSession_SendNotesSendsEmptyNotesOnlyDuringInitialUpdate(t *testing.T) {
	s := newTestSession(t)

	err := s.Session.SendNotes()
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNotes{
			messages.MustNewEbtReplicateNotes(nil),
		},
		s.Stream.sentNotes,
	)

	err = s.Session.SendNotes()
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNotes{
			messages.MustNewEbtReplicateNotes(nil),
		},
		s.Stream.sentNotes,
	)
}

func TestSession_SendNotesSendsNonEmptyNotesDuringConsecutiveUpdates(t *testing.T) {
	s := newTestSession(t)

	s.ContactsStorage.contacts = nil

	err := s.Session.SendNotes()
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNotes{
			messages.MustNewEbtReplicateNotes(nil),
		},
		s.Stream.sentNotes,
	)

	contact := replication.Contact{
		Who:       fixtures.SomeRefFeed(),
		Hops:      graph.MustNewHops(1),
		FeedState: replication.NewEmptyFeedState(),
	}

	s.ContactsStorage.contacts = []replication.Contact{contact}

	err = s.Session.SendNotes()
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNotes{
			messages.MustNewEbtReplicateNotes(nil),
			messages.MustNewEbtReplicateNotes([]messages.EbtReplicateNote{
				messages.MustNewEbtReplicateNote(
					contact.Who,
					true,
					true,
					0,
				),
			}),
		},
		s.Stream.sentNotes,
	)
}

type testSession struct {
	Session         *ebt.Session
	ContactsStorage *contactsStorage
	Stream          *mockStream
}

func newTestSession(t *testing.T) testSession {
	ctx := fixtures.TestContext(t)
	logger := fixtures.TestLogger(t)
	stream := newMockStream()
	contactsStorage := newContactsStorage()
	session := ebt.NewSession(
		ctx,
		stream,
		logger,
		nil,
		nil,
		contactsStorage,
	)

	return testSession{
		Session:         session,
		ContactsStorage: contactsStorage,
		Stream:          stream,
	}
}

type mockStream struct {
	sentNotes []messages.EbtReplicateNotes
}

func newMockStream() *mockStream {
	return &mockStream{}
}

func (m *mockStream) IncomingMessages(ctx context.Context) <-chan ebt.IncomingMessage {
	//TODO implement me
	panic("implement me")
}

func (m *mockStream) SendNotes(notes messages.EbtReplicateNotes) error {
	m.sentNotes = append(m.sentNotes, notes)
	return nil
}

func (m *mockStream) SendMessage(msg *message.Message) error {
	//TODO implement me
	panic("implement me")
}

type contactsStorage struct {
	contacts []replication.Contact
}

func newContactsStorage() *contactsStorage {
	return &contactsStorage{}
}

func (c contactsStorage) GetContacts() ([]replication.Contact, error) {
	return c.contacts, nil
}
