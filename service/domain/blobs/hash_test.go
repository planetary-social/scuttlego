package blobs_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/stretchr/testify/require"
)

func TestHasher(t *testing.T) {
	h := blobs.NewHasher()

	b := []byte("some bytes")
	n, err := h.Write(b)
	require.Equal(t, len(b), n)
	require.NoError(t, err)

	ref, err := h.SumRef()
	require.NoError(t, err)
	require.Equal(t, "&DSLNzBDm0Enb4a9RI9UIc/38Gk9YMG5Yy2JBvpRyAU0=.sha256", ref.String())
}

func TestVerify(t *testing.T) {
	id1 := fixtures.SomeRefBlob()
	id2 := fixtures.SomeRefBlob()

	h := newHashMock(id1.Bytes())

	err := blobs.Verify(id1, h)
	require.NoError(t, err)

	err = blobs.Verify(id2, h)
	require.Error(t, err)
}

type hashMock struct {
	SumReturnValue []byte
}

func newHashMock(sumReturnValue []byte) *hashMock {
	return &hashMock{SumReturnValue: sumReturnValue}
}

func (h hashMock) Write(p []byte) (n int, err error) {
	return 0, errors.New("not implemented")
}

func (h hashMock) Sum(b []byte) []byte {
	return h.SumReturnValue
}

func (h hashMock) Reset() {
}

func (h hashMock) Size() int {
	return 0
}

func (h hashMock) BlockSize() int {
	return 0
}
