package bolt

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

var (
	bucketBanList        = utils.BucketName("ban_list")
	bucketBanListHashes  = utils.BucketName("hashes")
	bucketBanListMapping = utils.BucketName("mapping")
)

type BanListHasher interface {
	HashForFeed(refs.Feed) (bans.Hash, error)
}

type BanListRepository struct {
	tx     *bbolt.Tx
	hasher BanListHasher
}

func NewBanListRepository(
	tx *bbolt.Tx,
	hasher BanListHasher,
) *BanListRepository {
	return &BanListRepository{
		tx:     tx,
		hasher: hasher,
	}
}

func (b BanListRepository) Add(hash bans.Hash) error {
	bucket, err := utils.CreateBucket(b.tx, b.hashesBucketPath())
	if err != nil {
		return errors.Wrap(err, "create bucket error")
	}

	if err := bucket.Put(hash.Bytes(), nil); err != nil {
		return errors.Wrap(err, "bucket put failed")
	}

	return nil
}

func (b BanListRepository) Remove(hash bans.Hash) error {
	bucket, err := utils.GetBucket(b.tx, b.hashesBucketPath())
	if err != nil {
		return errors.Wrap(err, "get bucket error")
	}

	if bucket == nil {
		return nil
	}

	if err := bucket.Delete(hash.Bytes()); err != nil {
		return errors.Wrap(err, "bucket delete failed")
	}

	return nil
}

func (b BanListRepository) Contains(hash bans.Hash) (bool, error) {
	bucket, err := utils.GetBucket(b.tx, b.hashesBucketPath())
	if err != nil {
		return false, errors.Wrap(err, "get bucket error")
	}

	if bucket == nil {
		return false, nil
	}

	foundKey, _ := bucket.Cursor().Seek(hash.Bytes())
	return bytes.Equal(foundKey, hash.Bytes()), nil
}

func (b BanListRepository) ContainsFeed(ref refs.Feed) (bool, error) {
	hash, err := b.hasher.HashForFeed(ref)
	if err != nil {
		return false, errors.Wrap(err, "failed to generate a hash")
	}

	return b.Contains(hash)
}

func (b BanListRepository) CreateFeedMapping(ref refs.Feed) error {
	hash, err := b.hasher.HashForFeed(ref)
	if err != nil {
		return errors.Wrap(err, "failed to generate a hash")
	}

	bucket, err := utils.CreateBucket(b.tx, b.mappingBucketPath())
	if err != nil {
		return errors.Wrap(err, "create bucket error")
	}

	v, err := json.Marshal(storedBannableRef{
		Typ: storedBannableRefTypeFeed,
		Ref: ref.String(),
	})
	if err != nil {
		return errors.Wrap(err, "json marshal failed")
	}

	if err := bucket.Put(hash.Bytes(), v); err != nil {
		return errors.Wrap(err, "bucket put failed")
	}

	return nil
}

func (b BanListRepository) RemoveFeedMapping(ref refs.Feed) error {
	hash, err := b.hasher.HashForFeed(ref)
	if err != nil {
		return errors.Wrap(err, "failed to generate a hash")
	}

	bucket, err := utils.GetBucket(b.tx, b.mappingBucketPath())
	if err != nil {
		return errors.Wrap(err, "get bucket error")
	}

	if bucket == nil {
		return nil
	}

	if err := bucket.Delete(hash.Bytes()); err != nil {
		return errors.Wrap(err, "bucket put failed")
	}

	return nil
}

func (b BanListRepository) LookupMapping(hash bans.Hash) (commands.BannableRef, error) {
	bucket, err := utils.GetBucket(b.tx, b.mappingBucketPath())
	if err != nil {
		return commands.BannableRef{}, errors.Wrap(err, "get bucket error")
	}

	if bucket == nil {
		return commands.BannableRef{}, commands.ErrBanListMappingNotFound
	}

	j := bucket.Get(hash.Bytes())
	if j == nil {
		return commands.BannableRef{}, commands.ErrBanListMappingNotFound
	}

	var v storedBannableRef
	if err := json.Unmarshal(j, &v); err != nil {
		return commands.BannableRef{}, errors.Wrap(err, "unmarshal failed")
	}

	switch v.Typ {
	case storedBannableRefTypeFeed:
		ref, err := refs.NewFeed(v.Ref)
		if err != nil {
			return commands.BannableRef{}, errors.Wrap(err, "failed to create a feed ref")
		}
		return commands.NewBannableRef(ref)
	default:
		return commands.BannableRef{}, fmt.Errorf("unknown type: '%s'", v.Typ)
	}
}

func (b BanListRepository) hashesBucketPath() []utils.BucketName {
	return []utils.BucketName{
		bucketBanList,
		bucketBanListHashes,
	}
}

func (b BanListRepository) mappingBucketPath() []utils.BucketName {
	return []utils.BucketName{
		bucketBanList,
		bucketBanListMapping,
	}
}

const (
	storedBannableRefTypeFeed = "feed"
)

type storedBannableRef struct {
	Typ string `json:"typ"`
	Ref string `json:"ref"`
}
