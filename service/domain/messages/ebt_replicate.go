package messages

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/replication/ebt"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	EbtReplicateProcedureName = rpc.MustNewProcedureName([]string{"ebt", "replicate"})
	EbtReplicateProcedure     = rpc.MustNewProcedure(EbtReplicateProcedureName, rpc.ProcedureTypeDuplex)
)

func NewEbtReplicate(arguments EbtReplicateArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		EbtReplicateProcedure.Name(),
		EbtReplicateProcedure.Typ(),
		j,
	)
}

type EbtReplicateArguments struct {
	version int
	format  EbtReplicateFormat
}

func NewEbtReplicateArguments(
	version int,
	format EbtReplicateFormat,
) (EbtReplicateArguments, error) {
	if format.IsZero() {
		return EbtReplicateArguments{}, errors.New("zero value of format")
	}

	return EbtReplicateArguments{
		version: version,
		format:  format,
	}, nil
}

func NewEbtReplicateArgumentsFromBytes(b []byte) (EbtReplicateArguments, error) {
	var args []ebtReplicateArgumentsTransport

	if err := json.Unmarshal(b, &args); err != nil {
		return EbtReplicateArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return EbtReplicateArguments{}, errors.New("expected exactly one argument")
	}

	arg := args[0]

	f, err := unmarshalEbtReplicateFormat(arg.Format)
	if err != nil {
		return EbtReplicateArguments{}, errors.Wrap(err, "could not unmarshal the format")
	}

	return NewEbtReplicateArguments(
		arg.Version,
		f,
	)
}
func (c EbtReplicateArguments) Version() int {
	return c.version
}

func (c EbtReplicateArguments) Format() EbtReplicateFormat {
	return c.format
}

func (c EbtReplicateArguments) MarshalJSON() ([]byte, error) {
	f, err := marshalEbtReplicateFormat(c.format)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal ebt format")
	}

	transport := []ebtReplicateArgumentsTransport{
		{
			Version: c.version,
			Format:  f,
		},
	}
	return json.Marshal(transport)
}

const ebtReplicateFormatTransportClassic = "classic"

func marshalEbtReplicateFormat(f EbtReplicateFormat) (string, error) {
	switch f {
	case EbtReplicateFormatClassic:
		return ebtReplicateFormatTransportClassic, nil
	default:
		return "", errors.New("unknown format")
	}
}

func unmarshalEbtReplicateFormat(f string) (EbtReplicateFormat, error) {
	switch f {
	case ebtReplicateFormatTransportClassic:
		return EbtReplicateFormatClassic, nil
	default:
		return EbtReplicateFormat{}, errors.New("unknown format")
	}
}

type ebtReplicateArgumentsTransport struct {
	Version int    `json:"version"`
	Format  string `json:"format"`
}

var (
	EbtReplicateFormatClassic = EbtReplicateFormat{"classic"}
)

type EbtReplicateFormat struct {
	s string
}

func (f EbtReplicateFormat) IsZero() bool {
	return f == EbtReplicateFormat{}
}

type NoteOrMessage struct {
	note    ebt.Note
	message message.Message
}
