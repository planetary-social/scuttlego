package transport

import (
	"io"
	"sync"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
)

type RawConnection struct {
	rwc    io.ReadWriteCloser
	lock   *sync.Mutex // lock guards against simultaneous writes to rw
	logger logging.Logger
}

func NewRawConnection(rwc io.ReadWriteCloser, logger logging.Logger) RawConnection {
	return RawConnection{
		rwc:    rwc,
		lock:   &sync.Mutex{},
		logger: logger.New("raw"),
	}
}

func (s RawConnection) Next() (*Message, error) {
	headerBuf := make([]byte, MessageHeaderSize)
	_, err := io.ReadFull(s.rwc, headerBuf[:])
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
	_, err = io.ReadFull(s.rwc, bodyBuf[:])
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

// Send marshals the message and writes it to the underlying writer. Send can be called from multiple goroutines.
func (s RawConnection) Send(msg *Message) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.loggerWithMessageFields(msg).Trace("sending a message")

	b, err := msg.Header.Bytes()
	if err != nil {
		return errors.Wrap(err, "failed to serialize the message header")
	}

	_, err = s.rwc.Write(b)
	if err != nil {
		return errors.Wrap(err, "failed to write the message header")
	}

	_, err = s.rwc.Write(msg.Body)
	if err != nil {
		return errors.Wrap(err, "failed to write the message body")
	}

	return nil
}

func (s RawConnection) Close() error {
	// todo send termination?
	return s.rwc.Close()
}

func (s RawConnection) loggerWithMessageFields(msg *Message) logging.Logger {
	l := s.logger.
		WithField("header.flags", msg.Header.Flags()).
		WithField("header.number", msg.Header.RequestNumber()).
		WithField("header.bodyLength", msg.Header.BodyLength())
	if msg.Header.Flags().BodyType() != MessageBodyTypeBinary {
		l = l.WithField("body", string(msg.Body))
	}
	return l
}

func isTermination(bytes []byte) bool {
	for _, b := range bytes {
		if b != 0 {
			return false
		}
	}
	return true
}
