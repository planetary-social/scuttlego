package commands

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/network"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type RoomsAliasRevoke struct {
	room    refs.Identity
	address network.Address
	alias   aliases.Alias
}

func NewRoomsAliasRevoke(
	room refs.Identity,
	address network.Address,
	alias aliases.Alias,
) (RoomsAliasRevoke, error) {
	if room.IsZero() {
		return RoomsAliasRevoke{}, errors.New("zero value of room")
	}

	if address.IsZero() {
		return RoomsAliasRevoke{}, errors.New("zero value of address")
	}

	if alias.IsZero() {
		return RoomsAliasRevoke{}, errors.New("zero value of alias")
	}

	return RoomsAliasRevoke{
		room:    room,
		address: address,
		alias:   alias,
	}, nil
}

func (r RoomsAliasRevoke) Room() refs.Identity {
	return r.room
}

func (r RoomsAliasRevoke) Address() network.Address {
	return r.address
}

func (r RoomsAliasRevoke) Alias() aliases.Alias {
	return r.alias
}

func (r RoomsAliasRevoke) IsZero() bool {
	return r.room.IsZero()
}

type RoomsAliasRevokeHandler struct {
	dialer Dialer
}

func NewRoomsAliasRevokeHandler(
	dialer Dialer,
) *RoomsAliasRevokeHandler {
	return &RoomsAliasRevokeHandler{
		dialer: dialer,
	}
}

func (h *RoomsAliasRevokeHandler) Handle(ctx context.Context, cmd RoomsAliasRevoke) error {
	if cmd.IsZero() {
		return errors.New("zero value of command")
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	peer, err := h.dialer.Dial(ctx, cmd.Room().Identity(), cmd.Address())
	if err != nil {
		return errors.Wrap(err, "dial error")
	}

	req, err := h.createRequest(cmd.Alias())
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

	rs, err := peer.Conn().PerformRequest(ctx, req)
	if err != nil {
		return errors.Wrap(err, "failed to perform a request")
	}

	response, ok := <-rs.Channel()
	if !ok {
		return errors.New("channel closed")
	}

	if err := response.Err; err != nil {
		return errors.Wrap(err, "received an error")
	}

	return nil
}

func (h *RoomsAliasRevokeHandler) createRequest(
	alias aliases.Alias,
) (*rpc.Request, error) {
	args, err := messages.NewRoomRevokeAliasArguments(alias)
	if err != nil {
		return nil, errors.Wrap(err, "could not create args")
	}

	req, err := messages.NewRoomRevokeAlias(args)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the request")
	}

	return req, nil
}
