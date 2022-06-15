package blobs

import (
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/refs"
)

type FilesystemStorage struct {
	path string
}

func NewFilesystemStorage(path string) (*FilesystemStorage, error) {
	s := &FilesystemStorage{path: path}

	if err := s.createStorage(); err != nil {
		return nil, errors.Wrap(err, "failed to create the storage directory")
	}

	if err := s.createFinished(); err != nil {
		return nil, errors.Wrap(err, "failed to create the finished directory")
	}

	if err := s.recreateTemporary(); err != nil {
		return nil, errors.Wrap(err, "failed to recreate the temporary directory")
	}

	return s, nil
}

const partSuffix = ".part"

func (f FilesystemStorage) Save(id refs.Blob, r io.Reader) error {
	hexRef := hex.EncodeToString(id.Bytes())

	tmpFile, err := os.CreateTemp(f.dirTemporary(), fmt.Sprintf("%s-*%s", hexRef, partSuffix))
	if err != nil {
		return errors.Wrap(err, "could not create a temporary file")
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	verifier := blobs.NewVerifier(id)

	if _, err := io.Copy(io.MultiWriter(tmpFile, verifier), io.LimitReader(r, blobs.MaxBlobSize().InBytes())); err != nil {
		return errors.Wrap(err, "failed to copy contents to a temporary file")
	}

	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "failed to close the temporary file")
	}

	oldPath := tmpFile.Name()
	_, oldFilename := path.Split(tmpFile.Name())
	newPath := f.pathFinished(strings.TrimSuffix(oldFilename, partSuffix))

	if err := os.Rename(oldPath, newPath); err != nil {
		return errors.Wrap(err, "failed to move the temporary file")
	}

	return nil
}

const onlyForMe = 0700

func (f FilesystemStorage) createStorage() error {
	return os.MkdirAll(f.dirStorage(), onlyForMe)
}

func (f FilesystemStorage) createFinished() error {
	return os.MkdirAll(f.dirFinished(), onlyForMe)
}

func (f FilesystemStorage) recreateTemporary() error {
	if err := os.RemoveAll(f.dirTemporary()); err != nil {
		return errors.Wrap(err, "failed to remove the temporary directory")
	}

	if err := os.MkdirAll(f.dirTemporary(), onlyForMe); err != nil {
		return errors.Wrap(err, "failed to create the temporary directory")
	}

	return nil
}

func (f FilesystemStorage) dirTemporary() string {
	return path.Join(f.path, "temporary")
}

func (f FilesystemStorage) pathTemporary(name string) string {
	return path.Join(f.dirTemporary(), name)
}

func (f FilesystemStorage) dirFinished() string {
	return path.Join(f.path, "finished")
}

func (f FilesystemStorage) pathFinished(name string) string {
	return path.Join(f.dirFinished(), name)
}

func (f FilesystemStorage) dirStorage() string {
	return path.Join(f.path, "storage")
}
