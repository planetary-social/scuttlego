package ebt

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type ContactsStorage interface {
	// GetContacts returns a list of contacts. Contacts are sorted by hops,
	// ascending. Contacts include the local feed.
	GetContacts() ([]replication.Contact, error)
}

type MessageWriter interface {
	WriteMessage(msg message.Message) error
}

type MessageStreamer interface {
	Handle(ctx context.Context, id refs.Feed, seq *message.Sequence, messageWriter MessageWriter)
}

type Stream interface {
	IncomingMessages(ctx context.Context) <-chan IncomingMessage
	SendNotes(notes messages.EbtReplicateNotes) error
	SendMessage(msg *message.Message) error
}

type IncomingMessage struct {
	notes *messages.EbtReplicateNotes
	msg   *message.RawMessage
	err   error
}

func NewIncomingMessageWithNote(notes messages.EbtReplicateNotes) IncomingMessage {
	return IncomingMessage{
		notes: &notes,
	}
}

func NewIncomingMessageWithMesage(msg message.RawMessage) IncomingMessage {
	return IncomingMessage{
		msg: &msg,
	}
}

func NewIncomingMessageWithErr(err error) IncomingMessage {
	return IncomingMessage{
		err: err,
	}
}

func (i IncomingMessage) Notes() (messages.EbtReplicateNotes, bool) {
	if i.notes != nil {
		return *i.notes, true
	}
	return messages.EbtReplicateNotes{}, false
}

func (i IncomingMessage) Msg() (message.RawMessage, bool) {
	if i.msg != nil {
		return *i.msg, true
	}
	return message.RawMessage{}, false
}

func (i IncomingMessage) Err() error {
	return i.err
}

type SessionRunner struct {
	logger            logging.Logger
	rawMessageHandler replication.RawMessageHandler
	contactsStorage   ContactsStorage
	streamer          MessageStreamer
}

func NewSessionRunner(
	logger logging.Logger,
	rawMessageHandler replication.RawMessageHandler,
	contactsStorage ContactsStorage,
	streamer MessageStreamer,
) *SessionRunner {
	return &SessionRunner{
		logger:            logger,
		rawMessageHandler: rawMessageHandler,
		contactsStorage:   contactsStorage,
		streamer:          streamer,
	}
}

func (s *SessionRunner) HandleStream(ctx context.Context, stream Stream) error {
	session := NewSession(stream, s.logger, s.rawMessageHandler, s.streamer, s.contactsStorage)
	go session.HandleIncomingMessages(ctx)
	for {
		if err := session.SendNotes(); err != nil {
			s.logger.WithError(err).Debug("error sending our notes")
		}

		select {
		case <-time.After(10 * time.Second):
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

type Session struct {
	stream Stream

	remoteNotes          map[string]messages.EbtReplicateNote
	lock                 sync.Mutex // guards remoteNotes
	sentNotes            *SentNotes
	sentNotesAtLeastOnce bool

	logger            logging.Logger
	rawMessageHandler replication.RawMessageHandler
	streamer          MessageStreamer
	contactsStorage   ContactsStorage
}

func NewSession(
	stream Stream,
	logger logging.Logger,
	rawMessageHandler replication.RawMessageHandler,
	streamer MessageStreamer,
	contactsStorage ContactsStorage,
) *Session {
	return &Session{
		stream: stream,

		remoteNotes: make(map[string]messages.EbtReplicateNote),
		sentNotes:   NewSentNotes(),

		logger:            logger.New("session"),
		rawMessageHandler: rawMessageHandler,
		streamer:          streamer,
		contactsStorage:   contactsStorage,
	}
}

func (s *Session) HandleIncomingMessages(ctx context.Context) {
	for incoming := range s.stream.IncomingMessages(ctx) {
		if err := s.handleIncomingMessage(ctx, incoming); err != nil {
			s.logger.WithError(err).Debug("error processing incoming message")
		}
	}
}

func (s *Session) handleIncomingMessage(ctx context.Context, incoming IncomingMessage) error {
	if err := incoming.Err(); err != nil {
		return errors.Wrap(err, "error receiving messages")
	}

	notes, ok := incoming.Notes()
	if ok {
		s.logger.WithField("number_of_notes", len(notes.Notes())).Debug("received notes")
		return s.handleIncomingNotes(ctx, notes)
	}

	msg, ok := incoming.Msg()
	if ok {
		s.logger.Debug("received a raw message")
		if err := s.rawMessageHandler.Handle(msg); err != nil {
			return errors.Wrap(err, "error handling a message")
		}
		return nil
	}

	return errors.New("logic error")
}

func (s *Session) handleIncomingNotes(ctx context.Context, notes messages.EbtReplicateNotes) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, note := range notes.Notes() {
		s.remoteNotes[note.Ref().String()] = note

		seq, err := s.parseSeq(note.Sequence())
		if err != nil {
			return errors.Wrap(err, "error parsing sequence")
		}

		// todo kill old streams
		s.streamer.Handle(ctx, note.Ref(), seq, NewStreamMessageWriter(s.stream))
	}

	return nil
}

func (s *Session) parseSeq(seq int) (*message.Sequence, error) {
	if seq <= 0 {
		return nil, nil
	}
	sequence, err := message.NewSequence(seq)
	if err != nil {
		return nil, errors.Wrap(err, "new sequence error")
	}
	return &sequence, nil
}

func (s *Session) SendNotes() error {
	contacts, err := s.contactsStorage.GetContacts()
	if err != nil {
		return errors.Wrap(err, "could not get the contacts")
	}

	notesToSend, err := s.sentNotes.Update(contacts)
	if err != nil {
		return errors.Wrap(err, "could not create the notes")
	}

	if notesToSend.Empty() && s.sentNotesAtLeastOnce {
		return nil
	}

	s.sentNotesAtLeastOnce = true

	s.logger.
		WithField("number_of_notes", len(notesToSend.Notes())).
		Debug("sending notes")

	return s.stream.SendNotes(notesToSend)
}

type StreamMessageWriter struct {
	stream Stream
}

func NewStreamMessageWriter(stream Stream) *StreamMessageWriter {
	return &StreamMessageWriter{
		stream: stream,
	}
}

func (s StreamMessageWriter) WriteMessage(msg message.Message) error {
	return s.stream.SendMessage(&msg)
}
