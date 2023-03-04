package transport_test

import (
	"errors"
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func BenchmarkRawConnection_Next(b *testing.B) {
	bodyBytes := fixtures.SomeBytesOfLength(100)

	flags, err := transport.NewMessageHeaderFlags(true, false, transport.MessageBodyTypeBinary)
	require.NoError(b, err)

	header, err := transport.NewMessageHeader(flags, uint32(len(bodyBytes)), 123)
	require.NoError(b, err)

	headerBytes, err := header.Bytes()
	require.NoError(b, err)

	rwc := newRwcMock(headerBytes, bodyBytes)
	logger := logging.NewDevNullLogger()

	connection := transport.NewRawConnection(rwc, logger)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := connection.Next()
		if err != nil {
			b.Fatal(err)
		}
	}
}

const (
	nextReadTypeHeader int = iota
	nextReadTypeBody
)

type rwcMock struct {
	nextReadType int
	header       []byte
	body         []byte
}

func newRwcMock(header, body []byte) *rwcMock {
	return &rwcMock{
		nextReadType: nextReadTypeHeader,
		header:       header,
		body:         body,
	}
}

func (a *rwcMock) Read(p []byte) (n int, err error) {
	if a.nextReadType == nextReadTypeHeader {
		copy(p, a.header)
		a.nextReadType = nextReadTypeBody
		return len(a.header), nil
	} else {
		copy(p, a.body)
		a.nextReadType = nextReadTypeHeader
		return len(a.body), nil
	}
}

func (a *rwcMock) Write(p []byte) (n int, err error) {
	return 0, errors.New("not implemented")
}

func (a *rwcMock) Close() error {
	return errors.New("not implemented")
}
