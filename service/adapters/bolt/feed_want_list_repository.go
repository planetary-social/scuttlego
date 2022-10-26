package bolt

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

type FeedWantListRepository struct {
	repo *WantListRepository
}

func NewFeedWantListRepository(
	tx *bbolt.Tx,
	currentTimeProvider commands.CurrentTimeProvider,
) *FeedWantListRepository {
	return &FeedWantListRepository{
		repo: NewWantListRepository(
			tx,
			currentTimeProvider,
			[]utils.BucketName{
				utils.BucketName("feed_want_list"),
			},
		),
	}
}

func (r FeedWantListRepository) Add(id refs.Feed, until time.Time) error {
	return r.repo.Add(id.String(), until)
}

func (r FeedWantListRepository) List() ([]refs.Feed, error) {
	var result []refs.Feed

	resultStrings, err := r.repo.List()
	if err != nil {
		return nil, errors.Wrap(err, "error querying the underlying repo")
	}

	for _, resultString := range resultStrings {
		ref, err := refs.NewFeed(resultString)
		if err != nil {
			return nil, errors.Wrap(err, "could not create a ref")
		}

		result = append(result, ref)
	}

	return result, nil
}
