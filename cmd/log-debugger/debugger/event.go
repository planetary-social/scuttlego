package debugger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/cmd/log-debugger/debugger/log"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
)

const (
	TimestampFormat = "2006-01-02 15:04:05.999999999 (MST)"

	FieldTimestamp    = "time"
	FieldMessage      = "msg"
	FieldPeerId       = logging.PeerIdContextLabel
	FieldConnectionId = logging.ConnectionIdContextLabel
	FieldStreamId     = logging.StreamIdContextLabel

	MessageLogFieldRequestNumber = "header.number"
	MessageLogFieldFlags         = "header.flags"
	MessageLogFieldBody          = "body"
	MessageLogMessageSent        = "sending a message"
	MessageLogMessageReceived    = "received a message"
)

type MessageType struct{ string }

var (
	MessageTypeReceived = MessageType{"received"}
	MessageTypeSent     = MessageType{"sent"}
)

type Event struct {
	Timestamp time.Time

	PeerId       string
	ConnectionId string
	StreamId     string

	Message *Message
	Entry   log.Entry
}

func NewEvent(e log.Entry) (*Event, error) {
	peerIdString := e[FieldPeerId]
	connectionIdString := e[FieldConnectionId]

	if peerIdString == "" && connectionIdString == "" {
		return nil, nil
	}

	if peerIdString == "" || connectionIdString == "" {
		return nil, errors.New("missing peer or connection id")
	}

	timestamp, err := time.Parse(TimestampFormat, e[FieldTimestamp])
	if err != nil {
		return nil, errors.Wrap(err, "error parsing the timestamp")
	}

	event := Event{
		Timestamp: timestamp,

		PeerId:       peerIdString,
		ConnectionId: connectionIdString,

		Entry: e,
	}

	if msg := e[FieldMessage]; msg == MessageLogMessageSent || msg == MessageLogMessageReceived {
		msg, err := NewMessage(e)
		if err != nil {
			return nil, errors.Wrap(err, "error creating a message")
		}

		streamIdInt, err := determineMessageStreamId(msg)
		if err != nil {
			return nil, errors.Wrap(err, "error determining message stream id")
		}

		event.StreamId = strconv.Itoa(streamIdInt)
		event.Message = &msg
	} else {
		streamIdString := e[FieldStreamId]
		if streamIdString == "" {
			return nil, nil
		}

		event.StreamId = streamIdString
	}

	return &event, nil
}

type Message struct {
	Type          MessageType
	Flags         string
	RequestNumber int
	Body          string
}

func NewMessage(entry log.Entry) (Message, error) {
	messageType, err := parseMessageType(entry)
	if err != nil {
		return Message{}, errors.Wrap(err, "error parsing message type")
	}

	requestNumber, err := strconv.Atoi(entry[MessageLogFieldRequestNumber])
	if err != nil {
		return Message{}, errors.Wrap(err, "error parsing stream number")
	}

	body, err := prettifyBody(entry[MessageLogFieldBody])
	if err != nil {
		return Message{}, errors.Wrap(err, "error pretifying the body")
	}

	return Message{
		Type:          messageType,
		Flags:         entry[MessageLogFieldFlags],
		RequestNumber: requestNumber,
		Body:          body,
	}, nil
}

func prettifyBody(body string) (string, error) {
	decoded, err := messages.NewEbtReplicateNotesFromBytes([]byte(body))
	if err == nil {
		annotatedNotes := make(map[string]string)

		for _, note := range decoded.Notes() {
			b, err := note.MarshalJSON()
			if err != nil {
				return "", errors.Wrap(err, "error marshaling a note")
			}

			annotatedNotes[note.Ref().String()] = fmt.Sprintf("%s (sequence=%d, receive=%t replicate=%t)", string(b), note.Sequence(), note.Receive(), note.Replicate())
		}

		j, err := json.Marshal(annotatedNotes)
		if err != nil {
			return "", errors.Wrap(err, "error marshaling annotated notes")
		}

		body = string(j)
	}

	bodyBuf := &bytes.Buffer{}
	if err := json.Indent(bodyBuf, []byte(body), "", "    "); err == nil {
		return bodyBuf.String(), nil
	}

	return body, nil
}

func parseMessageType(entry log.Entry) (MessageType, error) {
	switch entry[FieldMessage] {
	case MessageLogMessageSent:
		return MessageTypeSent, nil
	case MessageLogMessageReceived:
		return MessageTypeReceived, nil
	default:
		return MessageType{}, errors.New("unknown message type")
	}
}

func determineMessageStreamId(msg Message) (int, error) {
	switch msg.Type {
	case MessageTypeReceived:
		return -msg.RequestNumber, nil
	case MessageTypeSent:
		return msg.RequestNumber, nil
	default:
		return 0, errors.New("unknown message type")
	}
}
