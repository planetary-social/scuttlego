package ebt

import (
	"context"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
)

type RawMessageHandler interface {
	Handle(msg message.RawMessage) error
}

type Stream interface {
	IncomingMessages(ctx context.Context) <-chan IncomingMessage
	SendNote(note messages.EbtReplicateNote)
	SendMessage(msg *message.Message)
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
	remoteNotes map[string]messages.EbtReplicateNote
	lock        sync.Mutex // guards remoteWants

	logger            logging.Logger
	rawMessageHandler RawMessageHandler
}

func NewSessionRunner() *SessionRunner {
	return &SessionRunner{}
}

func (s *SessionRunner) HandleStream(ctx context.Context, stream Stream) error {
	go s.handleIncomingMessages(ctx, stream)

	// todo send

	return nil
}

func (s *SessionRunner) handleIncomingMessages(ctx context.Context, stream Stream) {
	for incoming := range stream.IncomingMessages(ctx) {
		if err := s.handleIncomingMessage(incoming); err != nil {
			s.logger.WithError(err).Debug("error processing incoming message")
		}
	}
}

func (s *SessionRunner) handleIncomingMessage(incoming IncomingMessage) error {
	if err := incoming.Err(); err != nil {
		return errors.Wrap(err, "error receiving messages")
	}

	notes, ok := incoming.Notes()
	if ok {
		return s.handleIncomingNotes(notes)
	}

	msg, ok := incoming.Msg()
	if ok {
		if err := s.rawMessageHandler.Handle(msg); err != nil {
			return errors.Wrap(err, "error handling a message")
		}
		return nil
	}

	return errors.New("logic error")
}

func (s *SessionRunner) handleIncomingNotes(notes messages.EbtReplicateNotes) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for _, note := range notes.Notes() {
		s.remoteNotes[note.Ref().String()] = note
	}

	// todo initate sending messages
	return nil
}
