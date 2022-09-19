package mux

import (
	"context"
	"fmt"

	"github.com/boreq/errors"
	"github.com/hashicorp/go-multierror"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
)

type Stream interface {
	IncomingMessages() (<-chan rpc.IncomingMessage, error)
	WriteMessage(body []byte) error
}

type CloserStream interface {
	IncomingMessages() (<-chan rpc.IncomingMessage, error)
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
	// If an error is returned it will be sent over the RPC connection,
	// otherwise the connection is terminated cleanly. Handle is executed in a
	// separate goroutine and can therefore block and process the request for as
	// long as it needs.
	Handle(ctx context.Context, s Stream, req *rpc.Request) error
}

type SynchronousHandler interface {
	// Procedure returns a specification of the procedure handled by this
	// handler. Mux routes requests bases on this value.
	Procedure() rpc.Procedure

	// Handle should perform actions requested by the provided request and
	// return responses using the provided response writer. Handler must close
	// the response writer. Handle isn't processed in a separate goroutine and
	// should avoid blocking. This is useful when there is a need to offload the
	// work to a set of workers without spawning extra goroutines to limit
	// pressure on the scheduler and garbage collector. This is useful in the
	// case of create history stream requests which are performed in very large
	// amounts when connecting to large pubs.
	Handle(ctx context.Context, s CloserStream, req *rpc.Request)
}

type Mux struct {
	handlers            map[string]Handler
	synchronousHandlers map[string]SynchronousHandler
	logger              logging.Logger
}

func NewMux(
	logger logging.Logger,
	handlers []Handler,
	synchronousHandlers []SynchronousHandler,
) (*Mux, error) {
	m := &Mux{
		handlers:            make(map[string]Handler),
		synchronousHandlers: make(map[string]SynchronousHandler),
		logger:              logger.New("mux"),
	}

	for _, handler := range handlers {
		if err := m.addHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a handler")
		}
	}

	for _, handler := range synchronousHandlers {
		if err := m.addSynchronousHandler(handler); err != nil {
			return nil, errors.Wrap(err, "could not add a synchronous handler")
		}
	}

	return m, nil
}

func (m Mux) HandleRequest(ctx context.Context, s CloserStream, req *rpc.Request) {
	var findHandlerErr error

	handler, err := m.getHandler(req)
	if err == nil {
		go func() {
			if err := handler.Handle(ctx, s, req); err != nil {
				m.logger.WithError(err).Debug("handler returned an error")
				if closeErr := s.CloseWithError(err); closeErr != nil {
					m.logger.WithError(closeErr).Debug("could not write an error returned by the handler")
				}
			}
		}()
		return
	} else {
		findHandlerErr = multierror.Append(findHandlerErr, err)
	}

	closingHandler, err := m.getSynchronousHandler(req)
	if err == nil {
		closingHandler.Handle(ctx, s, req)
		return
	} else {
		findHandlerErr = multierror.Append(findHandlerErr, err)
	}

	m.logger.WithError(findHandlerErr).Debug("handler not found")

	if err := s.CloseWithError(findHandlerErr); err != nil {
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

func (m Mux) addSynchronousHandler(handler SynchronousHandler) error {
	key := m.procedureNameToKey(handler.Procedure().Name())

	if err := m.checkKeyUnique(key); err != nil {
		return errors.Wrap(err, "handler is not unique")
	}

	m.logger.WithField("key", key).Debug("adding synchronous handler")
	m.synchronousHandlers[key] = handler
	return nil
}

func (m Mux) checkKeyUnique(key string) error {
	if _, ok := m.handlers[key]; ok {
		return fmt.Errorf("handler for method '%s' was already added", key)
	}

	if _, ok := m.synchronousHandlers[key]; ok {
		return fmt.Errorf("synchronous handler for method '%s' was already added", key)
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

func (m Mux) getSynchronousHandler(req *rpc.Request) (SynchronousHandler, error) {
	key := m.procedureNameToKey(req.Name())

	handler, ok := m.synchronousHandlers[key]
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
