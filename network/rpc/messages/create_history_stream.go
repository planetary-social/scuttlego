package messages

import (
	"encoding/json"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/network/rpc"
	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
)

var (
	CreateHistoryStreamProcedureName = rpc.MustNewProcedureName([]string{"createHistoryStream"})
	CreateHistoryStreamProcedure     = rpc.MustNewProcedure(CreateHistoryStreamProcedureName, rpc.ProcedureTypeSource)
)

func NewCreateHistoryStream(arguments CreateHistoryStreamArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		CreateHistoryStreamProcedure.Name(),
		CreateHistoryStreamProcedure.Typ(),
		true,
		j,
	)
}

type CreateHistoryStreamArguments struct {
	id       refs.Feed
	sequence *message.Sequence
	limit    *int
	live     *bool
	old      *bool
	keys     *bool
}

func NewCreateHistoryStreamArguments(
	id refs.Feed,
	sequence *message.Sequence, // nil => start from beginning
	limit *int, // nil => unlimited
	live *bool, // nil => false
	old *bool, // nil => true
	keys *bool, // nil => true
) (CreateHistoryStreamArguments, error) {
	// todo checks as some arguments can't be used together

	return CreateHistoryStreamArguments{
		id:       id,
		sequence: sequence,
		limit:    limit,
		live:     live,
		old:      old,
		keys:     keys,
	}, nil
}

func (c CreateHistoryStreamArguments) MarshalJSON() ([]byte, error) {
	transport := []createHistoryStreamArgumentsTransport{
		{
			Id:       c.id.String(),
			Sequence: sequencePointerToIntPointer(c.sequence),
			Limit:    c.limit,
			Live:     c.live,
			Old:      c.old,
			Keys:     c.keys,
		},
	}
	return json.Marshal(transport)
}

func sequencePointerToIntPointer(sequence *message.Sequence) *int {
	if sequence == nil {
		return nil
	}
	seq := sequence.Int()
	return &seq
}

type createHistoryStreamArgumentsTransport struct {
	Id       string `json:"id"`
	Sequence *int   `json:"sequence,omitempty"`
	Limit    *int   `json:"limit,omitempty"`
	Live     *bool  `json:"live,omitempty"`
	Old      *bool  `json:"old,omitempty"`
	Keys     *bool  `json:"keys,omitempty"`
}
