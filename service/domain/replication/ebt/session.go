package ebt

import (
	"context"
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"sync"
)

type RawMessageHandler interface {
	Handle(msg message.RawMessage) error
}

type EbtStream interface {
	IncomingMessages(ctx context.Context) <-chan IncomingMessage
	SendNote(note messages.EbtReplicateNote)
	SendMessage(msg *message.Message)
}

type IncomingMessage struct {
	note *messages.EbtReplicateNote
	msg  *message.RawMessage
	err  error
}

func NewIncomingMessageWithNote(note messages.EbtReplicateNote) IncomingMessage {
	return IncomingMessage{
		note: &note,
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

func (i IncomingMessage) Note() (messages.EbtReplicateNote, bool) {
	if i.note != nil {
		return *i.note, true
	}
	return messages.EbtReplicateNote{}, false
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
	lock sync.Mutex // guards remoteWants

	logger            logging.Logger
	rawMessageHandler RawMessageHandler
}

func NewSessionRunner() *SessionRunner {
	return &SessionRunner{}
}

func (s *SessionRunner) HandleStream(ctx context.Context, stream EbtStream) error {
	go s.handleIncomingMessages(ctx, stream)

	// todo send

	return nil
}

func (s *SessionRunner) handleIncomingMessages(ctx context.Context, stream EbtStream) {
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

	note, ok := incoming.Note()
	if ok {
		return s.handleIncomingNote(note)
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

func (s *SessionRunner) handleIncomingNote(note messages.EbtReplicateNote) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.remoteNotes[note.]

	// todo send messages
}
