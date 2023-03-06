package messages

import (
	"fmt"

	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	RoomAttendantsProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"room", "attendants"}),
		rpc.ProcedureTypeSource,
	)
)

func NewRoomAttendants() (*rpc.Request, error) {
	return rpc.NewRequest(
		RoomAttendantsProcedure.Name(),
		RoomAttendantsProcedure.Typ(),
		[]byte("[]"),
	)
}

type RoomAttendantsResponseState struct {
	ids []refs.Identity
}

func NewRoomAttendantsResponseStateFromBytes(b []byte) (RoomAttendantsResponseState, error) {
	var transport roomAttendantsResponseStateTransport
	if err := jsoniter.Unmarshal(b, &transport); err != nil {
		return RoomAttendantsResponseState{}, errors.Wrap(err, "json unmarshal failed")
	}

	if transport.Type != "state" {
		return RoomAttendantsResponseState{}, errors.New("invalid response type")
	}

	var refsSlice []refs.Identity

	for _, refString := range transport.Ids {
		ref, err := refs.NewIdentity(refString)
		if err != nil {
			return RoomAttendantsResponseState{}, errors.Wrap(err, "error creating a ref")
		}

		refsSlice = append(refsSlice, ref)
	}

	return RoomAttendantsResponseState{
		ids: refsSlice,
	}, nil
}

func (r RoomAttendantsResponseState) Ids() []refs.Identity {
	return r.ids
}

type RoomAttendantsResponseJoinedOrLeft struct {
	typ RoomAttendantsResponseType
	id  refs.Identity
}

func NewRoomAttendantsResponseJoinedOrLeftFromBytes(b []byte) (RoomAttendantsResponseJoinedOrLeft, error) {
	var transport roomAttendantsResponseJoinedOrLeftTransport
	if err := jsoniter.Unmarshal(b, &transport); err != nil {
		return RoomAttendantsResponseJoinedOrLeft{}, errors.Wrap(err, "json unmarshal failed")
	}

	typ, err := decodeRoomAttendantsResponseType(transport.Type)
	if err != nil {
		return RoomAttendantsResponseJoinedOrLeft{}, errors.Wrap(err, "error decoding response type")
	}

	ref, err := refs.NewIdentity(transport.Id)
	if err != nil {
		return RoomAttendantsResponseJoinedOrLeft{}, errors.Wrap(err, "error creating a ref")
	}

	return RoomAttendantsResponseJoinedOrLeft{
		typ: typ,
		id:  ref,
	}, nil
}

func (r RoomAttendantsResponseJoinedOrLeft) Typ() RoomAttendantsResponseType {
	return r.typ
}

func (r RoomAttendantsResponseJoinedOrLeft) Id() refs.Identity {
	return r.id
}

type roomAttendantsResponseStateTransport struct {
	Type string   `json:"type"`
	Ids  []string `json:"ids"`
}

type roomAttendantsResponseJoinedOrLeftTransport struct {
	Type string `json:"type"`
	Id   string `json:"id"`
}

type RoomAttendantsResponseType struct {
	s string
}

var (
	RoomAttendantsResponseTypeJoined = RoomAttendantsResponseType{"joined"}
	RoomAttendantsResponseTypeLeft   = RoomAttendantsResponseType{"left"}
)

func decodeRoomAttendantsResponseType(s string) (RoomAttendantsResponseType, error) {
	switch s {
	case "joined":
		return RoomAttendantsResponseTypeJoined, nil
	case "left":
		return RoomAttendantsResponseTypeLeft, nil
	default:
		return RoomAttendantsResponseType{}, fmt.Errorf("unknown type: '%s'", s)
	}
}
