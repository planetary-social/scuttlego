package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	TunnelConnectProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"tunnel", "connect"}),
		rpc.ProcedureTypeDuplex,
	)
)

func NewTunnelConnectToPortal(arguments TunnelConnectToPortalArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		TunnelConnectProcedure.Name(),
		TunnelConnectProcedure.Typ(),
		j,
	)
}

type TunnelConnectToPortalArguments struct {
	portal refs.Identity
	target refs.Identity
}

func NewTunnelConnectToPortalArguments(
	portal refs.Identity,
	target refs.Identity,
) (TunnelConnectToPortalArguments, error) {
	if portal.IsZero() {
		return TunnelConnectToPortalArguments{}, errors.New("zero value of portal identity")
	}

	if target.IsZero() {
		return TunnelConnectToPortalArguments{}, errors.New("zero value of target identity")
	}

	return TunnelConnectToPortalArguments{
		portal: portal,
		target: target,
	}, nil
}

func (i TunnelConnectToPortalArguments) MarshalJSON() ([]byte, error) {
	return json.Marshal([]tunnelConnectToPortalArgumentsTransport{
		{
			Portal: i.portal.String(),
			Target: i.target.String(),
		},
	})
}

type tunnelConnectToPortalArgumentsTransport struct {
	Portal string `json:"portal"`
	Target string `json:"target"`
}
