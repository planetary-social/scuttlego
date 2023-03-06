package replication_test

import (
	"context"
	"testing"

	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/internal/mocks"
	"github.com/planetary-social/scuttlego/service/domain/replication"
	"github.com/planetary-social/scuttlego/service/domain/transport"
	"github.com/stretchr/testify/require"
)

func TestNegotiator(t *testing.T) {
	someError := fixtures.SomeError()

	testCases := []struct {
		Name            string
		EbtError        error
		ChsError        error
		ExpectedError   error
		ExpectedEbtCall bool
		ExpectedChsCall bool
	}{
		{
			Name:            "replicate_exists_if_ebt_returns_an_error",
			EbtError:        someError,
			ChsError:        nil,
			ExpectedError:   someError,
			ExpectedEbtCall: true,
			ExpectedChsCall: false,
		},
		{
			Name:            "replicate_exists_if_ebt_returns_no_error",
			EbtError:        nil,
			ChsError:        nil,
			ExpectedError:   nil,
			ExpectedEbtCall: true,
			ExpectedChsCall: false,
		},
		{
			Name:            "replicate_calls_chs_if_ebt_returns_err_peer_does_not_support_ebt",
			EbtError:        replication.ErrPeerDoesNotSupportEBT,
			ChsError:        nil,
			ExpectedError:   nil,
			ExpectedEbtCall: true,
			ExpectedChsCall: true,
		},
		{
			Name:            "replicate_returns_an_error_if_chs_returns_an_error",
			EbtError:        replication.ErrPeerDoesNotSupportEBT,
			ChsError:        someError,
			ExpectedError:   someError,
			ExpectedEbtCall: true,
			ExpectedChsCall: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			logger := fixtures.TestLogger(t)
			ebtReplicator := newReplicatorMock()
			chsReplicator := newReplicatorMock()
			negotiator := replication.NewNegotiator(logger, ebtReplicator, chsReplicator)

			ebtReplicator.ReturnError = testCase.EbtError
			chsReplicator.ReturnError = testCase.ChsError

			ctx := fixtures.TestContext(t)
			conn := mocks.NewConnectionMock(ctx)
			peer := transport.MustNewPeer(fixtures.SomePublicIdentity(), conn)

			err := negotiator.Replicate(ctx, peer)
			if testCase.ExpectedError != nil {
				require.ErrorIs(t, err, testCase.ExpectedError)
			} else {
				require.NoError(t, err)
			}

			if testCase.ExpectedEbtCall {
				require.Equal(t, []replicatorMockReplicateCall{{Peer: peer}}, ebtReplicator.ReplicateCalls)
			} else {
				require.Empty(t, ebtReplicator.ReplicateCalls)
			}

			if testCase.ExpectedChsCall {
				require.Equal(t, []replicatorMockReplicateCall{{Peer: peer}}, chsReplicator.ReplicateCalls)
			} else {
				require.Empty(t, chsReplicator.ReplicateCalls)
			}
		})
	}
}

type replicatorMock struct {
	ReplicateCalls []replicatorMockReplicateCall
	ReturnError    error
}

func newReplicatorMock() *replicatorMock {
	return &replicatorMock{}
}

func (r *replicatorMock) Replicate(ctx context.Context, peer transport.Peer) error {
	r.ReplicateCalls = append(r.ReplicateCalls, replicatorMockReplicateCall{Peer: peer})
	return r.ReturnError
}

type replicatorMockReplicateCall struct {
	Peer transport.Peer
}
