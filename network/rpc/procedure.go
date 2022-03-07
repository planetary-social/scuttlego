package rpc

import "github.com/boreq/errors"

type Procedure struct {
	name ProcedureName
	typ  ProcedureType
}

func (p Procedure) Name() ProcedureName {
	return p.name
}

func (p Procedure) Typ() ProcedureType {
	return p.typ
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
