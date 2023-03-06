package messages

import (
	"encoding/json"
	"strconv"

	"github.com/boreq/errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

var (
	EbtReplicateProcedure = rpc.MustNewProcedure(
		rpc.MustNewProcedureName([]string{"ebt", "replicate"}),
		rpc.ProcedureTypeDuplex,
	)
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

type EbtReplicateNotes struct {
	notes []EbtReplicateNote
}

func NewEbtReplicateNotesFromBytes(b []byte) (EbtReplicateNotes, error) {
	var frontier map[string]ebtReplicateNoteTransport
	if err := jsoniter.Unmarshal(b, &frontier); err != nil {
		return EbtReplicateNotes{}, errors.Wrap(err, "json unmarshal error")
	}

	var notes []EbtReplicateNote

	for feedRefString, note := range frontier {
		ref, err := refs.NewFeed(feedRefString)
		if err != nil {
			return EbtReplicateNotes{}, errors.Wrap(err, "error creating a ref")
		}

		note, err := NewEbtReplicateNote(ref, note.Receive, note.Replicate, note.Seq)
		if err != nil {
			return EbtReplicateNotes{}, errors.Wrap(err, "error creating a note")
		}

		notes = append(notes, note)
	}

	return EbtReplicateNotes{
		notes: notes,
	}, nil
}

func NewEbtReplicateNotes(notes []EbtReplicateNote) (EbtReplicateNotes, error) {
	v := map[string]struct{}{}
	for _, note := range notes {
		if note.IsZero() {
			return EbtReplicateNotes{}, errors.New("note is zero")
		}

		if _, ok := v[note.Ref().String()]; ok {
			return EbtReplicateNotes{}, errors.New("duplicate note")
		}

		v[note.Ref().String()] = struct{}{}
	}

	return EbtReplicateNotes{
		notes: notes,
	}, nil
}

func MustNewEbtReplicateNotes(notes []EbtReplicateNote) EbtReplicateNotes {
	v, err := NewEbtReplicateNotes(notes)
	if err != nil {
		panic(err)
	}
	return v
}

func (n *EbtReplicateNotes) Notes() []EbtReplicateNote {
	result := make([]EbtReplicateNote, len(n.notes))
	copy(result, n.notes)
	return result
}

func (n *EbtReplicateNotes) Empty() bool {
	return len(n.notes) == 0
}

func (n *EbtReplicateNotes) MarshalJSON() ([]byte, error) {
	frontier := make(map[string]ebtReplicateNoteTransport, len(n.notes))
	for _, note := range n.notes {
		frontier[note.Ref().String()] = ebtReplicateNoteTransport{
			Seq:       note.Sequence(),
			Replicate: note.Replicate(),
			Receive:   note.Receive(),
		}
	}
	return jsoniter.Marshal(frontier)
}

type EbtReplicateNote struct {
	ref       refs.Feed
	receive   bool
	replicate bool
	sequence  int
}

func NewEbtReplicateNote(ref refs.Feed, receive, replicate bool, sequence int) (EbtReplicateNote, error) {
	if ref.IsZero() {
		return EbtReplicateNote{}, errors.New("zero value of feed ref")
	}

	return EbtReplicateNote{
		ref:       ref,
		receive:   receive,
		replicate: replicate,
		sequence:  sequence,
	}, nil
}

func MustNewEbtReplicateNote(ref refs.Feed, receive, replicate bool, sequence int) EbtReplicateNote {
	v, err := NewEbtReplicateNote(ref, receive, replicate, sequence)
	if err != nil {
		panic(err)
	}
	return v
}

func (e EbtReplicateNote) Ref() refs.Feed {
	return e.ref
}

func (e EbtReplicateNote) Receive() bool {
	return e.receive
}

func (e EbtReplicateNote) Replicate() bool {
	return e.replicate
}

func (e EbtReplicateNote) Sequence() int {
	return e.sequence
}

func (e EbtReplicateNote) IsZero() bool {
	return e.ref.IsZero()
}

func (e EbtReplicateNote) Equal(o EbtReplicateNote) bool {
	return e.ref.Equal(o.ref) &&
		e.receive == o.receive &&
		e.replicate == o.replicate &&
		e.sequence == o.sequence
}

func (e EbtReplicateNote) MarshalJSON() ([]byte, error) {
	v := &ebtReplicateNoteTransport{
		Seq:       e.sequence,
		Replicate: e.replicate,
		Receive:   e.receive,
	}
	return v.MarshalJSON()
}

type ebtReplicateNoteTransport struct {
	Seq       int
	Replicate bool
	Receive   bool
}

func (s *ebtReplicateNoteTransport) UnmarshalJSON(b []byte) error {
	i, err := strconv.ParseInt(string(b), 10, 32)
	if err != nil {
		return errors.Wrap(err, "failed to unmarshal the number")
	}

	s.Replicate = i != -1
	s.Receive = !(i&1 == 1)
	s.Seq = int(i >> 1)
	return nil
}

func (s *ebtReplicateNoteTransport) MarshalJSON() ([]byte, error) {
	var i int
	if !s.Replicate {
		return []byte("-1"), nil
	}
	i = s.Seq
	i = i << 1 // make room for the rx bit
	if s.Receive {
		i |= 0
	} else {
		i |= 1
	}
	return []byte(strconv.FormatInt(int64(i), 10)), nil
}
