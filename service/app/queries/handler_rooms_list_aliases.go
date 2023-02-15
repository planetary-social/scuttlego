package queries

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type RoomsListAliases struct {
	room    refs.Identity
	address network.Address
}

func NewRoomsListAliases(
	room refs.Identity,
	address network.Address,
) (RoomsListAliases, error) {
	if room.IsZero() {
		return RoomsListAliases{}, errors.New("zero value of room")
	}

	if address.IsZero() {
		return RoomsListAliases{}, errors.New("zero value of address")
	}

	return RoomsListAliases{
		room:    room,
		address: address,
	}, nil
}

func (r RoomsListAliases) Room() refs.Identity {
	return r.room
}

func (r RoomsListAliases) Address() network.Address {
	return r.address
}

func (r RoomsListAliases) IsZero() bool {
	return r.room.IsZero()
}

type RoomsListAliasesHandler struct {
	local  refs.Identity
	dialer Dialer
}

func NewRoomsListAliasesHandler(
	dialer Dialer,
	local identity.Public,
) (*RoomsListAliasesHandler, error) {
	identityRef, err := refs.NewIdentityFromPublic(local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a ref")
	}

	return &RoomsListAliasesHandler{
		dialer: dialer,
		local:  identityRef,
	}, nil
}

func (h *RoomsListAliasesHandler) Handle(ctx context.Context, query RoomsListAliases) ([]aliases.Alias, error) {
	if query.IsZero() {
		return nil, errors.New("zero value of query")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	peer, err := h.dialer.Dial(ctx, query.Room().Identity(), query.Address())
	if err != nil {
		return nil, errors.Wrap(err, "dial error")
	}

	req, err := h.createRequest()
	if err != nil {
		return nil, errors.Wrap(err, "could not create a request")
	}

	rs, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to perform a request")
	}

	response, ok := <-rs.Channel()
	if !ok {
		return nil, errors.New("channel closed")
	}

	if err := response.Err; err != nil {
		return nil, errors.Wrap(err, "received an error")
	}

	listAliasesResponse, err := messages.NewRoomListAliasesResponseFromBytes(response.Value.Bytes())
	if err != nil {
		return nil, errors.Wrap(err, "error creating the response")
	}

	return listAliasesResponse.Aliases(), nil
}

func (h *RoomsListAliasesHandler) createRequest() (*rpc.Request, error) {
	args, err := messages.NewRoomListAliasesArguments(h.local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create args")
	}

	req, err := messages.NewRoomListAliases(args)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the request")
	}

	return req, nil
}
