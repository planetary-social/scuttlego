package mux

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type ResponseWriter interface {
	WriteMessage(body []byte) error
}

type ResponseWriterCloser interface {
	WriteMessage(body []byte) error
	CloseWithError(err error) error
}

type Handler interface {
	// Procedure returns a specification of the procedure handled by this
	// handler. Mux routes requests bases on this value.
	Procedure() rpc.Procedure

	// Handle should perform actions requested by the provided request and
	// return responses using the provided response writer. The handler returns
	// errors to make the flow of control within the handler easier to follow.
	// If an error is returned it will be sent over the RPC connection.
	Handle(ctx context.Context, rw ResponseWriter, req *rpc.Request) error
}

type ClosingHandler interface {
	// Procedure returns a specification of the procedure handled by this
	// handler. Mux routes requests bases on this value.
	Procedure() rpc.Procedure

	// Handle should perform actions requested by the provided request and
	// return responses using the provided response writer. Handler must close
	// the response writer.
	Handle(ctx context.Context, rw ResponseWriterCloser, req *rpc.Request)
}

type Mux struct {
	handlers        map[string]Handler
	closingHandlers map[string]ClosingHandler
	logger          logging.Logger
}

func NewMux(
	logger logging.Logger,
	handlers []Handler,
	closingHandlers []ClosingHandler,
) (*Mux, error) {
	m := &Mux{
		handlers: make(map[string]Handler),
		logger:   logger.New("mux"),
	}

	for _, handler := range handlers {
		if err := m.addHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a handler")
		}
	}

	for _, handler := range closingHandlers {
		if err := m.addClosingHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a closing handler")
		}
	}

	return m, nil
}

func (m Mux) HandleRequest(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
	var findHandlerErr error

	handler, err := m.getHandler(req)
	if err == nil {
		if err := handler.Handle(ctx, rw, req); err != nil {
			if err := rw.CloseWithError(err); err != nil {
				m.logger.WithError(err).Debug("could not write an error returned by the handler")
			}
		}
		return
	} else {
		findHandlerErr = multierror.Append(findHandlerErr, err)
	}

	closingHandler, err := m.getClosingHandler(req)
	if err == nil {
		closingHandler.Handle(ctx, rw, req)
		return
	} else {
		findHandlerErr = multierror.Append(findHandlerErr, err)
	}

	m.logger.WithError(findHandlerErr).Debug("handler not found")

	if err := rw.CloseWithError(findHandlerErr); err != nil {
		m.logger.WithError(err).Debug("could not write an error")
	}
}

func (m Mux) addHandler(handler Handler) error {
	key := m.procedureNameToKey(handler.Procedure().Name())

	if err := m.checkKeyUnique(key); err != nil {
		return errors.Wrap(err, "handler is not unique")
	}

	m.logger.WithField("key", key).Debug("adding handler")
	m.handlers[key] = handler
	return nil
}

func (m Mux) addClosingHandler(handler ClosingHandler) error {
	key := m.procedureNameToKey(handler.Procedure().Name())

	if err := m.checkKeyUnique(key); err != nil {
		return errors.Wrap(err, "handler is not unique")
	}

	m.logger.WithField("key", key).Debug("adding closing handler")
	m.closingHandlers[key] = handler
	return nil
}

func (m Mux) checkKeyUnique(key string) error {
	if _, ok := m.handlers[key]; ok {
		return fmt.Errorf("handler for method '%s' was already added", key)
	}

	if _, ok := m.closingHandlers[key]; ok {
		return fmt.Errorf("closing handler for method '%s' was already added", key)
	}

	return nil
}

func (m Mux) getHandler(req *rpc.Request) (Handler, error) {
	key := m.procedureNameToKey(req.Name())

	handler, ok := m.handlers[key]
	if !ok {
		return nil, errors.New("handler not found")
	}

	if handler.Procedure().Typ() != req.Type() {
		return nil, errors.New("unexpected procedure type")
	}

	return handler, nil
}

func (m Mux) getClosingHandler(req *rpc.Request) (ClosingHandler, error) {
	key := m.procedureNameToKey(req.Name())

	handler, ok := m.closingHandlers[key]
	if !ok {
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
