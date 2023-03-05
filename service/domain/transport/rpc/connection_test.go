package rpc_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestIncomingRequests(t *testing.T) {
	flagsNoTermination := transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON)
	flagsTermination := transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON)

	testCases := []struct {
		Name                 string
		MessagesToReceive    []*transport.Message
		ExpectedSentMessages []*transport.Message
	}{
		{
			Name: "wrong_order",
			MessagesToReceive: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						2,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							fixtures.SomeProcedureType(),
							[]byte("[]"),
						),
					),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							fixtures.SomeProcedureType(),
							[]byte("[]"),
						),
					),
				},
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-1,
					),
					Body: []byte("true"),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-2,
					),
					Body: []byte("true"),
				},
			},
		},
		{
			Name: "source",
			MessagesToReceive: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeSource,
							[]byte("[]"),
						),
					),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						2,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeSource,
							[]byte("[]"),
						),
					),
				},
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-1,
					),
					Body: []byte("true"),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-2,
					),
					Body: []byte("true"),
				},
			},
		},
		{
			Name: "async",
			MessagesToReceive: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeAsync,
							[]byte("[]"),
						),
					),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						2,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeAsync,
							[]byte("[]"),
						),
					),
				},
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-1,
					),
					Body: []byte("true"),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-2,
					),
					Body: []byte("true"),
				},
			},
		},
		{
			Name: "rooms_contains_no_request_type",
			MessagesToReceive: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeAsync,
							[]byte("[]"),
						),
					),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						2,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeAsync,
							[]byte("[]"),
						),
					),
				},
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-1,
					),
					Body: []byte("true"),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-2,
					),
					Body: []byte("true"),
				},
			},
		},
		{
			Name: "duplex",
			MessagesToReceive: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeDuplex,
							[]byte("[]"),
						),
					),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: []byte("12345"),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						2,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeDuplex,
							[]byte("[]"),
						),
					),
				},
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-1,
					),
					Body: []byte("true"),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-2,
					),
					Body: []byte("true"),
				},
			},
		},
		{
			Name: "manifest_has_invalid_name",
			MessagesToReceive: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						1,
					),
					Body: []byte(`{ "name": "manifest", "args": [], "type": "async" }`),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsNoTermination,
						fixtures.SomeUint32(),
						2,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeAsync,
							[]byte("[]"),
						),
					),
				},
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						38,
						-1,
					),
					Body: []byte(`{"error":"received malformed request"}`),
				},
				{
					Header: transport.MustNewMessageHeader(
						flagsTermination,
						4,
						-2,
					),
					Body: []byte("true"),
				},
			},
		},
	}

	for i := range testCases {
		testCase := testCases[i]
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(fixtures.TestContext(t))
			logger := fixtures.SomeLogger()
			raw := newRawConnectionMock()

			handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
				err := s.CloseWithError(nil)
				require.NoError(t, err)
			})

			conn, err := rpc.NewConnection(fixtures.SomeConnectionId(), fixtures.SomeBool(), raw, handler, logger)
			require.NoError(t, err)

			connectionLoopClosed := make(chan struct{})
			t.Cleanup(func() {
				cancel()
				<-connectionLoopClosed
			})

			go func() {
				defer close(connectionLoopClosed)
				if err := conn.Loop(ctx); err != nil {
					logger.Debug().WithError(err).Message("conn loop exited")
				}
			}()
			defer conn.Close()

			go raw.ReceiveMessages(testCase.MessagesToReceive...)

			require.Eventually(t,
				func() bool {
					sentMessages := raw.SentMessages()
					t.Log("number of sent messages:", len(sentMessages))
					for i, msg := range sentMessages {
						t.Log(i, fmt.Sprintf("%#v", msg))
					}
					return len(sentMessages) == len(testCase.ExpectedSentMessages)
				},
				5*time.Second,
				100*time.Millisecond,
			)

			sentMessages := raw.SentMessages()

			sort.Slice(testCase.ExpectedSentMessages, func(i, j int) bool {
				return testCase.ExpectedSentMessages[i].Header.RequestNumber() < testCase.ExpectedSentMessages[j].Header.RequestNumber()
			})

			sort.Slice(sentMessages, func(i, j int) bool {
				return sentMessages[i].Header.RequestNumber() < sentMessages[j].Header.RequestNumber()
			})

			require.Equal(
				t,
				testCase.ExpectedSentMessages,
				sentMessages,
			)
		})
	}
}

func TestPrematureTerminationByRemote(t *testing.T) {
	const requestNumber = 1

	testCases := []struct {
		Name                 string
		Handler              func(ctx context.Context, s rpc.Stream, req *rpc.Request)
		ExpectedSentMessages []*transport.Message
	}{
		{
			Name: "sending_close_is_not_automatic",
			Handler: func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
				<-ctx.Done()
				require.Error(t, ctx.Err())
			},
			ExpectedSentMessages: []*transport.Message{},
		},
		{
			Name: "sending_close_with_error",
			Handler: func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
				<-ctx.Done()
				require.Error(t, ctx.Err())

				err := s.CloseWithError(ctx.Err())
				require.NoError(t, err)
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON),
						28,
						-requestNumber,
					),
					Body: []byte(`{"error":"context canceled"}`),
				},
			},
		},
		{
			Name: "sending_close_without_error",
			Handler: func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
				<-ctx.Done()
				require.Error(t, ctx.Err())

				err := s.CloseWithError(nil)
				require.NoError(t, err)
			},
			ExpectedSentMessages: []*transport.Message{
				{
					Header: transport.MustNewMessageHeader(
						transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON),
						4,
						-requestNumber,
					),
					Body: []byte("true"),
				},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			logger := fixtures.SomeLogger()
			raw := newRawConnectionMock()

			var requestHandlerTerminatedCorrectly atomic.Bool

			handler := newRequestHandlerFunc(func(ctx context.Context, s rpc.Stream, req *rpc.Request) {
				testCase.Handler(ctx, s, req)
				requestHandlerTerminatedCorrectly.Store(true)
			})

			conn, err := rpc.NewConnection(fixtures.SomeConnectionId(), fixtures.SomeBool(), raw, handler, logger)
			require.NoError(t, err)

			go func() {
				if err := conn.Loop(ctx); err != nil {
					logger.Debug().WithError(err).Message("conn loop exited")
				}
			}()
			defer conn.Close()

			go raw.ReceiveMessages(
				&transport.Message{
					Header: transport.MustNewMessageHeader(
						transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), false, transport.MessageBodyTypeJSON),
						fixtures.SomeUint32(),
						requestNumber,
					),
					Body: rpc.MustMarshalRequestBody(
						rpc.MustNewRequest(
							fixtures.SomeProcedureName(),
							rpc.ProcedureTypeSource,
							[]byte("[]"),
						),
					),
				},
				&transport.Message{
					Header: transport.MustNewMessageHeader(
						transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), true, transport.MessageBodyTypeJSON),
						fixtures.SomeUint32(),
						requestNumber,
					),
					Body: fixtures.SomeBytes(),
				},
			)

			require.Eventually(t, func() bool { return requestHandlerTerminatedCorrectly.Load() }, 1*time.Second, 10*time.Millisecond)
			require.Equal(t, testCase.ExpectedSentMessages, raw.SentMessages())
		})
	}
}

type handlerFunc func(ctx context.Context, s rpc.Stream, req *rpc.Request)

type requestHandlerFunc struct {
	h handlerFunc
}

func newRequestHandlerFunc(h handlerFunc) rpc.RequestHandler {
	return &requestHandlerFunc{
		h: h,
	}
}

func (r *requestHandlerFunc) HandleRequest(ctx context.Context, s rpc.Stream, req *rpc.Request) {
	r.h(ctx, s, req)
}

type rawConnectionMock struct {
	sentMessages []*transport.Message

	closedCh chan struct{}
	closed   bool
	mutex    sync.Mutex

	incoming chan *transport.Message
}

func newRawConnectionMock() *rawConnectionMock {
	return &rawConnectionMock{
		closedCh: make(chan struct{}),
		incoming: make(chan *transport.Message),
	}
}

func (r *rawConnectionMock) ReceiveMessages(messages ...*transport.Message) {
	for _, msg := range messages {
		select {
		case r.incoming <- msg:
		case <-r.closedCh:
			return
		}
	}
}

func (r *rawConnectionMock) Next() (*transport.Message, error) {
	select {
	case msg := <-r.incoming:
		return msg, nil
	case <-r.closedCh:
		return nil, errors.New("raw connection is closed")
	}
}

func (r *rawConnectionMock) Send(msg *transport.Message) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if r.closed {
		return errors.New("raw connection is closed")
	}
	r.sentMessages = append(r.sentMessages, msg)
	return nil
}

func (r *rawConnectionMock) Close() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	if !r.closed {
		r.closed = true
		close(r.closedCh)
	}
	return nil
}

func (r *rawConnectionMock) SentMessages() []*transport.Message {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	result := make([]*transport.Message, len(r.sentMessages))
	copy(result, r.sentMessages)
	return result
}
