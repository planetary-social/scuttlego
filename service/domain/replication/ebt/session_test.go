package ebt_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/replication/mocks"
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

	s.ContactsStorage.Contacts = nil

	err := s.Session.SendNotes()
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNotes{
			messages.MustNewEbtReplicateNotes(nil),
		},
		s.Stream.sentNotes,
	)

	contact := replication.MustNewContact(
		fixtures.SomeRefFeed(),
		graph.MustNewHops(1),
		replication.NewEmptyFeedState(),
	)

	s.ContactsStorage.Contacts = []replication.Contact{contact}

	err = s.Session.SendNotes()
	require.NoError(t, err)

	require.Equal(t,
		[]messages.EbtReplicateNotes{
			messages.MustNewEbtReplicateNotes(nil),
			messages.MustNewEbtReplicateNotes([]messages.EbtReplicateNote{
				messages.MustNewEbtReplicateNote(
					contact.Who(),
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

	s.ContactsStorage.Contacts = nil

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

	go func() {
		err := s.Session.HandleIncomingMessagesLoop()
		t.Log(err)
	}()

	require.Eventually(t,
		func() bool {
			return len(s.FeedRequester.RequestCalls()) == 1
		},
		time.Second, 10*time.Millisecond,
	)
	require.Equal(t, message.MustNewSequence(1), *s.FeedRequester.RequestCalls()[0].Seq)
	require.Equal(t, ref, s.FeedRequester.RequestCalls()[0].Ref)
}

func TestSession_NotesWithReceiveOrReplicateSetToFalseCallRequestedFeedsCancel(t *testing.T) {
	ref := fixtures.SomeRefFeed()

	testCases := []struct {
		Name string
		Note messages.EbtReplicateNote
	}{
		{
			Name: "receive_false",
			Note: messages.MustNewEbtReplicateNote(
				ref,
				false,
				true,
				1,
			),
		},
		{
			Name: "replicate_false",
			Note: messages.MustNewEbtReplicateNote(
				ref,
				true,
				false,
				1,
			),
		},
		{
			Name: "receive_and_replicate_false",
			Note: messages.MustNewEbtReplicateNote(
				ref,
				false,
				false,
				1,
			),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			s := newTestSession(t)

			s.ContactsStorage.Contacts = nil

			go func() {
				s.Stream.ReceiveIncomingMessage(s.Ctx, ebt.NewIncomingMessageWithNotes(
					messages.MustNewEbtReplicateNotes(
						[]messages.EbtReplicateNote{testCase.Note},
					),
				))
			}()

			go func() {
				err := s.Session.HandleIncomingMessagesLoop()
				t.Log(err)
			}()

			require.Eventually(t,
				func() bool {
					return len(s.FeedRequester.CancelCalls()) == 1
				},
				time.Second, 10*time.Millisecond,
			)
			require.Equal(t, ref, s.FeedRequester.CancelCalls()[0].Ref)
		})
	}
}

func TestSession_ErrorsWhenProcessingRawMessagesDontTerminateTheSession(t *testing.T) {
	s := newTestSession(t)

	s.RawMessageHandler.HandleError = fixtures.SomeError()

	go func() {
		s.Stream.ReceiveIncomingMessage(s.Ctx,
			ebt.NewIncomingMessageWithMessage(
				fixtures.SomeRawMessage(),
			),
		)
	}()

	ch := make(chan error)
	go func() {
		defer close(ch)
		ch <- s.Session.HandleIncomingMessagesLoop()
	}()

	require.Eventually(t,
		func() bool {
			return len(s.RawMessageHandler.HandleCalls()) > 0
		},
		1*time.Second, 10*time.Millisecond,
	)

	select {
	case err := <-ch:
		t.Fatalf("loop returned '%s' instead of continuing to run", err)
	case <-time.After(1 * time.Second):
		t.Log("ok")
	}
}

type testSession struct {
	Session           *ebt.Session
	ContactsStorage   *mocks.ContactsStorageMock
	Stream            *mockStream
	MessageStreamer   *messageStreamerMock
	Ctx               context.Context
	FeedRequester     *feedRequesterMock
	RawMessageHandler *rawMessageHandlerMock
}

func newTestSession(t *testing.T) testSession {
	ctx := fixtures.TestContext(t)
	logger := fixtures.TestLogger(t)
	stream := newMockStream()
	contactsStorage := mocks.NewContactsStorageMock()
	fr := newFeedRequesterMock()
	handler := newRawMessageHandlerMock()
	session := ebt.NewSession(
		ctx,
		stream,
		logger,
		handler,
		contactsStorage,
		fr,
	)

	return testSession{
		Session:           session,
		ContactsStorage:   contactsStorage,
		Stream:            stream,
		FeedRequester:     fr,
		RawMessageHandler: handler,
		Ctx:               ctx,
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

func (m *mockStream) RemoteIdentity() identity.Public {
	return fixtures.SomePublicIdentity()
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

type feedRequesterMock struct {
	requestCalls []feedRequesterRequestCall
	cancelCalls  []feedRequesterCancelCall
	lock         sync.Mutex
}

func newFeedRequesterMock() *feedRequesterMock {
	return &feedRequesterMock{}
}

func (f *feedRequesterMock) Request(ctx context.Context, ref refs.Feed, seq *message.Sequence) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.requestCalls = append(f.requestCalls, feedRequesterRequestCall{
		Ctx: ctx,
		Ref: ref,
		Seq: seq,
	})
}

func (f *feedRequesterMock) Cancel(ref refs.Feed) {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.cancelCalls = append(f.cancelCalls, feedRequesterCancelCall{
		Ref: ref,
	})
}

func (f *feedRequesterMock) RequestCalls() []feedRequesterRequestCall {
	f.lock.Lock()
	defer f.lock.Unlock()

	tmp := make([]feedRequesterRequestCall, len(f.requestCalls))
	copy(tmp, f.requestCalls)
	return tmp
}

func (f *feedRequesterMock) CancelCalls() []feedRequesterCancelCall {
	f.lock.Lock()
	defer f.lock.Unlock()

	tmp := make([]feedRequesterCancelCall, len(f.cancelCalls))
	copy(tmp, f.cancelCalls)
	return tmp
}

type feedRequesterRequestCall struct {
	Ctx context.Context
	Ref refs.Feed
	Seq *message.Sequence
}

type feedRequesterCancelCall struct {
	Ref refs.Feed
}

type rawMessageHandlerMock struct {
	handleCalls []message.RawMessage
	HandleError error
	lock        sync.Mutex
}

func newRawMessageHandlerMock() *rawMessageHandlerMock {
	return &rawMessageHandlerMock{}
}

func (r *rawMessageHandlerMock) Handle(replicatedFrom identity.Public, msg message.RawMessage) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	r.handleCalls = append(r.handleCalls, msg)
	return r.HandleError
}

func (r *rawMessageHandlerMock) HandleCalls() []message.RawMessage {
	r.lock.Lock()
	defer r.lock.Unlock()
	tmp := make([]message.RawMessage, len(r.handleCalls))
	copy(tmp, r.handleCalls)
	return tmp
}
