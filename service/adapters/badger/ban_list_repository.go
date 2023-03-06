package badger

import (
	"fmt"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var (
	bucketBanListKeyComponent = utils.MustNewKeyComponent([]byte("ban_list"))

	bucketBanListHashesKey = utils.MustNewKey(
		bucketBanListKeyComponent,
		utils.MustNewKeyComponent([]byte("hashes")),
	)

	bucketBanListMappingKey = utils.MustNewKey(
		bucketBanListKeyComponent,
		utils.MustNewKeyComponent([]byte("mapping")),
	)
)

type BanListHasher interface {
	HashForFeed(refs.Feed) (bans.Hash, error)
}

type BanListRepository struct {
	tx     *badger.Txn
	hasher BanListHasher
}

func NewBanListRepository(
	tx *badger.Txn,
	hasher BanListHasher,
) *BanListRepository {
	return &BanListRepository{
		tx:     tx,
		hasher: hasher,
	}
}

func (b BanListRepository) Add(hash bans.Hash) error {
	bucket := b.hashesBucket()

	if err := bucket.Set(hash.Bytes(), nil); err != nil {
		return errors.Wrap(err, "put failed")
	}

	return nil
}

func (b BanListRepository) Remove(hash bans.Hash) error {
	bucket := b.hashesBucket()

	if err := bucket.Delete(hash.Bytes()); err != nil {
		return errors.Wrap(err, "delete failed")
	}

	return nil
}

func (b BanListRepository) Clear() error {
	bucket := b.hashesBucket()

	return bucket.DeleteBucket()
}

func (b BanListRepository) Contains(hash bans.Hash) (bool, error) {
	bucket := b.hashesBucket()

	if _, err := bucket.Get(hash.Bytes()); err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return false, nil
		}

		return false, errors.Wrap(err, "get failed")
	}

	return true, nil
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

	bucket := b.mappingBucket()

	v, err := jsoniter.Marshal(storedBannableRef{
		Typ: storedBannableRefTypeFeed,
		Ref: ref.String(),
	})
	if err != nil {
		return errors.Wrap(err, "json marshal failed")
	}

	if err := bucket.Set(hash.Bytes(), v); err != nil {
		return errors.Wrap(err, "bucket set failed")
	}

	return nil
}

func (b BanListRepository) RemoveFeedMapping(ref refs.Feed) error {
	hash, err := b.hasher.HashForFeed(ref)
	if err != nil {
		return errors.Wrap(err, "failed to generate a hash")
	}

	bucket := b.mappingBucket()

	if err := bucket.Delete(hash.Bytes()); err != nil {
		return errors.Wrap(err, "bucket delete failed")
	}

	return nil
}

func (b BanListRepository) LookupMapping(hash bans.Hash) (commands.BannableRef, error) {
	bucket := b.mappingBucket()

	item, err := bucket.Get(hash.Bytes())
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return commands.BannableRef{}, commands.ErrBanListMappingNotFound
		}

		return commands.BannableRef{}, errors.Wrap(err, "get failed")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return commands.BannableRef{}, errors.Wrap(err, "error getting item value")
	}

	var v storedBannableRef
	if err := jsoniter.Unmarshal(value, &v); err != nil {
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

func (b BanListRepository) List() ([]bans.Hash, error) {
	bucket := b.hashesBucket()

	var result []bans.Hash
	if err := bucket.ForEach(func(item utils.Item) error {
		keyInBucket, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error getting key in bucket")
		}

		h, err := bans.NewHash(keyInBucket.Bytes())
		if err != nil {
			return errors.Wrap(err, "error creating a hash")
		}

		result = append(result, h)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}
	return result, nil
}

func (b BanListRepository) hashesBucket() utils.Bucket {
	return utils.MustNewBucket(b.tx, bucketBanListHashesKey)
}

func (b BanListRepository) mappingBucket() utils.Bucket {
	return utils.MustNewBucket(b.tx, bucketBanListMappingKey)
}

const (
	storedBannableRefTypeFeed = "feed"
)

type storedBannableRef struct {
	Typ string `json:"typ"`
	Ref string `json:"ref"`
}
