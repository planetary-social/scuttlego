package rpc

import (
	"strings"

	"github.com/boreq/errors"
)

type Request struct {
	name      ProcedureName
	typ       ProcedureType
	stream    bool
	arguments []byte
}

func NewRequest(name ProcedureName, typ ProcedureType, stream bool, arguments []byte) (Request, error) {
	if name.IsZero() {
		return Request{}, errors.New("zero value of request name")
	}

	if typ.IsZero() {
		return Request{}, errors.New("zero value of request type")
	}

	return Request{
		name:      name,
		typ:       typ,
		stream:    stream,
		arguments: arguments,
	}, nil
}

func (r Request) Name() ProcedureName {
	return r.name
}

func (r Request) Type() ProcedureType {
	return r.typ
}

func (r Request) Stream() bool {
	return r.stream
}

func (r Request) Arguments() []byte {
	tmp := make([]byte, len(r.arguments))
	copy(tmp, r.arguments)
	return tmp
}

type Response struct {
	b []byte
}

func NewResponse(b []byte) *Response {
	return &Response{b: b}
}

func (r Response) Bytes() []byte {
	return r.b
}

type ProcedureName struct {
	name []string
}

func NewProcedureName(name []string) (ProcedureName, error) {
	if len(name) == 0 {
		return ProcedureName{}, errors.New("name must have at least one component")
	}

	for _, component := range name {
		if len(component) == 0 {
			return ProcedureName{}, errors.New("name components can't be empty strings")
		}
	}

	return ProcedureName{
		name: name,
	}, nil
}

func MustNewProcedureName(name []string) ProcedureName {
	v, err := NewProcedureName(name)
	if err != nil {
		panic(err)
	}
	return v
}

func (n ProcedureName) Components() []string {
	tmp := make([]string, len(n.name))
	copy(tmp, n.name)
	return tmp
}

func (n ProcedureName) IsZero() bool {
	return len(n.name) == 0
}

func (n ProcedureName) String() string {
	return strings.Join(n.name, ".")
}

type ProcedureType struct {
	s string
}

func (t ProcedureType) IsZero() bool {
	return t == ProcedureType{}
}

var (
	ProcedureTypeUnknown = ProcedureType{"unknown"} // some procedures don't have a type
	ProcedureTypeSource  = ProcedureType{"source"}
	ProcedureTypeDuplex  = ProcedureType{"duplex"}
	ProcedureTypeAsync   = ProcedureType{"async"}
)
