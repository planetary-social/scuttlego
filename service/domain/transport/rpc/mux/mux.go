package mux

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
)

type ResponseWriter interface {
	WriteMessage(body []byte) error
}

type Handler interface {
	// Procedure returns a specification of the procedure handled by this handler. Mux routes requests bases on this
	// value.
	Procedure() rpc.Procedure

	// Handle should perform actions requested by the provided request and return responses using the provided
	// response writer. The handler returns errors to make the flow of control within the handler easier to follow.
	// If an error is returned it will be sent over the RPC connection. Request is never nil.
	Handle(ctx context.Context, rw ResponseWriter, req *rpc.Request) error
}

type Mux struct {
	handlers map[string]Handler
	logger   logging.Logger
}

func NewMux(logger logging.Logger, handlers []Handler) (*Mux, error) {
	m := &Mux{
		handlers: make(map[string]Handler),
		logger:   logger.New("mux"),
	}

	for _, handler := range handlers {
		if err := m.addHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a handler")
		}
	}

	return m, nil
}

func (m Mux) HandleRequest(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
	handler, err := m.getHandler(req)
	if err != nil {
		if err := rw.CloseWithError(errors.New("method not supported")); err != nil {
			m.logger.WithError(err).Debug("could not write an error")
		}
		return
	}

	if err := handler.Handle(ctx, rw, req); err != nil {
		if err := rw.CloseWithError(err); err != nil {
			m.logger.WithError(err).Debug("could not write an error returned by the handler")
		}
		return
	}
}

func (m Mux) addHandler(handler Handler) error {
	key := m.procedureNameToKey(handler.Procedure().Name())

	if _, ok := m.handlers[key]; ok {
		return fmt.Errorf("handler for method '%s' was already added", key)
	}

	m.logger.WithField("key", key).Debug("adding handler")
	m.handlers[key] = handler
	return nil
}

func (m Mux) getHandler(req *rpc.Request) (Handler, error) {
	key := m.procedureNameToKey(req.Name())

	handler, ok := m.handlers[key]
	if !ok {
		m.logger.WithField("key", key).Debug("handler not found")
		return nil, errors.New("handler not found")
	}

	if handler.Procedure().Typ() != req.Type() {
		return nil, errors.New("unexpected procedure type")
	}

	return handler, nil
}

func (m Mux) procedureNameToKey(name rpc.ProcedureName) string {
	return name.String()
}
