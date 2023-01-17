package tests_test

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/app/commands/tests"
	"github.com/stretchr/testify/require"
)

const (
	testdata    = "testdata"
	testdataOld = "gossb_repository_in_old_format"
	testdataNew = "gossb_repository_in_new_format"
)

func TestDeleteGoSSBRepositoryInOldFormat(t *testing.T) {
	testCases := []struct {
		Name              string
		TestdataDirectory string
		ShouldRemove      bool
	}{
		{
			Name:              "old_format_repo_should_be_removed",
			TestdataDirectory: testdataOld,
			ShouldRemove:      true,
		},
		{
			Name:              "new_format_repo_should_not_be_removed",
			TestdataDirectory: testdataNew,
			ShouldRemove:      false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ts := tests.BuildCommandIntegrationTest(t)

			testDir := currentDirectory()
			tmpDirectory := fixtures.Directory(t)

			testdataDirectory := filepath.Join(testDir, testdata, testCase.TestdataDirectory)

			testdataCopy := filepath.Join(tmpDirectory, "repo")

			err := copyDirectory(testdataDirectory, testdataCopy)
			require.NoError(t, err)

			ctx := fixtures.TestContext(t)

			ok, err := dirExists(testdataCopy)
			require.NoError(t, err)
			require.True(t, ok)

			cmd, err := commands.NewDeleteGoSSBRepositoryInOldFormat(testdataCopy)
			require.NoError(t, err)

			err = ts.DeleteGoSSBRepositoryInOldFormat.Handle(ctx, cmd)
			require.NoError(t, err)

			ok, err = dirExists(testdataCopy)
			require.NoError(t, err)
			require.Equal(t, ok, !testCase.ShouldRemove)
		})
	}
}

func dirExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return false, nil
		}
		return false, err
	}
	if !info.IsDir() {
		return false, errors.New("not a dir")
	}
	return true, nil
}

func copyDirectory(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "received an error")
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return errors.Wrap(err, "error creating relative path")
		}

		target := filepath.Join(dst, rel)

		if info.IsDir() {
			if err := os.MkdirAll(target, 0700); err != nil {
				return errors.Wrap(err, "error creating directory")
			}
		} else {
			containingDirectory := filepath.Dir(target)

			if err := os.MkdirAll(containingDirectory, 0700); err != nil {
				return errors.Wrap(err, "error creating directory")
			}

			srcF, err := os.Open(path)
			if err != nil {
				return errors.Wrap(err, "error opening a file")
			}

			dstF, err := os.Create(target)
			if err != nil {
				return errors.Wrap(err, "error creating a file")
			}

			_, err = io.Copy(dstF, srcF)
			if err != nil {
				return errors.Wrap(err, "error copying a file")
			}

			if err := srcF.Close(); err != nil {
				return errors.Wrap(err, "error closing src file")
			}

			if err := dstF.Close(); err != nil {
				return errors.Wrap(err, "error closing dst file")
			}
		}

		return nil
	})
}

func currentDirectory() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Dir(filename)
}
