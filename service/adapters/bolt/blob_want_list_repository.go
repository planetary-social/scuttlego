package bolt

import (
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

type BlobWantListRepository struct {
	repo *WantListRepository
}

func NewBlobWantListRepository(
	tx *bbolt.Tx,
	currentTimeProvider commands.CurrentTimeProvider,
) *BlobWantListRepository {
	return &BlobWantListRepository{
		repo: NewWantListRepository(
			tx,
			currentTimeProvider,
			[]utils.BucketName{
				utils.BucketName("blob_want_list"),
			},
		),
	}
}

func (r BlobWantListRepository) Add(id refs.Blob, until time.Time) error {
	return r.repo.Add(id.String(), until)
}

func (r BlobWantListRepository) Contains(id refs.Blob) (bool, error) {
	return r.repo.Contains(id.String())
}

func (r BlobWantListRepository) Delete(id refs.Blob) error {
	return r.repo.Delete(id.String())
}

func (r BlobWantListRepository) List() ([]refs.Blob, error) {
	var result []refs.Blob

	resultStrings, err := r.repo.List()
	if err != nil {
		return nil, errors.Wrap(err, "error querying the underlying repo")
	}

	for _, resultString := range resultStrings {
		ref, err := refs.NewBlob(resultString)
		if err != nil {
			return nil, errors.Wrap(err, "could not create a ref")
		}

		result = append(result, ref)
	}

	return result, nil
}
