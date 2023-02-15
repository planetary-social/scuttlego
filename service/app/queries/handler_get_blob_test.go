package queries_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestGetBlob(t *testing.T) {
	id := fixtures.SomeRefBlob()
	data := fixtures.SomeBytes()
	correctSize := blobs.MustNewSize(int64(len(data)))
	incorrectSize := blobs.MustNewSize(correctSize.InBytes() + int64(fixtures.SomePositiveInt()))

	testCases := []struct {
		Name          string
		QueryId       refs.Blob
		QuerySize     *blobs.Size
		QueryMax      *blobs.Size
		ExpectedError error
	}{
		{
			Name:          "only_id",
			QueryId:       id,
			QuerySize:     nil,
			QueryMax:      nil,
			ExpectedError: nil,
		},
		{
			Name:          "id_and_correct_size",
			QueryId:       id,
			QuerySize:     internal.Ptr(correctSize),
			QueryMax:      nil,
			ExpectedError: nil,
		},
		{
			Name:          "id_and_incorrect_size",
			QueryId:       id,
			QuerySize:     internal.Ptr(incorrectSize),
			QueryMax:      nil,
			ExpectedError: errors.New("blob size doesn't match the provided size"),
		},
		{
			Name:          "id_and_max_above_size",
			QueryId:       id,
			QuerySize:     nil,
			QueryMax:      internal.Ptr(blobs.MustNewSize(correctSize.InBytes() + 1)),
			ExpectedError: nil,
		},
		{
			Name:          "id_and_max_equal_to_size",
			QueryId:       id,
			QuerySize:     nil,
			QueryMax:      internal.Ptr(correctSize),
			ExpectedError: nil,
		},
		{
			Name:          "id_and_max_below_size",
			QueryId:       id,
			QuerySize:     nil,
			QueryMax:      internal.Ptr(blobs.MustNewSize(correctSize.InBytes() - 1)),
			ExpectedError: errors.New("blob is larger than the provided max size"),
		},
		{
			Name:          "size_wins_over_max",
			QueryId:       id,
			QuerySize:     internal.Ptr(blobs.MustNewSize(1)),
			QueryMax:      internal.Ptr(blobs.MustNewSize(1)),
			ExpectedError: errors.New("blob size doesn't match the provided size"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			q, err := di.BuildTestQueries(t)
			require.NoError(t, err)

			q.BlobStorage.MockBlob(id, data)

			query, err := queries.NewGetBlob(
				testCase.QueryId,
				testCase.QuerySize,
				testCase.QueryMax,
			)
			require.NoError(t, err)

			rc, err := q.Queries.GetBlob.Handle(query)
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
				require.Nil(t, rc)
			} else {
				require.NoError(t, err)
				require.NotNil(t, rc)
			}
		})
	}
}
