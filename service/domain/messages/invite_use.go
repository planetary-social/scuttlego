package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	InviteUseProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"invite", "use"}),
		rpc.ProcedureTypeAsync,
	)
)

func NewInviteUse(arguments InviteUseArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		InviteUseProcedure.Name(),
		InviteUseProcedure.Typ(),
		j,
	)
}

type InviteUseArguments struct {
	feed refs.Identity // todo feed or identity? specification seems to be confused
}

func NewInviteUseArguments(feed refs.Identity) (InviteUseArguments, error) {
	if feed.IsZero() {
		return InviteUseArguments{}, errors.New("zero value of feed")
	}
	return InviteUseArguments{feed: feed}, nil
}

func NewInviteUseArgumentsFromBytes(b []byte) (InviteUseArguments, error) {
	var args []inviteUseArgumentsTransport

	if err := json.Unmarshal(b, &args); err != nil {
		return InviteUseArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return InviteUseArguments{}, errors.New("expected exactly one argument")
	}

	feed, err := refs.NewIdentity(args[0].Feed)
	if err != nil {
		return InviteUseArguments{}, errors.Wrap(err, "could not create an identity ref")
	}

	return NewInviteUseArguments(feed)
}

func (i InviteUseArguments) MarshalJSON() ([]byte, error) {
	return json.Marshal([]inviteUseArgumentsTransport{
		{
			Feed: i.feed.String(),
		},
	})
}

type inviteUseArgumentsTransport struct {
	Feed string `json:"feed"`
}
