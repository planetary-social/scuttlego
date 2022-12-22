package notx_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/stretchr/testify/require"
)

func Test(t *testing.T) {
	ts := di.BuildBadgerNoTxTestAdapters(t)

	_, err := ts.ReadBlobWantListRepository.List()
	require.NoError(t, err)
}
