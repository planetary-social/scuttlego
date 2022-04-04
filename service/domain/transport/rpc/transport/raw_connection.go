package transport

import (
	"io"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type RawConnection struct {
	rw     io.ReadWriteCloser
	logger logging.Logger
}

func NewRawConnection(rw io.ReadWriteCloser, logger logging.Logger) RawConnection {
	return RawConnection{
		rw:     rw,
		logger: logger.New("raw"),
	}
}

func (s RawConnection) Next() (*Message, error) {
	headerBuf := make([]byte, MessageHeaderSize)
	_, err := io.ReadFull(s.rw, headerBuf[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not read the message header")
	}

	if isTermination(headerBuf) {
		return nil, errors.New("other side has terminated the connection")
	}

	messageHeader, err := NewMessageHeaderFromBytes(headerBuf)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message header")
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

	s.loggerWithMessageFields(&msg).Trace("received a message")

	return &msg, nil
}

func (s RawConnection) Send(msg *Message) error {
	s.loggerWithMessageFields(msg).Trace("sending a message")

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

func (s RawConnection) loggerWithMessageFields(msg *Message) logging.Logger {
	return s.logger.
		WithField("header.flags", msg.Header.Flags()).
		WithField("header.number", msg.Header.RequestNumber()).
		WithField("header.bodyLength", msg.Header.BodyLength()).
		WithField("body", string(msg.Body))
}

func isTermination(bytes []byte) bool {
	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}
	return true
}
