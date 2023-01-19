package debugger

import (
	"strconv"
	"strings"

	"github.com/boreq/errors"
)

const (
	MessageSent     = "sending a message"
	MessageReceived = "received a message"

	FieldName         = "name"
	FieldMessage      = "message"
	FieldHeaderNumber = "header.number"
	FieldHeaderFlags  = "header.flags"
	FieldBody         = "body"

	FieldNameValueSuffix = ".raw"
)

type MessageType struct{ string }

var (
	MessageTypeReceived = MessageType{"received"}
	MessageTypeSent     = MessageType{"sent"}
)

type InitiatedBy struct{ string }

var (
	InitiatedByRemoteNode = InitiatedBy{"initiated_by_remote"}
	InitiatedByLocalNode  = InitiatedBy{"initiated_by_local"}
)

// Sessions uses session numbers as keys.
type Sessions map[int]Session

func NewSessions() Sessions {
	return make(Sessions)
}

func (s Sessions) AddMessage(message Message) error {
	sessionNumber, err := s.determineSession(message)
	if err != nil {
		return errors.Wrap(err, "error determining session number")
	}

	v, ok := s[sessionNumber]
	if !ok {
		v = Session{
			Number:       sessionNumber,
			InititatedBy: s.determineInitiatedBy(sessionNumber),
			Messages:     nil,
		}
	}

	v.Messages = append(v.Messages, message)
	s[sessionNumber] = v
	return nil
}

func (s Sessions) determineSession(message Message) (int, error) {
	switch message.Type {
	case MessageTypeReceived:
		return -message.RequestNumber, nil
	case MessageTypeSent:
		return message.RequestNumber, nil
	default:
		return 0, errors.New("unknown message type")
	}
}

func (s Sessions) determineInitiatedBy(sessionNumber int) InitiatedBy {
	if sessionNumber > 0 {
		return InitiatedByLocalNode
	} else {
		return InitiatedByRemoteNode
	}
}

type Session struct {
	// Number is the request number interpreted from the perspective of the
	// logging party. This means that streams initiated by the local node have
	// positive stream numbers and streams initiated by the remote have negative
	// numbers.
	Number       int
	InititatedBy InitiatedBy
	Messages     []Message
}

type Message struct {
	Type MessageType

	Flags         string
	RequestNumber int
	Body          string

	Entry Entry
}

func NewMessage(entry Entry) (Message, error) {
	messageType, err := parseMessageType(entry)
	if err != nil {
		return Message{}, errors.Wrap(err, "error parsing message type")
	}

	requestNumber, err := strconv.Atoi(entry[FieldHeaderNumber])
	if err != nil {
		return Message{}, errors.Wrap(err, "error parsing stream number")
	}

	return Message{
		Type: messageType,

		Flags:         entry[FieldHeaderFlags],
		RequestNumber: requestNumber,
		Body:          entry[FieldBody],

		Entry: entry,
	}, nil

}

func parseMessageType(entry Entry) (MessageType, error) {
	switch entry[FieldMessage] {
	case MessageSent:
		return MessageTypeSent, nil
	case MessageReceived:
		return MessageTypeReceived, nil
	default:
		return MessageType{}, errors.New("unknown message type")
	}
}

type Groups struct {
	Peers map[string]Sessions
}

func NewGroups() *Groups {
	return &Groups{
		Peers: make(map[string]Sessions),
	}
}

func (g *Groups) Add(e Entry) error {
	peer := e[FieldName]
	peer = strings.TrimSuffix(peer, FieldNameValueSuffix)

	if msg := e[FieldMessage]; msg == MessageSent || msg == MessageReceived {
		sessions, ok := g.Peers[peer]
		if !ok {
			sessions = NewSessions()
			g.Peers[peer] = sessions
		}

		message, err := NewMessage(e)
		if err != nil {
			return errors.Wrap(err, "error creating a message")
		}

		if err := sessions.AddMessage(message); err != nil {
			return errors.Wrap(err, "error adding a message")
		}
	}

	return nil
}
