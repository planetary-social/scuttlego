package messages

import (
	"encoding/base64"
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	RoomRegisterAliasProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"room", "registerAlias"}),
		rpc.ProcedureTypeAsync,
	)
)

func NewRoomRegisterAlias(arguments RoomRegisterAliasArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		RoomRegisterAliasProcedure.Name(),
		RoomRegisterAliasProcedure.Typ(),
		j,
	)
}

type RoomRegisterAliasArguments struct {
	alias     aliases.Alias
	signature aliases.RegistrationSignature
}

func NewRoomRegisterAliasArguments(
	alias aliases.Alias,
	signature aliases.RegistrationSignature,
) (RoomRegisterAliasArguments, error) {
	if alias.IsZero() {
		return RoomRegisterAliasArguments{}, errors.New("zero value of alias")
	}

	if signature.IsZero() {
		return RoomRegisterAliasArguments{}, errors.New("zero value of signature")
	}

	return RoomRegisterAliasArguments{
		alias:     alias,
		signature: signature,
	}, nil
}

func (i RoomRegisterAliasArguments) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{
		i.alias.String(),
		base64.StdEncoding.EncodeToString(i.signature.Bytes()) + ".sig.ed25519",
	})
}

type RoomRegisterAliasResponse struct {
	url aliases.AliasEndpointURL
}

func NewRoomRegisterAliasResponseFromBytes(b []byte) (RoomRegisterAliasResponse, error) {
	url, err := aliases.NewAliasEndpointURL(string(b))
	if err != nil {
		return RoomRegisterAliasResponse{}, errors.Wrap(err, "could not create an endpoint url")
	}
	return RoomRegisterAliasResponse{url: url}, nil
}

func (r RoomRegisterAliasResponse) AliasEndpointURL() aliases.AliasEndpointURL {
	return r.url
}
