package rooms

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type PeerRPCAdapter struct {
	logger logging.Logger
}

func NewPeerRPCAdapter(logger logging.Logger) *PeerRPCAdapter {
	return &PeerRPCAdapter{
		logger: logger.New("rooms_peer_rpc_adapter"),
	}
}

func (a *PeerRPCAdapter) GetMetadata(ctx context.Context, peer transport.Peer) (messages.RoomMetadataResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	req, err := messages.NewRoomMetadata()
	if err != nil {
		return messages.RoomMetadataResponse{}, errors.Wrap(err, "could not create a request")
	}

	stream, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return messages.RoomMetadataResponse{}, errors.Wrap(err, "could not perform a request")
	}

	for v := range stream.Channel() {
		if err := v.Err; err != nil {
			return messages.RoomMetadataResponse{}, errors.Wrap(err, "received an error")
		}

		metadataResponse, err := messages.NewRoomMetadataResponse(v.Value.Bytes())
		if err != nil {
			return messages.RoomMetadataResponse{}, errors.Wrap(err, "could not parse the response")
		}

		return metadataResponse, nil
	}

	return messages.RoomMetadataResponse{}, errors.New("received no responses")
}

func (a *PeerRPCAdapter) GetAttendants(ctx context.Context, peer transport.Peer) (<-chan RoomAttendantsEvent, error) {
	req, err := messages.NewRoomAttendants()
	if err != nil {
		return nil, errors.Wrap(err, "could not create a request")
	}

	stream, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "could not perform a request")
	}

	ch := make(chan RoomAttendantsEvent)

	go func() {
		defer close(ch)

		if err := a.streamAttendants(ctx, ch, stream); err != nil {
			a.logger.WithError(err).Debug("attendants stream error")
		}
	}()

	return ch, nil
}

func (a *PeerRPCAdapter) streamAttendants(ctx context.Context, ch chan RoomAttendantsEvent, stream rpc.ResponseStream) error {
	v, ok := <-stream.Channel()
	if !ok {
		return errors.New("stream channel closed before getting the first message")
	}

	if err := v.Err; err != nil {
		return errors.Wrap(err, "remote error")
	}

	events, err := a.parseFirstGetAttendantsMessage(v.Value.Bytes())
	if err != nil {
		return errors.Wrap(err, "error parsing the first message")
	}

	for _, event := range events {
		select {
		case ch <- event:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	for v := range stream.Channel() {
		if err := v.Err; err != nil {
			return errors.Wrap(err, "remote error")
		}

		event, err := a.parseNextGetAttendantsMessage(v.Value.Bytes())
		if err != nil {
			return errors.Wrap(err, "error parsing follow up messages")
		}

		select {
		case ch <- event:
			continue
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return errors.New("stream channel closed")
}

func (a *PeerRPCAdapter) parseFirstGetAttendantsMessage(v []byte) ([]RoomAttendantsEvent, error) {
	msg, err := messages.NewRoomAttendantsResponseState(v)
	if err != nil {
		return nil, errors.Wrap(err, "error parsing the message")
	}

	var result []RoomAttendantsEvent

	for _, ref := range msg.Ids() {
		event, err := NewRoomAttendantsEvent(RoomAttendantsEventTypeJoined, ref)
		if err != nil {
			return nil, errors.Wrap(err, "error creating the event")
		}

		result = append(result, event)
	}

	return result, nil
}

func (a *PeerRPCAdapter) parseNextGetAttendantsMessage(v []byte) (RoomAttendantsEvent, error) {
	msg, err := messages.NewRoomAttendantsResponseJoinedOrLeft(v)
	if err != nil {
		return RoomAttendantsEvent{}, errors.Wrap(err, "error parsing the message")
	}

	typ, err := NewRoomAttendantsEventTypeFromRoomAttendantsReponseType(msg.Typ())
	if err != nil {
		return RoomAttendantsEvent{}, errors.Wrap(err, "error converting the type")
	}

	event, err := NewRoomAttendantsEvent(typ, msg.Id())
	if err != nil {
		return RoomAttendantsEvent{}, errors.Wrap(err, "error creating the event")
	}

	return event, nil
}

type RoomAttendantsEvent struct {
	typ RoomAttendantsEventType
	id  refs.Identity
}

func NewRoomAttendantsEvent(typ RoomAttendantsEventType, id refs.Identity) (RoomAttendantsEvent, error) {
	if typ.IsZero() {
		return RoomAttendantsEvent{}, errors.New("zero value of typ")
	}
	if id.IsZero() {
		return RoomAttendantsEvent{}, errors.New("zero value of id")
	}
	return RoomAttendantsEvent{
		typ: typ,
		id:  id,
	}, nil
}

func (e *RoomAttendantsEvent) IsZero() bool {
	return e.id.IsZero()
}

type RoomAttendantsEventType struct {
	s string
}

func NewRoomAttendantsEventTypeFromRoomAttendantsReponseType(v messages.RoomAttendantsResponseType) (RoomAttendantsEventType, error) {
	switch v {
	case messages.RoomAttendantsResponseTypeJoined:
		return RoomAttendantsEventTypeJoined, nil
	case messages.RoomAttendantsResponseTypeLeft:
		return RoomAttendantsEventTypeLeft, nil
	default:
		return RoomAttendantsEventType{}, errors.New("unknown response type")
	}
}

func (t RoomAttendantsEventType) IsZero() bool {
	return t == RoomAttendantsEventType{}
}

var (
	RoomAttendantsEventTypeJoined = RoomAttendantsEventType{"joined"}
	RoomAttendantsEventTypeLeft   = RoomAttendantsEventType{"left"}
)
