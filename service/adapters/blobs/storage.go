package blobs

import (
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

const onlyForMe = 0700

const partialFileSuffix = ".part"

const filenameSeparator = "-"

type FilesystemStorage struct {
	path   string
	logger logging.Logger
}

func NewFilesystemStorage(path string, logger logging.Logger) (*FilesystemStorage, error) {
	s := &FilesystemStorage{
		path:   path,
		logger: logger,
	}

	if err := s.createStorage(); err != nil {
		return nil, errors.Wrap(err, "failed to create the storage directory")
	}

	if err := s.createTemporary(); err != nil {
		return nil, errors.Wrap(err, "failed to create the temporary directory")
	}

	if err := s.removeTemporaryFiles(); err != nil {
		return nil, errors.Wrap(err, "failed to remove old temporary files")
	}

	return s, nil
}

func (f FilesystemStorage) Store(id refs.Blob, r io.Reader) error {
	hexRef := hex.EncodeToString(id.Bytes())

	pattern := fmt.Sprintf("%s%s%d%s*%s", hexRef, filenameSeparator, time.Now().Unix(), filenameSeparator, partialFileSuffix)

	tmpFile, err := os.CreateTemp(f.dirTemporary(), pattern)
	if err != nil {
		return errors.Wrap(err, "could not create a temporary file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	h := blobs.NewHasher()

	if _, err := io.Copy(io.MultiWriter(tmpFile, h), io.LimitReader(r, blobs.MaxBlobSize().InBytes())); err != nil {
		return errors.Wrap(err, "failed to copy contents to a temporary file")
	}

	if err := blobs.Verify(id, h); err != nil {
		return errors.Wrap(err, "failed to verify the file")
	}

	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "failed to close the temporary file")
	}

	oldName := tmpFile.Name()
	newName := f.pathStorage(hexRef)

	if err := os.Rename(oldName, newName); err != nil {
		return errors.Wrap(err, "failed to rename the file")
	}

	return nil
}

func (f FilesystemStorage) createStorage() error {
	return os.MkdirAll(f.dirStorage(), onlyForMe)
}

func (f FilesystemStorage) createTemporary() error {
	return os.MkdirAll(f.dirTemporary(), onlyForMe)
}

func (f FilesystemStorage) dirTemporary() string {
	return path.Join(f.path, "temporary")
}

func (f FilesystemStorage) dirStorage() string {
	return path.Join(f.path, "storage")
}

func (f FilesystemStorage) pathStorage(name string) string {
	return path.Join(f.dirStorage(), name)
}

func (f FilesystemStorage) removeTemporaryFiles() error {
	return filepath.WalkDir(f.dirTemporary(), func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if err := os.Remove(path); err != nil {
				return errors.Wrap(err, "could not remove one of the old temporary files")
			}
		}
		return nil
	})
}
