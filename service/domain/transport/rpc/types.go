package rpc

import (
	"encoding/json"
	"strings"

	"github.com/boreq/errors"
)

type Request struct {
	name      ProcedureName
	typ       ProcedureType
	arguments json.RawMessage
}

func NewRequest(name ProcedureName, typ ProcedureType, arguments json.RawMessage) (*Request, error) {
	if name.IsZero() {
		return nil, errors.New("zero value of request name")
	}

	if typ.IsZero() {
		return nil, errors.New("zero value of request type")
	}

	return &Request{
		name:      name,
		typ:       typ,
		arguments: arguments,
	}, nil
}

func MustNewRequest(name ProcedureName, typ ProcedureType, arguments json.RawMessage) *Request {
	v, err := NewRequest(name, typ, arguments)
	if err != nil {
		panic(err)
	}

	return v
}

func (r Request) Name() ProcedureName {
	return r.name
}

func (r Request) Type() ProcedureType {
	return r.typ
}

func (r Request) Arguments() json.RawMessage {
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

func (n ProcedureName) Equal(o ProcedureName) bool {
	if len(o.name) != len(n.name) {
		return false
	}
	for i := range n.name {
		if n.name[i] != o.name[i] {
			return false
		}
	}
	return true
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

type Procedure struct {
	name ProcedureName
	typ  ProcedureType
}

func NewProcedure(name ProcedureName, typ ProcedureType) (Procedure, error) {
	if name.IsZero() {
		return Procedure{}, errors.New("zero value of name")
	}

	if typ.IsZero() {
		return Procedure{}, errors.New("zero value of type")
	}

	return Procedure{
		name: name,
		typ:  typ,
	}, nil
}

func MustNewProcedure(name ProcedureName, typ ProcedureType) Procedure {
	v, err := NewProcedure(name, typ)
	if err != nil {
		panic(err)
	}
	return v
}

func (p Procedure) Name() ProcedureName {
	return p.name
}

func (p Procedure) Typ() ProcedureType {
	return p.typ
}
