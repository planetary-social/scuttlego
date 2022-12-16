package migrations_test

import (
	"os"
	"path"
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/migrations"
	"github.com/stretchr/testify/require"
)

func TestGoSSBRepoReader_GetMessages_ReturnsNoResultsIfDirectoryDoesNotExist(t *testing.T) {
	reader := migrations.NewGoSSBRepoReader(
		fixtures.TestLogger(t),
	)

	ctx := fixtures.TestContext(t)
	directory := fixtures.Directory(t)
	nonexistentDirectory := path.Join(directory, "nonexistent")

	_, err := os.Stat(nonexistentDirectory)
	require.ErrorIs(t, err, os.ErrNotExist)

	ch, err := reader.GetMessages(ctx, nonexistentDirectory, nil)
	require.NoError(t, err)

	for range ch {
		t.Fatal("got a value")
	}

	_, err = os.Stat(nonexistentDirectory)
	require.ErrorIs(t, err, os.ErrNotExist)
}
