package rpc

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type Handler interface {
	// Procedure returns a specification of the procedure handled by this handler. Mux routes requests bases on this
	// value.
	Procedure() Procedure

	// Handle should perform actions requested by the provided request and return the response using the provided
	// response writer. Request is never nil.
	Handle(req *Request, rw *ResponseWriter) error
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
	key := m.procedureNameToKey(handler.Procedure().Name())

	if _, ok := m.handlers[key]; ok {
		return fmt.Errorf("handler for method '%s' was already added", key)
	}

	m.logger.WithField("key", key).Debug("registering handler")
	m.handlers[key] = handler
	return nil
}

func (m Mux) HandleRequest(req *Request, rw *ResponseWriter) {
	handler, err := m.getHandler(req)
	if err != nil {
		if err := rw.CloseWithError(errors.New("method not supported")); err != nil {
			m.logger.WithError(err).Debug("could not write an error")
		}
		return
	}

	if err := handler.Handle(req, rw); err != nil {
		if err := rw.CloseWithError(err); err != nil {
			m.logger.WithError(err).Debug("could not write an error returned by the handler")
		}
		return
	}
}

func (m Mux) getHandler(req *Request) (Handler, error) {
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

func (m Mux) procedureNameToKey(name ProcedureName) string {
	return name.String()
}
