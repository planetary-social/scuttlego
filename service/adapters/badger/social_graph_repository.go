package badger

import (
	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	jsoniter "github.com/json-iterator/go"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/graph"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

var (
	socialGraphRepositoryBucketGraph = utils.MustNewKeyComponent([]byte("graph"))
)

type UpdateContactFn func(*feeds.Contact) error

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

func (s *SocialGraphRepository) UpdateContact(author, target refs.Identity, f UpdateContactFn) error {
	bucket := s.getFeedBucket(author)

	contact, err := s.loadOrCreateContact(bucket, author, target)
	if err != nil {
		return errors.Wrap(err, "failed to load the existing contact")
	}

	if err := f(contact); err != nil {
		return errors.Wrap(err, "error updating the contact")
	}

	b, err := jsoniter.Marshal(newStoredContact(contact))
	if err != nil {
		return errors.Wrap(err, "could not marshal contact")
	}

	return bucket.Set(s.key(target), b)
}

func (s *SocialGraphRepository) Remove(author refs.Identity) error {
	bucket := s.getFeedBucket(author)
	return bucket.DeleteBucket()
}

func (s *SocialGraphRepository) GetContacts(node refs.Identity) ([]*feeds.Contact, error) {
	bucket := s.getFeedBucket(node)

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
	if err := jsoniter.Unmarshal(storedValue, &c); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal the existing value")
	}
	return feeds.NewContactFromHistory(author, target, c.Following, c.Blocking)
}

func (s *SocialGraphRepository) getFeedBucket(ref refs.Identity) utils.Bucket {
	return utils.MustNewBucket(s.tx, s.pathFunc(ref))
}

func (s *SocialGraphRepository) pathFunc(who refs.Identity) utils.Key {
	return utils.MustNewKey(
		socialGraphRepositoryBucketGraph,
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
