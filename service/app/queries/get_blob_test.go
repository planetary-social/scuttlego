package queries_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/di"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/app/queries"
	"github.com/stretchr/testify/require"
)

func TestGetBlob(t *testing.T) {
	q, err := di.BuildTestQueries()
	require.NoError(t, err)

	id := fixtures.SomeRefBlob()

	q.BlobStorage.MockBlob(id, fixtures.SomeBytes())

	query := queries.GetBlob{
		Id: id,
	}

	rc, err := q.Queries.GetBlob.Handle(query)
	require.NoError(t, err)
	require.NotNil(t, rc)
}
