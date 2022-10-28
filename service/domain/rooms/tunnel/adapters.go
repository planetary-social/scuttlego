package tunnel

import (
	"bytes"
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/mux"
)

type ResponseStreamReadWriteCloserAdapter struct {
	cancel context.CancelFunc
	stream rpc.ResponseStream
	buf    *bytes.Buffer
}

func NewResponseStreamReadWriteCloserAdapter(stream rpc.ResponseStream, cancel context.CancelFunc) *ResponseStreamReadWriteCloserAdapter {
	return &ResponseStreamReadWriteCloserAdapter{
		stream: stream,
		cancel: cancel,
		buf:    &bytes.Buffer{},
	}
}

func (s ResponseStreamReadWriteCloserAdapter) Read(p []byte) (int, error) {
	if s.buf.Len() == 0 {
		resp, ok := <-s.stream.Channel()
		if !ok {
			return 0, errors.New("channel closed")
		}

		if err := resp.Err; err != nil {
			return 0, errors.Wrap(err, "stream returned an error")
		}

		s.buf.Write(resp.Value.Bytes())
	}

	return s.buf.Read(p)
}

func (s ResponseStreamReadWriteCloserAdapter) Write(p []byte) (n int, err error) {
	return len(p), s.stream.WriteMessage(p)
}

func (s ResponseStreamReadWriteCloserAdapter) Close() error {
	s.cancel()
	return nil
}

type StreamReadWriteCloserAdapter struct {
	stream mux.Stream
	cancel context.CancelFunc
	buf    *bytes.Buffer
}

func NewStreamReadWriteCloserAdapter(stream mux.Stream, cancel context.CancelFunc) *StreamReadWriteCloserAdapter {
	return &StreamReadWriteCloserAdapter{
		stream: stream,
		cancel: cancel,
		buf:    &bytes.Buffer{},
	}
}

func (s StreamReadWriteCloserAdapter) Read(p []byte) (n int, err error) {
	if s.buf.Len() == 0 {
		ch, err := s.stream.IncomingMessages()
		if err != nil {
			return 0, errors.Wrap(err, "failed to get incoming message channel")
		}

		resp, ok := <-ch
		if !ok {
			return 0, errors.New("channel closed")
		}

		s.buf.Write(resp.Body)
	}

	return s.buf.Read(p)
}

func (s StreamReadWriteCloserAdapter) Write(p []byte) (n int, err error) {
	return len(p), s.stream.WriteMessage(p)
}

func (s StreamReadWriteCloserAdapter) Close() error {
	s.cancel()
	return nil
}
