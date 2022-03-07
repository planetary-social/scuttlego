package adapters

import (
	"encoding/json"
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/refs"
	"go.etcd.io/bbolt"
)

type SocialGraphRepository struct {
	tx *bbolt.Tx
}

func NewSocialGraphRepository(tx *bbolt.Tx) *SocialGraphRepository {
	return &SocialGraphRepository{tx: tx}
}

func (r SocialGraphRepository) HasContact(contact refs.Identity) (bool, error) {
	return true, nil
}

func (s *SocialGraphRepository) Follow(who, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Following = true
	})
}

func (s *SocialGraphRepository) Unfollow(who, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Following = false
	})
}

func (s *SocialGraphRepository) Block(who, contact refs.Identity) error {
	return s.modifyContact(who, contact, func(c *storedContact) {
		c.Blocking = true
	})
}

func (s *SocialGraphRepository) modifyContact(who, contact refs.Identity, f func(c *storedContact)) error {
	bucket, err := s.createFeedBucket(who)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	key := s.key(contact)

	var c storedContact

	value := bucket.Get(key)
	if value != nil {
		if err := json.Unmarshal(value, &c); err != nil {
			return errors.Wrap(err, "failed to unmarshal the existing value")
		}
	}

	f(&c)

	b, err := json.Marshal(c)
	if err != nil {
		return errors.Wrap(err, "could not marshal contact")
	}

	return bucket.Put(key, b)

}

func (r *SocialGraphRepository) createFeedBucket(ref refs.Identity) (bucket *bbolt.Bucket, err error) {
	bucketNames := r.pathFunc(ref)
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return createBucket(r.tx, bucketNames)
}

func (r *SocialGraphRepository) getFeedBucket(ref refs.Identity) (*bbolt.Bucket, error) {
	bucketNames := r.pathFunc(ref)
	if len(bucketNames) == 0 {
		return nil, errors.New("path func returned an empty slice")
	}

	return getBucket(r.tx, bucketNames), nil
}

func (r *SocialGraphRepository) pathFunc(who refs.Identity) []BucketName {
	return []BucketName{
		BucketName("graph"),
		BucketName(who.String()),
	}
}

func (s *SocialGraphRepository) key(target refs.Identity) BucketName {
	return []byte(target.String())
}

type storedContact struct {
	Following bool `json:"following"`
	Blocking  bool `json:"blocking"`
}
