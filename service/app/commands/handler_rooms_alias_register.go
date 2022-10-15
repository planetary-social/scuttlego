package commands

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

type RoomsAliasRegister struct {
	room    refs.Identity
	address network.Address
	alias   aliases.Alias
}

func NewRoomsAliasRegister(
	room refs.Identity,
	address network.Address,
	alias aliases.Alias,
) (RoomsAliasRegister, error) {
	if room.IsZero() {
		return RoomsAliasRegister{}, errors.New("zero value of room")
	}

	if address.IsZero() {
		return RoomsAliasRegister{}, errors.New("zero value of address")
	}

	if alias.IsZero() {
		return RoomsAliasRegister{}, errors.New("zero value of alias")
	}

	return RoomsAliasRegister{
		room:    room,
		address: address,
		alias:   alias,
	}, nil
}

func (r RoomsAliasRegister) Room() refs.Identity {
	return r.room
}

func (r RoomsAliasRegister) Address() network.Address {
	return r.address
}

func (r RoomsAliasRegister) Alias() aliases.Alias {
	return r.alias
}

func (r RoomsAliasRegister) IsZero() bool {
	return r.room.IsZero()
}

type RoomsAliasRegisterHandler struct {
	dialer Dialer
	local  identity.Private
}

func NewRoomsAliasRegisterHandler(
	dialer Dialer,
	local identity.Private,
) *RoomsAliasRegisterHandler {
	return &RoomsAliasRegisterHandler{
		dialer: dialer,
		local:  local,
	}
}

func (h *RoomsAliasRegisterHandler) Handle(ctx context.Context, cmd RoomsAliasRegister) (aliases.AliasEndpointURL, error) {
	if cmd.IsZero() {
		return aliases.AliasEndpointURL{}, errors.New("zero value of command")
	}

	user, err := refs.NewIdentityFromPublic(h.local.Public())
	if err != nil {
		return aliases.AliasEndpointURL{}, errors.Wrap(err, "failed to create user ref")
	}

	msg, err := aliases.NewRegistrationMessage(cmd.Alias(), user, cmd.Room())
	if err != nil {
		return aliases.AliasEndpointURL{}, errors.Wrap(err, "error creating a registration")
	}

	signature, err := aliases.NewRegistrationSignature(msg, h.local)
	if err != nil {
		return aliases.AliasEndpointURL{}, errors.Wrap(err, "error creating a signature")
	}

	response, err := h.registerAlias(ctx, cmd, signature)
	if err != nil {
		return aliases.AliasEndpointURL{}, errors.Wrap(err, "could not contact the pub and redeem the invite")
	}

	registerAliasResponse, err := messages.NewRoomRegisterAliasResponseFromBytes(response.Bytes())
	if err != nil {
		return aliases.AliasEndpointURL{}, errors.Wrap(err, "error creating the response")
	}

	return registerAliasResponse.AliasEndpointURL(), nil
}

func (h *RoomsAliasRegisterHandler) registerAlias(ctx context.Context, cmd RoomsAliasRegister, signature aliases.RegistrationSignature) (*rpc.Response, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	peer, err := h.dialer.Dial(ctx, cmd.Room().Identity(), cmd.Address())
	if err != nil {
		return nil, errors.Wrap(err, "dial error")
	}

	req, err := h.createRequest(cmd.Alias(), signature)
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

	return response.Value, nil
}

func (h *RoomsAliasRegisterHandler) createRequest(
	alias aliases.Alias,
	signature aliases.RegistrationSignature,
) (*rpc.Request, error) {
	args, err := messages.NewRoomRegisterAliasArguments(alias, signature)
	if err != nil {
		return nil, errors.Wrap(err, "could not create args")
	}

	req, err := messages.NewRoomRegisterAlias(args)
	if err != nil {
		return nil, errors.Wrap(err, "could not create the request")
	}

	return req, nil
}
