package bolt

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

const socialGraphRepositoryBucket = "graph"

type SocialGraphRepository struct {
	local identity.Public
	hops  graph.Hops
	tx    *bbolt.Tx
}

func NewSocialGraphRepository(tx *bbolt.Tx, local identity.Public, hops graph.Hops) *SocialGraphRepository {
	return &SocialGraphRepository{tx: tx, local: local, hops: hops}
}

func (s *SocialGraphRepository) GetSocialGraph() (*graph.SocialGraph, error) {
	localRef, err := refs.NewIdentityFromPublic(s.local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a local ref")
	}
	return graph.NewSocialGraph(localRef, s.hops, s)
}

type UpdateContactFn func(*feeds.Contact) error

func (s *SocialGraphRepository) UpdateContact(author, target refs.Identity, f UpdateContactFn) error {
	bucket, err := s.createFeedBucket(author)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	key := s.key(target)
	value := bucket.Get(key)

	contact, err := s.loadOrCreateContact(author, target, value)
	if err != nil {
		return errors.Wrap(err, "failed to load the existing contact")
	}

	if err := f(contact); err != nil {
		return errors.Wrap(err, "error updating the contact")
	}

	b, err := json.Marshal(newStoredContact(contact))
	if err != nil {
		return errors.Wrap(err, "could not marshal contact")
	}

	return bucket.Put(key, b)
}

func (s *SocialGraphRepository) Remove(author refs.Identity) error {
	return s.deleteFeedBucket(author)
}

func (s *SocialGraphRepository) GetContacts(node refs.Identity) ([]*feeds.Contact, error) {
	bucket, err := s.getFeedBucket(node)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	var result []*feeds.Contact

	if err := bucket.ForEach(func(k, v []byte) error {
		targetRef, err := refs.NewIdentity(string(k)) // todo is this certainly a copy or are we reusing the slice illegally
		if err != nil {
			return errors.Wrap(err, "could not create contact ref")
		}

		contact, err := s.loadContact(node, targetRef, v)
		if err != nil {
			return errors.Wrap(err, "failed to load the contact")
		}

		result = append(result, contact)

		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "iteration failed")
	}

	return result, nil
}

func (s *SocialGraphRepository) loadOrCreateContact(author, target refs.Identity, storedValue []byte) (*feeds.Contact, error) {
	if storedValue != nil {
		return s.loadContact(author, target, storedValue)
	}
	return feeds.NewContact(author, target)
}

func (s *SocialGraphRepository) loadContact(author, target refs.Identity, storedValue []byte) (*feeds.Contact, error) {
	var c storedContact
	if err := json.Unmarshal(storedValue, &c); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the existing value")
	}
	return feeds.NewContactFromHistory(author, target, c.Following, c.Blocking)
}

func (s *SocialGraphRepository) createFeedBucket(ref refs.Identity) (*bbolt.Bucket, error) {
	return utils.CreateBucket(s.tx, s.pathFunc(ref))
}

func (s *SocialGraphRepository) getFeedBucket(ref refs.Identity) (*bbolt.Bucket, error) {
	return utils.GetBucket(s.tx, s.pathFunc(ref))
}

func (s *SocialGraphRepository) deleteFeedBucket(ref refs.Identity) error {
	return utils.DeleteBucket(
		s.tx,
		[]utils.BucketName{
			utils.BucketName(socialGraphRepositoryBucket),
		},
		utils.BucketName(ref.String()),
	)
}

func (s *SocialGraphRepository) pathFunc(who refs.Identity) []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName(socialGraphRepositoryBucket),
		utils.BucketName(who.String()),
	}
}

func (s *SocialGraphRepository) key(target refs.Identity) utils.BucketName {
	return []byte(target.String())
}

type storedContact struct {
	Following bool `json:"following"`
	Blocking  bool `json:"blocking"`
}

func newStoredContact(contact *feeds.Contact) storedContact {
	return storedContact{
		Following: contact.Following(),
		Blocking:  contact.Blocking(),
	}
}
