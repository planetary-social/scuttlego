package badger

import (
	"context"
	"time"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/logging"
)

const badgerGarbageCollectionErrorDelay = 1 * time.Minute

type GarbageCollector struct {
	db     *badger.DB
	logger logging.Logger
}

func NewGarbageCollector(db *badger.DB, logger logging.Logger) *GarbageCollector {
	return &GarbageCollector{db: db, logger: logger.New("badger_garbage_collector")}
}

func (g *GarbageCollector) Run(ctx context.Context) error {
	for {
		if err := g.gc(); err != nil {
			if !errors.Is(err, badger.ErrNoRewrite) {
				g.logger.WithError(err).Error("error performing garbage collection")
			}

			select {
			case <-time.After(badgerGarbageCollectionErrorDelay):
				continue
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		g.logger.Debug("garbage collected a file")

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			continue
		}
	}
}

func (g *GarbageCollector) gc() error {
	return g.db.RunValueLogGC(0.5)
}
