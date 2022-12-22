package badger_test

//func TestReadMessageRepository_Count(t *testing.T) {
//	ts := di.BuildBadgerTestAdapters(t)
//
//
//		msg := fixtures.SomeMessage(fixtures.SomeSequence(), fixtures.SomeRefFeed())
//
//		a, err := di.BuildTestAdapters(db)
//		require.NoError(t, err)
//
//		count, err := a.MessageRepository.Count()
//		require.NoError(t, err)
//		require.Equal(t, 0, count)
//
//		err = db.Update(func(tx *bbolt.Tx) error {
//			adapters, err := di.BuildTxTestAdapters(tx)
//			require.NoError(t, err)
//
//			return adapters.MessageRepository.Put(msg)
//		})
//		require.NoError(t, err)
//
//		count, err = a.MessageRepository.Count()
//		require.NoError(t, err)
//		require.Equal(t, 1, count)
//	}
