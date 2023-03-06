package messages

import (
	"encoding/json"
	"time"

	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	CreateHistoryStreamProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"createHistoryStream"}),
		rpc.ProcedureTypeSource,
	)
)

func NewCreateHistoryStream(arguments CreateHistoryStreamArguments) (*rpc.Request, error) {
	j, err := arguments.MarshalJSON()
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal arguments")
	}

	return rpc.NewRequest(
		CreateHistoryStreamProcedure.Name(),
		CreateHistoryStreamProcedure.Typ(),
		j,
	)
}

const (
	defaultLive = false
	defaultOld  = true
	defaultKeys = true
)

type CreateHistoryStreamArguments struct {
	id       refs.Feed
	sequence *message.Sequence
	limit    *int
	live     bool
	old      bool
	keys     bool
}

func NewCreateHistoryStreamArguments(
	id refs.Feed,
	sequence *message.Sequence, // nil => start from beginning
	limit *int,
	live *bool,
	old *bool,
	keys *bool,
) (CreateHistoryStreamArguments, error) {
	return CreateHistoryStreamArguments{
		id:       id,
		sequence: sequence,
		limit:    limit,
		live:     valueOrDefault(live, defaultLive),
		old:      valueOrDefault(old, defaultOld),
		keys:     valueOrDefault(keys, defaultKeys),
	}, nil
}

func NewCreateHistoryStreamArgumentsFromBytes(b []byte) (CreateHistoryStreamArguments, error) {
	var args []createHistoryStreamArgumentsTransport
	if err := jsoniter.Unmarshal(b, &args); err != nil {
		return CreateHistoryStreamArguments{}, errors.Wrap(err, "json unmarshal failed")
	}

	if len(args) != 1 {
		return CreateHistoryStreamArguments{}, errors.New("expected exactly one argument")
	}

	arg := args[0]

	id, err := refs.NewFeed(arg.Id)
	if err != nil {
		return CreateHistoryStreamArguments{}, errors.Wrap(err, "invalid feed ref")
	}

	if arg.Sequence != nil && arg.Seq != nil {
		if *arg.Sequence != *arg.Seq {
			return CreateHistoryStreamArguments{}, errors.New("inconsistent sequence argument")
		}
	}

	var sequence *message.Sequence

	if arg.Sequence != nil {
		tmp, err := message.NewSequence(*arg.Sequence)
		if err != nil {
			return CreateHistoryStreamArguments{}, errors.Wrap(err, "invalid sequence")
		}
		sequence = &tmp
	}

	if arg.Seq != nil {
		tmp, err := message.NewSequence(*arg.Seq)
		if err != nil {
			return CreateHistoryStreamArguments{}, errors.Wrap(err, "invalid sequence")
		}
		sequence = &tmp
	}

	return NewCreateHistoryStreamArguments(
		id,
		sequence,
		arg.Limit,
		arg.Live,
		arg.Old,
		arg.Keys,
	)
}

func (c CreateHistoryStreamArguments) Id() refs.Feed {
	return c.id
}

func (c CreateHistoryStreamArguments) Sequence() *message.Sequence {
	return c.sequence
}

func (c CreateHistoryStreamArguments) Limit() *int {
	return c.limit
}

func (c CreateHistoryStreamArguments) Live() bool {
	return c.live
}

func (c CreateHistoryStreamArguments) Old() bool {
	return c.old
}

func (c CreateHistoryStreamArguments) Keys() bool {
	return c.keys
}

func (c CreateHistoryStreamArguments) MarshalJSON() ([]byte, error) {
	transport := []createHistoryStreamArgumentsTransport{
		{
			Id:       c.id.String(),
			Sequence: sequencePointerToIntPointer(c.sequence),
			Limit:    c.limit,
			Live:     nilIfDefault(c.live, defaultLive),
			Old:      nilIfDefault(c.old, defaultOld),
			Keys:     nilIfDefault(c.keys, defaultKeys),
		},
	}
	return jsoniter.Marshal(transport)
}

type CreateHistoryStreamResponse struct {
	key       refs.Message
	value     message.RawMessage
	timestamp time.Time
}

func NewCreateHistoryStreamResponse(
	key refs.Message,
	value message.RawMessage,
	timestamp time.Time,
) *CreateHistoryStreamResponse {
	return &CreateHistoryStreamResponse{
		key:       key,
		value:     value,
		timestamp: timestamp,
	}
}

func (c CreateHistoryStreamResponse) MarshalJSON() ([]byte, error) {
	transport := createHistoryStreamResponseTransport{
		Key:       c.key.String(),
		Value:     c.value.Bytes(),
		Timestamp: c.timestamp.UnixMilli(),
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
	Seq      *int   `json:"seq,omitempty"`
	Limit    *int   `json:"limit,omitempty"`
	Live     *bool  `json:"live,omitempty"`
	Old      *bool  `json:"old,omitempty"`
	Keys     *bool  `json:"keys,omitempty"`
}

type createHistoryStreamResponseTransport struct {
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	Timestamp int64           `json:"timestamp"`
}

func valueOrDefault(v *bool, def bool) bool {
	if v != nil {
		return *v
	}
	return def
}

func nilIfDefault(v bool, def bool) *bool {
	if v == def {
		return nil
	}
	return &v
}
