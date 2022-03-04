package transport

import (
	"io"

	"github.com/boreq/errors"
)

type RawConnection struct {
	rw io.ReadWriteCloser
}

func NewRawConnection(rw io.ReadWriteCloser) RawConnection {
	return RawConnection{
		rw: rw,
	}
}

func (s RawConnection) Next() (*Message, error) {
	headerBuf := make([]byte, MessageHeaderSize)
	_, err := io.ReadFull(s.rw, headerBuf[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not read the message header")
	}

	messageHeader, err := NewMessageHeaderFromBytes(headerBuf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the message header")
	}

	bodyBuf := make([]byte, messageHeader.bodyLength)
	_, err = io.ReadFull(s.rw, bodyBuf[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not read the message body")
	}

	msg, err := NewMessage(messageHeader, bodyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message")
	}

	return &msg, nil
}

func (s RawConnection) Send(msg *Message) error {
	b, err := msg.Header.Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to serialize the message header")
	}

	_, err = s.rw.Write(b)
	if err != nil {
		return errors.Wrap(err, "failed to write the message header")
	}

	_, err = s.rw.Write(msg.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write the message body")
	}

	return nil
}

func (s RawConnection) Close() error {
	return s.rw.Close()
}
