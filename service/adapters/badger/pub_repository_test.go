package badger_test

import (
	"fmt"
	"sort"
	"testing"

	"github.com/planetary-social/scuttlego/di"
	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/adapters/badger"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestPubRepository_Delete_SamePubSameAuthorSameAddressSameMessage(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	pub := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.PubRepository.Put(pub)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Content().Key()}, pubs)

		sources, err := adapters.PubRepository.ListAddresses(pub.Content().Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub.Content().Host(), pub.Content().Port()),
					Sources: []badger.PubAddressSource{
						{
							Identity: pub.Who(),
							Messages: []refs.Message{
								pub.Message(),
							},
						},
					},
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.PubRepository.Delete(pub.Message())
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub.Message())
		require.NoError(t, err)
		require.Empty(t, pubs)

		sources, err := adapters.PubRepository.ListAddresses(pub.Content().Key())
		require.NoError(t, err)
		require.Empty(t, sources)

		return nil
	})
	require.NoError(t, err)
}

func TestPubRepository_Delete_SamePubSameAuthorSameAddressDifferentMessage(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	authorIdenRef := fixtures.SomeRefIdentity()
	pub := content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt())

	pub1 := feeds.NewPubToSave(
		authorIdenRef,
		fixtures.SomeRefMessage(),
		pub,
	)

	pub2 := feeds.NewPubToSave(
		authorIdenRef,
		fixtures.SomeRefMessage(),
		pub,
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Key()}, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Key()}, pubs)

		expectedMessages := []refs.Message{
			pub1.Message(),
			pub2.Message(),
		}

		sort.Slice(expectedMessages, func(i, j int) bool {
			return expectedMessages[i].String() < expectedMessages[j].String()
		})

		sources, err := adapters.PubRepository.ListAddresses(pub.Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub.Host(), pub.Port()),
					Sources: []badger.PubAddressSource{
						{
							Identity: authorIdenRef,
							Messages: expectedMessages,
						},
					},
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Empty(t, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Key()}, pubs)

		expectedMessages := []refs.Message{
			pub2.Message(),
		}

		sources, err := adapters.PubRepository.ListAddresses(pub.Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub.Host(), pub.Port()),
					Sources: []badger.PubAddressSource{
						{
							Identity: authorIdenRef,
							Messages: expectedMessages,
						},
					},
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)
}

func TestPubRepository_Delete_SamePubDifferentAuthorSameAddressDifferentMessage(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	pub := content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt())

	pub1 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		pub,
	)

	pub2 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		pub,
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Key()}, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Key()}, pubs)

		expectedSources := []badger.PubAddressSource{
			{
				Identity: pub1.Who(),
				Messages: []refs.Message{pub1.Message()},
			},
			{
				Identity: pub2.Who(),
				Messages: []refs.Message{pub2.Message()},
			},
		}

		sort.Slice(expectedSources, func(i, j int) bool {
			return expectedSources[i].Identity.String() < expectedSources[j].Identity.String()
		})

		sources, err := adapters.PubRepository.ListAddresses(pub.Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub.Host(), pub.Port()),
					Sources: expectedSources,
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Empty(t, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub.Key()}, pubs)

		expectedSources := []badger.PubAddressSource{
			{
				Identity: pub2.Who(),
				Messages: []refs.Message{pub2.Message()},
			},
		}

		sort.Slice(expectedSources, func(i, j int) bool {
			return expectedSources[i].Identity.String() < expectedSources[j].Identity.String()
		})

		sources, err := adapters.PubRepository.ListAddresses(pub.Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub.Host(), pub.Port()),
					Sources: expectedSources,
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)
}

func TestPubRepository_Delete_SamePubDifferentAuthorDifferentAddressDifferentMessage(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	pubIden := fixtures.SomeRefIdentity()

	pub1 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(pubIden, fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	pub2 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(pubIden, fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pubIden}, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pubIden}, pubs)

		expectedAddresses := []badger.PubAddress{
			{
				Address: fmt.Sprintf("%s:%d", pub1.Content().Host(), pub1.Content().Port()),
				Sources: []badger.PubAddressSource{
					{
						Identity: pub1.Who(),
						Messages: []refs.Message{pub1.Message()},
					},
				},
			},
			{
				Address: fmt.Sprintf("%s:%d", pub2.Content().Host(), pub2.Content().Port()),
				Sources: []badger.PubAddressSource{
					{
						Identity: pub2.Who(),
						Messages: []refs.Message{pub2.Message()},
					},
				},
			},
		}

		sort.Slice(expectedAddresses, func(i, j int) bool {
			return expectedAddresses[i].Address < expectedAddresses[j].Address
		})

		sources, err := adapters.PubRepository.ListAddresses(pubIden)
		require.NoError(t, err)
		require.Equal(t,
			expectedAddresses,
			sources,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Empty(t, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pubIden}, pubs)

		expectedAddresses := []badger.PubAddress{
			{
				Address: fmt.Sprintf("%s:%d", pub2.Content().Host(), pub2.Content().Port()),
				Sources: []badger.PubAddressSource{
					{
						Identity: pub2.Who(),
						Messages: []refs.Message{pub2.Message()},
					},
				},
			},
		}

		sort.Slice(expectedAddresses, func(i, j int) bool {
			return expectedAddresses[i].Address < expectedAddresses[j].Address
		})

		sources, err := adapters.PubRepository.ListAddresses(pubIden)
		require.NoError(t, err)
		require.Equal(t,
			expectedAddresses,
			sources,
		)

		return nil
	})
	require.NoError(t, err)
}

func TestPubRepository_Delete_DifferentPubDifferentAuthorDifferentAddressDifferentMessage(t *testing.T) {
	ts := di.BuildBadgerTestAdapters(t)

	pub1 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	pub2 := feeds.NewPubToSave(
		fixtures.SomeRefIdentity(),
		fixtures.SomeRefMessage(),
		content.MustNewPub(fixtures.SomeRefIdentity(), fixtures.SomeString(), fixtures.SomeNonNegativeInt()),
	)

	err := ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		err := adapters.PubRepository.Put(pub1)
		require.NoError(t, err)

		err = adapters.PubRepository.Put(pub2)
		require.NoError(t, err)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub1.Content().Key()}, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub2.Content().Key()}, pubs)

		sources, err := adapters.PubRepository.ListAddresses(pub1.Content().Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub1.Content().Host(), pub1.Content().Port()),
					Sources: []badger.PubAddressSource{
						{
							Identity: pub1.Who(),
							Messages: []refs.Message{pub1.Message()},
						},
					},
				},
			},
			sources,
		)

		sources, err = adapters.PubRepository.ListAddresses(pub2.Content().Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub2.Content().Host(), pub2.Content().Port()),
					Sources: []badger.PubAddressSource{
						{
							Identity: pub2.Who(),
							Messages: []refs.Message{pub2.Message()},
						},
					},
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.Update(func(adapters badger.TestAdapters) error {
		return adapters.PubRepository.Delete(pub1.Message())
	})
	require.NoError(t, err)

	err = ts.TransactionProvider.View(func(adapters badger.TestAdapters) error {
		pubs, err := adapters.PubRepository.ListPubs(pub1.Message())
		require.NoError(t, err)
		require.Empty(t, pubs)

		pubs, err = adapters.PubRepository.ListPubs(pub2.Message())
		require.NoError(t, err)
		require.Equal(t, []refs.Identity{pub2.Content().Key()}, pubs)

		sources, err := adapters.PubRepository.ListAddresses(pub1.Content().Key())
		require.NoError(t, err)
		require.Empty(t, sources)

		sources, err = adapters.PubRepository.ListAddresses(pub2.Content().Key())
		require.NoError(t, err)
		require.Equal(t,
			[]badger.PubAddress{
				{
					Address: fmt.Sprintf("%s:%d", pub2.Content().Host(), pub2.Content().Port()),
					Sources: []badger.PubAddressSource{
						{
							Identity: pub2.Who(),
							Messages: []refs.Message{pub2.Message()},
						},
					},
				},
			},
			sources,
		)

		return nil
	})
	require.NoError(t, err)
}
