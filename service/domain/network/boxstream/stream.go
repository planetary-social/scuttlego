package boxstream

import (
	"bytes"
	"io"

	"github.com/planetary-social/go-ssb/service/domain/identity"

	"github.com/hashicorp/go-multierror"

	"github.com/boreq/errors"
	"go.cryptoscope.co/secretstream/boxstream"
)

const maxBoxStreamBodyLength = 4096

type Key [32]byte
type Nonce [24]byte

type HandshakeResult struct {
	Remote identity.Public

	WriteSecret Key
	WriteNonce  Nonce

	ReadSecret Key
	ReadNonce  Nonce
}

type Stream struct {
	remote identity.Public
	closer io.Closer
	boxer  *boxstream.Boxer

	readBytesCh chan []byte
	readErrCh   chan error
	readBuf     bytes.Buffer
}

func NewStream(rw io.ReadWriteCloser, handshakeResult HandshakeResult) (*Stream, error) {
	// todo don't modify handshake result

	boxer := boxstream.NewBoxer(rw, (*[24]byte)(&handshakeResult.WriteNonce), (*[32]byte)(&handshakeResult.WriteSecret))
	unboxer := boxstream.NewUnboxer(rw, (*[24]byte)(&handshakeResult.ReadNonce), (*[32]byte)(&handshakeResult.ReadSecret))

	s := &Stream{
		remote: handshakeResult.Remote,
		closer: rw,
		boxer:  boxer,

		readBytesCh: make(chan []byte),
		readErrCh:   make(chan error),
	}

	go s.readLoop(unboxer)

	return s, nil
}

func (s Stream) Remote() identity.Public {
	return s.remote
}

// Write will always return 0 as the number of bytes written due to limitations of the underlying implementation.
func (s Stream) Write(p []byte) (int, error) {
	for i := 0; i < len(p); i += maxBoxStreamBodyLength {
		chunkEnd := min(len(p), i+maxBoxStreamBodyLength)
		if err := s.writeChunk(p[i:chunkEnd]); err != nil {
			return 0, errors.Wrap(err, "failed to write a chunk")
		}
	}
	return 0, nil
}

func (s *Stream) Read(p []byte) (n int, err error) {
	if s.readBuf.Len() > 0 {
		return s.readBuf.Read(p)
	}

	// channels are used to block if there is no data
	select {
	case b, ok := <-s.readBytesCh:
		if !ok {
			return 0, io.EOF
		}
		s.readBuf.Write(b)
	case err, ok := <-s.readErrCh:
		if !ok {
			return 0, io.EOF
		}
		return 0, err
	}

	return s.readBuf.Read(p)
}

func (s Stream) Close() error {
	var err error
	err = multierror.Append(err, s.boxer.WriteGoodbye())
	err = multierror.Append(err, s.closer.Close())
	return err
}

// writeChunk accepts at most 4096 bytes.
func (s *Stream) writeChunk(p []byte) (err error) {
	if len(p) > maxBoxStreamBodyLength {
		return errors.New("chunk too long")
	}

	return s.boxer.WriteMessage(p)
}

// readLoop is here to fix the fact that boxstream.Unboxer does not implement io.Reader.
func (s *Stream) readLoop(unboxer *boxstream.Unboxer) {
	defer close(s.readBytesCh)
	defer close(s.readErrCh)

	for {
		message, err := unboxer.ReadMessage()
		if err != nil {
			s.readErrCh <- err
			return
		}

		s.readBytesCh <- message
	}
}

func min(a, b int) int {
	if a < b {
		return a
	} else {
		return b
	}
}
