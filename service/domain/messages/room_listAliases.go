package messages

import (
	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	RoomListAliasesProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"room", "listAliases"}),
		rpc.ProcedureTypeAsync,
	)
)

func NewRoomListAliases(arguments RoomListAliasesArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		RoomListAliasesProcedure.Name(),
		RoomListAliasesProcedure.Typ(),
		j,
	)
}

type RoomListAliasesArguments struct {
	identity refs.Identity
}

func NewRoomListAliasesArguments(
	identity refs.Identity,
) (RoomListAliasesArguments, error) {
	if identity.IsZero() {
		return RoomListAliasesArguments{}, errors.New("zero value of identity")
	}

	return RoomListAliasesArguments{
		identity: identity,
	}, nil
}

func (i RoomListAliasesArguments) MarshalJSON() ([]byte, error) {
	return jsoniter.Marshal([]string{
		i.identity.String(),
	})
}

type RoomListAliasesResponse struct {
	aliases []aliases.Alias
}

func NewRoomListAliasesResponseFromBytes(b []byte) (RoomListAliasesResponse, error) {
	var aliasesAsStrings []string
	if err := jsoniter.Unmarshal(b, &aliasesAsStrings); err != nil {
		return RoomListAliasesResponse{}, errors.Wrap(err, "json unmarshal failed")
	}

	var aliasesSlice []aliases.Alias
	for _, aliasString := range aliasesAsStrings {
		alias, err := aliases.NewAlias(aliasString)
		if err != nil {
			return RoomListAliasesResponse{}, errors.Wrap(err, "error creating an alias")
		}
		aliasesSlice = append(aliasesSlice, alias)
	}

	return RoomListAliasesResponse{aliases: aliasesSlice}, nil
}

func (r RoomListAliasesResponse) Aliases() []aliases.Alias {
	return r.aliases
}
