package transport

import (
	"io"

	"github.com/planetary-social/go-ssb/logging"

	"github.com/boreq/errors"
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

	messageHeader, err := NewMessageHeaderFromBytes(headerBuf)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read the message header")
	}

	bodyBuf := make([]byte, messageHeader.bodyLength)
	_, err = io.ReadFull(s.rw, bodyBuf[:])
	if err != nil {
		return nil, errors.Wrap(err, "could not read the message body")
	}

	//s.logger.WithField("body", string(bodyBuf)).Debug("receivedMessage")

	msg, err := NewMessage(messageHeader, bodyBuf)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a message")
	}

	return &msg, nil
}

func (s RawConnection) Send(msg *Message) error {
	//s.logger.WithField("body", string(msg.Body)).Debug("sending message")

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
