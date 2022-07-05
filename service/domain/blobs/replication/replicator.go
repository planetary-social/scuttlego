package replication

import (
	"context"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/transport"
)

type ReplicationManager interface {
	HandleOutgoingCreateWantsRequest(ctx context.Context, ch <-chan messages.BlobWithSizeOrWantDistance, peer transport.Peer) error
}

type Replicator struct {
	manager ReplicationManager
}

func NewReplicator(manager ReplicationManager) *Replicator {
	return &Replicator{
		manager: manager,
	}
}

func (r Replicator) Replicate(ctx context.Context, peer transport.Peer) error {
	request, err := messages.NewBlobsCreateWants()
	if err != nil {
		return errors.Wrap(err, "could not create a request")
	}

	rs, err := peer.Conn().PerformRequest(ctx, request)
	if err != nil {
		return errors.Wrap(err, "could not perform a request")
	}

	ch := make(chan messages.BlobWithSizeOrWantDistance)
	defer close(ch)

	if err := r.manager.HandleOutgoingCreateWantsRequest(ctx, ch, peer); err != nil {
		return errors.Wrap(err, "could not handle outgoing create wants request")
	}

	for response := range rs.Channel() {
		if err := response.Err; err != nil {
			return errors.Wrap(err, "received an error")
		}

		sizeOrWantDistances, err := messages.NewBlobsCreateWantsResponseFromBytes(response.Value.Bytes())
		if err != nil {
			return errors.Wrap(err, "failed to unmarshal the response")
		}

		for _, blobWithSizeOrWantDistance := range sizeOrWantDistances.List() {
			select {
			case ch <- blobWithSizeOrWantDistance:
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}

	return nil
}
