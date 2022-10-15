package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	RoomRevokeAliasProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"room", "revokeAlias"}),
		rpc.ProcedureTypeAsync,
	)
)

func NewRoomRevokeAlias(arguments RoomRevokeAliasArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		RoomRevokeAliasProcedure.Name(),
		RoomRevokeAliasProcedure.Typ(),
		j,
	)
}

type RoomRevokeAliasArguments struct {
	alias aliases.Alias
}

func NewRoomRevokeAliasArguments(
	alias aliases.Alias,
) (RoomRevokeAliasArguments, error) {
	if alias.IsZero() {
		return RoomRevokeAliasArguments{}, errors.New("zero value of alias")
	}

	return RoomRevokeAliasArguments{
		alias: alias,
	}, nil
}

func (i RoomRevokeAliasArguments) MarshalJSON() ([]byte, error) {
	return json.Marshal([]string{
		i.alias.String(),
	})
}
