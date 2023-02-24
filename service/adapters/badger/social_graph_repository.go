package badger

import (
	"encoding/json"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const socialGraphRepositoryBucket = "graph"

type SocialGraphRepository struct {
	tx            *badger.Txn
	local         identity.Public
	hops          graph.Hops
	banList       *BanListRepository
	banListHasher BanListHasher
}

func NewSocialGraphRepository(
	tx *badger.Txn,
	local identity.Public,
	hops graph.Hops,
	banList *BanListRepository,
	banListHasher BanListHasher,
) *SocialGraphRepository {
	return &SocialGraphRepository{
		tx:            tx,
		local:         local,
		hops:          hops,
		banList:       banList,
		banListHasher: banListHasher,
	}
}

func (s *SocialGraphRepository) GetSocialGraph() (graph.SocialGraph, error) {
	localRef, err := refs.NewIdentityFromPublic(s.local)
	if err != nil {
		return graph.SocialGraph{}, errors.Wrap(err, "could not create a local ref")
	}
	banList, err := graph.NewCachedBanList(s.banListHasher, s.banList)
	if err != nil {
		return graph.SocialGraph{}, errors.Wrap(err, "could not create a cached ban list")
	}
	return graph.NewSocialGraphBuilder(s, banList, s.hops, localRef).Build()
}

func (s *SocialGraphRepository) GetSocialGraphBuilder() (*graph.SocialGraphBuilder, error) {
	localRef, err := refs.NewIdentityFromPublic(s.local)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a local ref")
	}
	banList, err := graph.NewCachedBanList(s.banListHasher, s.banList)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a cached ban list")
	}
	return graph.NewSocialGraphBuilder(s, banList, s.hops, localRef), nil
}

type UpdateContactFn func(*feeds.Contact) error

func (s *SocialGraphRepository) UpdateContact(author, target refs.Identity, f UpdateContactFn) error {
	bucket, err := s.createFeedBucket(author)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	contact, err := s.loadOrCreateContact(bucket, author, target)
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

	return bucket.Set(s.key(target), b)
}

func (s *SocialGraphRepository) Remove(author refs.Identity) error {
	return s.deleteFeedBucket(author)
}

func (s *SocialGraphRepository) GetContacts(node refs.Identity) ([]*feeds.Contact, error) {
	bucket, err := s.createFeedBucket(node)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	var result []*feeds.Contact

	if err := bucket.ForEach(func(item utils.Item) error {
		keyInBucket, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "error determining key in bucket")
		}

		targetRef, err := refs.NewIdentity(string(keyInBucket.Bytes()))
		if err != nil {
			return errors.Wrap(err, "could not create contact ref")
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return errors.Wrap(err, "could not get the value")
		}

		contact, err := s.loadContact(node, targetRef, val)
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

func (s *SocialGraphRepository) loadOrCreateContact(bucket utils.Bucket, author, target refs.Identity) (*feeds.Contact, error) {
	item, err := bucket.Get(s.key(target))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return feeds.NewContact(author, target)
		}
		return nil, errors.Wrap(err, "error reading the existing contact")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return nil, errors.Wrap(err, "could not get the value")
	}

	contact, err := s.loadContact(author, target, value)
	if err != nil {
		return nil, errors.Wrap(err, "error loading the contact")
	}

	return contact, nil
}

func (s *SocialGraphRepository) loadContact(author, target refs.Identity, storedValue []byte) (*feeds.Contact, error) {
	var c storedContact
	if err := json.Unmarshal(storedValue, &c); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the existing value")
	}
	return feeds.NewContactFromHistory(author, target, c.Following, c.Blocking)
}

func (s *SocialGraphRepository) createFeedBucket(ref refs.Identity) (utils.Bucket, error) {
	return utils.NewBucket(s.tx, s.pathFunc(ref))
}

func (s *SocialGraphRepository) deleteFeedBucket(ref refs.Identity) error {
	bucket, err := s.createFeedBucket(ref)
	if err != nil {
		return errors.Wrap(err, "error creating the bucket")
	}
	return bucket.DeleteBucket()
}

func (s *SocialGraphRepository) pathFunc(who refs.Identity) utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte(socialGraphRepositoryBucket)),
		utils.MustNewKeyComponent([]byte(who.String())),
	)
}

func (s *SocialGraphRepository) key(target refs.Identity) []byte {
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
