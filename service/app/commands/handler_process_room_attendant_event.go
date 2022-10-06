package commands

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/rooms"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type ProcessRoomAttendantEvent struct {
	portal transport.Peer
	event  rooms.RoomAttendantsEvent
}

func NewProcessRoomAttendantEvent(
	portal transport.Peer,
	event rooms.RoomAttendantsEvent,
) (ProcessRoomAttendantEvent, error) {
	if portal.IsZero() {
		return ProcessRoomAttendantEvent{}, errors.New("zero value of portal")
	}
	if event.IsZero() {
		return ProcessRoomAttendantEvent{}, errors.New("zero value of event")
	}
	return ProcessRoomAttendantEvent{
		portal: portal,
		event:  event,
	}, nil
}

func (e ProcessRoomAttendantEvent) IsZero() bool {
	return e.event.IsZero()
}

type ProcessRoomAttendantEventHandler struct {
	peerManager PeerManager
}

func NewProcessRoomAttendantEventHandler(peerManager PeerManager) *ProcessRoomAttendantEventHandler {
	return &ProcessRoomAttendantEventHandler{peerManager: peerManager}
}

func (h *ProcessRoomAttendantEventHandler) Handle(cmd ProcessRoomAttendantEvent) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")

	}
	fmt.Println("application", cmd)

	return nil
}
