package messages

import (
	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
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

func NewTunnelConnectToTarget(arguments TunnelConnectToTargetArguments) (*rpc.Request, error) {
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
	return jsoniter.Marshal([]tunnelConnectToPortalArgumentsTransport{
		{
			Portal: i.portal.String(),
			Target: i.target.String(),
		},
	})
}

type TunnelConnectToTargetArguments struct {
	portal refs.Identity
	target refs.Identity
	origin refs.Identity
}

func NewTunnelConnectToTargetArguments(
	portal refs.Identity,
	target refs.Identity,
	origin refs.Identity,
) (TunnelConnectToTargetArguments, error) {
	if portal.IsZero() {
		return TunnelConnectToTargetArguments{}, errors.New("zero value of portal identity")
	}

	if target.IsZero() {
		return TunnelConnectToTargetArguments{}, errors.New("zero value of target identity")
	}

	if origin.IsZero() {
		return TunnelConnectToTargetArguments{}, errors.New("zero value of origin identity")
	}

	return TunnelConnectToTargetArguments{
		portal: portal,
		target: target,
		origin: origin,
	}, nil
}

func NewTunnelConnectToTargetArgumentsFromBytes(b []byte) (TunnelConnectToTargetArguments, error) {
	var args []tunnelConnectToTargetArgumentsTransport
	if err := jsoniter.Unmarshal(b, &args); err != nil {
		return TunnelConnectToTargetArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return TunnelConnectToTargetArguments{}, errors.New("expected exactly one argument")
	}

	portal, err := refs.NewIdentity(args[0].Portal)
	if err != nil {
		return TunnelConnectToTargetArguments{}, errors.New("error creating portal ref")
	}

	target, err := refs.NewIdentity(args[0].Target)
	if err != nil {
		return TunnelConnectToTargetArguments{}, errors.New("error creating target ref")
	}

	origin, err := refs.NewIdentity(args[0].Origin)
	if err != nil {
		return TunnelConnectToTargetArguments{}, errors.New("error creating origin ref")
	}

	return TunnelConnectToTargetArguments{
		portal: portal,
		target: target,
		origin: origin,
	}, nil
}

func (t TunnelConnectToTargetArguments) Portal() refs.Identity {
	return t.portal
}

func (t TunnelConnectToTargetArguments) Target() refs.Identity {
	return t.target
}

func (t TunnelConnectToTargetArguments) Origin() refs.Identity {
	return t.origin
}

func (t TunnelConnectToTargetArguments) MarshalJSON() ([]byte, error) {
	return jsoniter.Marshal([]tunnelConnectToTargetArgumentsTransport{
		{
			Portal: t.portal.String(),
			Target: t.target.String(),
			Origin: t.origin.String(),
		},
	})
}

type tunnelConnectToPortalArgumentsTransport struct {
	Portal string `json:"portal"`
	Target string `json:"target"`
}

type tunnelConnectToTargetArgumentsTransport struct {
	Portal string `json:"portal"`
	Target string `json:"target"`
	Origin string `json:"origin"`
}
