package replication_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/stretchr/testify/require"
)

func TestFeedState_ZeroValue(t *testing.T) {
	_, err := replication.NewFeedState(message.Sequence{})
	require.EqualError(t, err, "zero value of sequence", "zero value of sequence is not accepted, the other constructor should be used instead")
}
