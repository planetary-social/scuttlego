package ebt_test

import (
	"context"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
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

func TestSession_NotesWithReceiveAndReplicateSetToTrueCallRequestedFeedsRequest(t *testing.T) {
	s := newTestSession(t)

	s.ContactsStorage.contacts = nil

	ref := fixtures.SomeRefFeed()

	go func() {
		s.Stream.ReceiveIncomingMessage(s.Ctx, ebt.NewIncomingMessageWithNotes(
			messages.MustNewEbtReplicateNotes(
				[]messages.EbtReplicateNote{
					messages.MustNewEbtReplicateNote(
						ref,
						true,
						true,
						1,
					),
				}),
		))
	}()

	require.Eventually(t,
		func() bool {
			return len(s.FeedRequester.RequestCalls) == 1
		},
		time.Second, 10*time.Millisecond,
	)
	require.Equal(t, message.MustNewSequence(1), *s.FeedRequester.RequestCalls[0].Seq)
	require.Equal(t, ref, s.FeedRequester.RequestCalls[0].Ref)
}

func TestSession_NotesWithReceiveSetToFalseCallRequestedFeedsCancel(t *testing.T) {
	s := newTestSession(t)

	s.ContactsStorage.contacts = nil

	ref := fixtures.SomeRefFeed()

	go func() {
		s.Stream.ReceiveIncomingMessage(s.Ctx, ebt.NewIncomingMessageWithNotes(
			messages.MustNewEbtReplicateNotes(
				[]messages.EbtReplicateNote{
					messages.MustNewEbtReplicateNote(
						ref,
						false,
						true,
						1,
					),
				}),
		))
	}()

	require.Eventually(t,
		func() bool {
			return len(s.FeedRequester.CancelCalls) == 1
		},
		time.Second, 10*time.Millisecond,
	)
	require.Equal(t, ref, s.FeedRequester.CancelCalls[0].Ref)
}

func TestSession_NotesWithReplicateSetToFalseCallRequestedFeedsCancel(t *testing.T) {
	s := newTestSession(t)

	s.ContactsStorage.contacts = nil

	ref := fixtures.SomeRefFeed()

	go func() {
		s.Stream.ReceiveIncomingMessage(s.Ctx, ebt.NewIncomingMessageWithNotes(
			messages.MustNewEbtReplicateNotes(
				[]messages.EbtReplicateNote{
					messages.MustNewEbtReplicateNote(
						ref,
						true,
						false,
						1,
					),
				}),
		))
	}()

	require.Eventually(t,
		func() bool {
			return len(s.FeedRequester.CancelCalls) == 1
		},
		time.Second, 10*time.Millisecond,
	)
	require.Equal(t, ref, s.FeedRequester.CancelCalls[0].Ref)
}

type testSession struct {
	Session         *ebt.Session
	ContactsStorage *contactsStorage
	Stream          *mockStream
	MessageStreamer *messageStreamerMock
	Ctx             context.Context
	FeedRequester   *feedRequesterMock
}

func newTestSession(t *testing.T) testSession {
	ctx := fixtures.TestContext(t)
	logger := fixtures.TestLogger(t)
	stream := newMockStream()
	contactsStorage := newContactsStorage()
	fr := newFeedRequesterMock()
	session := ebt.NewSession(
		ctx,
		stream,
		logger,
		nil,
		contactsStorage,
		fr,
	)
	go session.HandleIncomingMessages()

	return testSession{
		Session:         session,
		ContactsStorage: contactsStorage,
		Stream:          stream,
		FeedRequester:   fr,
		Ctx:             ctx,
	}
}

type mockStream struct {
	sentNotes []messages.EbtReplicateNotes
	in        chan ebt.IncomingMessage
}

func newMockStream() *mockStream {
	return &mockStream{
		in: make(chan ebt.IncomingMessage),
	}
}

func (m *mockStream) IncomingMessages(ctx context.Context) <-chan ebt.IncomingMessage {
	out := make(chan ebt.IncomingMessage)
	go func() {
		defer close(out)
		for v := range m.in {
			select {
			case <-ctx.Done():
				return
			case out <- v:
				continue
			}
		}
	}()
	return out
}

func (m *mockStream) ReceiveIncomingMessage(ctx context.Context, incomingMessage ebt.IncomingMessage) {
	select {
	case m.in <- incomingMessage:
	case <-ctx.Done():
	}
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

type feedRequesterMock struct {
	RequestCalls []feedRequesterRequestCall
	CancelCalls  []feedRequesterCancelCall
}

func newFeedRequesterMock() *feedRequesterMock {
	return &feedRequesterMock{}
}

func (f *feedRequesterMock) Request(ctx context.Context, ref refs.Feed, seq *message.Sequence) {
	f.RequestCalls = append(f.RequestCalls, feedRequesterRequestCall{
		Ctx: ctx,
		Ref: ref,
		Seq: seq,
	})
}

func (f *feedRequesterMock) Cancel(ref refs.Feed) {
	f.CancelCalls = append(f.CancelCalls, feedRequesterCancelCall{
		Ref: ref,
	})
}

type feedRequesterRequestCall struct {
	Ctx context.Context
	Ref refs.Feed
	Seq *message.Sequence
}

type feedRequesterCancelCall struct {
	Ref refs.Feed
}
