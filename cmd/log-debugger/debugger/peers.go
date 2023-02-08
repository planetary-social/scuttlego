package debugger

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/cmd/log-debugger/debugger/log"
)

// Peers uses peer ids as keys.
type Peers map[string]Connections

func NewPeers() Peers {
	return make(map[string]Connections)
}

func (p Peers) Add(e log.Entry) error {
	event, err := NewEvent(e)
	if err != nil {
		return errors.Wrap(err, "error creating a new event")
	}

	if event == nil {
		return nil
	}

	connections, ok := p[event.PeerId]
	if !ok {
		connections = NewConnections()
		p[event.PeerId] = connections
	}

	if err := connections.AddEvent(*event); err != nil {
		return errors.Wrap(err, "error adding an event")
	}

	return nil
}

// Connections uses connection ids as keys.
type Connections map[string]Connection

func NewConnections() Connections {
	return make(Connections)
}

func (s Connections) AddEvent(event Event) error {
	connection, ok := s[event.ConnectionId]
	if !ok {
		connection = NewConnection()
		s[event.ConnectionId] = connection
	}

	return connection.AddEvent(event)
}

// Connection uses stream ids as keys.
type Connection map[string]Stream

func NewConnection() Connection {
	return make(Connection)
}

func (s Connection) AddEvent(event Event) error {
	stream, ok := s[event.StreamId]
	if !ok {
		stream = NewStream()
	}

	stream.Events = append(stream.Events, event)

	s[event.StreamId] = stream
	return nil
}

type Stream struct {
	Events []Event
}

func NewStream() Stream {
	return Stream{}
}
