package ebt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/messages"
)

type EbtStream interface {
	IncomingMessages() <-chan IncomingMessage
	SendNote(note Note)
	SendMessage(msg *message.Message)
}

type IncomingMessage struct {
	note messages.EbtReplicateNote
	msg  message.RawMessage
	err  error
}

func NewIncomingMessageWithNote(note messages.EbtReplicateNote) IncomingMessage {
	return IncomingMessage{
		note: note,
	}
}

func NewIncomingMessageWithMesage(msg message.RawMessage) IncomingMessage {
	return IncomingMessage{
		msg: msg,
	}
}

func NewIncomingMessageWithErr(err error) IncomingMessage {
	return IncomingMessage{
		err: err,
	}
}

func (i IncomingMessage) Note() messages.EbtReplicateNote {
	return i.note
}

func (i IncomingMessage) Msg() message.RawMessage {
	return i.msg
}

func (i IncomingMessage) Err() error {
	return i.err
}

type SessionRunner struct {
}

func (s *SessionRunner) HandleStream(stream EbtStream) error {
	for incoming := range stream.IncomingMessages() {
		if err := incoming.Err(); err != nil {
			return errors.Wrap(err, "error receiving messages")
		}

		// process incoming messages
	}

	return nil
}
