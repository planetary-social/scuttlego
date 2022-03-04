package rpc

import (
	"context"
	"fmt"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type Handler interface {
	Procedure() Procedure

	// Handle should perform actions requested by the provided request and return the response using the provided
	// response writer. Request is never nil.
	Handle(req *Request, rw ResponseWriter) error
}

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

type Mux struct {
	handlers map[string]Handler
	logger   logging.Logger
}

func NewMux(logger logging.Logger) *Mux {
	return &Mux{
		handlers: make(map[string]Handler),
		logger:   logger.New("mux"),
	}
}

func (m Mux) AddHandler(handler Handler) error {
	key := handler.Procedure().Name().String()

	if _, ok := m.handlers[key]; ok {
		return fmt.Errorf("handler for method '%s' was already added", key)
	}

	m.logger.WithFields(logging.Fields{"key": key}).Debug("registering handler")
	m.handlers[key] = handler
	return nil
}

func (m Mux) Serve(ctx context.Context, conn *Connection) error {
	for {
		req, err := conn.NextRequest(ctx)
		if err != nil {
			return errors.Wrap(err, "could not receive a request")
		}

		go m.handleRequest(req, conn)
	}
}

func (m Mux) handleRequest(req *Request, conn *Connection) {
	handler, err := m.getHandler(req)
	if err != nil {
		return
		// todo return error via connection
	}

	rw := NewResponseWriter(req, conn)

	if err := handler.Handle(req, rw); err != nil {
		return
		// todo return error via connection
	}
}

func (m Mux) getHandler(req *Request) (Handler, error) {
	key := req.name.String()

	handler, ok := m.handlers[key]
	if !ok {
		m.logger.WithFields(logging.Fields{"key": key}).Debug("handler not found")
		return nil, errors.New("handler not found")
	}

	if handler.Procedure().Typ() != req.Type() {
		return nil, errors.New("unexpected procedure type")
	}

	return handler, nil
}
