package rpc_test

import (
	"context"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := fixtures.TestContext(t)
			logger := fixtures.SomeLogger()
			raw := newRawConnectionMock()

			handler := newRequestHandlerFunc(func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
				err := rw.CloseWithError(nil)
				require.NoError(t, err)
			})

			conn, err := rpc.NewConnection(ctx, fixtures.SomeConnectionId(), raw, handler, logger)
			require.NoError(t, err)
			defer conn.Close()

			go raw.ReceiveMessages(testCase.MessagesToReceive...)

			require.Eventually(t,
				func() bool {
					return len(raw.SentMessages) == len(testCase.ExpectedSentMessages)
				},
				1*time.Second,
				10*time.Millisecond,
			)

			sort.Slice(testCase.ExpectedSentMessages, func(i, j int) bool {
				return testCase.ExpectedSentMessages[i].Header.RequestNumber() < testCase.ExpectedSentMessages[j].Header.RequestNumber()
			})

			sort.Slice(raw.SentMessages, func(i, j int) bool {
				return raw.SentMessages[i].Header.RequestNumber() < raw.SentMessages[j].Header.RequestNumber()
			})

			require.Equal(
				t,
				testCase.ExpectedSentMessages,
				raw.SentMessages,
			)
		})
	}
}

func TestPrematureTerminationByRemote(t *testing.T) {
	const requestNumber = 1

	testCases := []struct {
		Name                 string
		Handler              func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request)
		ExpectedSentMessages []*transport.Message
	}{
		{
			Name: "sending_close_is_not_automatic",
			Handler: func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
				<-ctx.Done()
				require.Error(t, ctx.Err())
			},
			ExpectedSentMessages: nil,
		},
		{
			Name: "sending_close_with_error",
			Handler: func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
				<-ctx.Done()
				require.Error(t, ctx.Err())

				err := rw.CloseWithError(ctx.Err())
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
			Handler: func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
				<-ctx.Done()
				require.Error(t, ctx.Err())

				err := rw.CloseWithError(nil)
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

			var requestHandlerTerminatedCorrectly bool

			handler := newRequestHandlerFunc(func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
				testCase.Handler(ctx, rw, req)
				requestHandlerTerminatedCorrectly = true
			})

			conn, err := rpc.NewConnection(ctx, fixtures.SomeConnectionId(), raw, handler, logger)
			require.NoError(t, err)
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
						transport.MustNewMessageHeaderFlags(fixtures.SomeBool(), true, fixtures.SomeMessageBodyType()),
						fixtures.SomeUint32(),
						requestNumber,
					),
					Body: fixtures.SomeBytes(),
				},
			)

			require.Eventually(t, func() bool { return requestHandlerTerminatedCorrectly }, 1*time.Second, 10*time.Millisecond)
			require.Equal(t, testCase.ExpectedSentMessages, raw.SentMessages)
		})
	}
}

type handlerFunc func(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request)

type requestHandlerFunc struct {
	h handlerFunc
}

func newRequestHandlerFunc(h handlerFunc) rpc.RequestHandler {
	return &requestHandlerFunc{
		h: h,
	}
}

func (r *requestHandlerFunc) HandleRequest(ctx context.Context, rw rpc.ResponseWriter, req *rpc.Request) {
	r.h(ctx, rw, req)
}

type rawConnectionMock struct {
	SentMessages []*transport.Message

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
	r.SentMessages = append(r.SentMessages, msg)
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
