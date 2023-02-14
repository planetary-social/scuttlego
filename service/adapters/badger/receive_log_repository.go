package badger

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/service/adapters/badger/utils"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ReceiveLogRepository struct {
	tx                *badger.Txn
	messageRepository *MessageRepository
}

func NewReceiveLogRepository(tx *badger.Txn, messageRepository *MessageRepository) *ReceiveLogRepository {
	return &ReceiveLogRepository{
		tx:                tx,
		messageRepository: messageRepository,
	}
}

func (r ReceiveLogRepository) Put(id refs.Message) error {
	messagesToSequences, err := r.createMessagesToSequencesBucket(id)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	sequencesToMessages, err := r.createSequencesToMessagesBucket()
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	receiveLogSequence, err := r.getReceiveLogSequence()
	if err != nil {
		return errors.Wrap(err, "could not get the sequence")
	}

	nextSequence, err := receiveLogSequence.Next()
	if err != nil {
		return errors.Wrap(err, "could not get the next sequence")
	}

	receiveLogSeq, err := common.NewReceiveLogSequence(int(nextSequence - 1)) // NextSequence starts with 1 while our log is 0 indexed
	if err != nil {
		return errors.Wrap(err, "failed to create receive log sequence")
	}

	if err := r.put(id, receiveLogSeq, sequencesToMessages, messagesToSequences); err != nil {
		return errors.Wrap(err, "put failed")
	}

	return nil
}

func (r ReceiveLogRepository) PutUnderSpecificSequence(id refs.Message, sequence common.ReceiveLogSequence) error {
	messagesToSequences, err := r.createMessagesToSequencesBucket(id)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	sequencesToMessages, err := r.createSequencesToMessagesBucket()
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	receiveLogSequence, err := r.getReceiveLogSequence()
	if err != nil {
		return errors.Wrap(err, "could not get the sequence")
	}

	if err := r.put(id, sequence, sequencesToMessages, messagesToSequences); err != nil {
		return errors.Wrap(err, "put failed")
	}

	existingSequence, err := receiveLogSequence.Get()
	if err != nil {
		return errors.Wrap(err, "error getting the existing sequence")
	}

	if targetSequence := uint64(sequence.Int()) + 1; existingSequence <= targetSequence { // Automatic insert happens under the set value hence + 1, see Put
		if err := receiveLogSequence.Set(targetSequence); err != nil {
			return errors.Wrap(err, "error setting sequence")
		}
	}

	return nil
}

func (r ReceiveLogRepository) put(id refs.Message, receiveLogSeq common.ReceiveLogSequence, sequencesToMessages utils.Bucket, messagesToSequences utils.Bucket) error {
	seqBytes := r.marshalSequence(receiveLogSeq)
	refBytes := r.marshalRef(id)

	if err := sequencesToMessages.Set(seqBytes, refBytes); err != nil {
		return errors.Wrap(err, "failed to put into sequences to messages bucket")
	}

	if err := messagesToSequences.Set(seqBytes, nil); err != nil {
		return errors.Wrap(err, "failed to put into messages to sequences bucket")
	}

	return nil
}

func (r ReceiveLogRepository) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	bucket, err := r.createSequencesToMessagesBucket()
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	var result []queries.LogMessage

	it := bucket.Iterator()
	defer it.Close()

	for it.Seek(itob(uint64(startSeq.Int()))); it.ValidForBucket(); it.Next() {
		item := it.Item()

		keyInBucket, err := bucket.KeyInBucket(item)
		if err != nil {
			return nil, errors.Wrap(err, "could not determine key in bucket")
		}

		receiveLogSequence, err := r.unmarshalSequence(keyInBucket.Bytes())
		if err != nil {
			return nil, errors.Wrap(err, "could not load the key")
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return nil, errors.Wrap(err, "could not get the value")
		}

		msg, err := r.loadMessage(val)
		if err != nil {
			return nil, errors.Wrap(err, "could not load a message")
		}

		result = append(result, queries.LogMessage{
			Message:  msg,
			Sequence: receiveLogSequence,
		})

		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

func (r ReceiveLogRepository) GetMessage(seq common.ReceiveLogSequence) (message.Message, error) {
	bucket, err := r.createSequencesToMessagesBucket()
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not create a bucket")
	}

	item, err := bucket.Get(r.marshalSequence(seq))
	if err != nil {
		if errors.Is(err, badger.ErrKeyNotFound) {
			return message.Message{}, common.ErrReceiveLogEntryNotFound
		}
		return message.Message{}, errors.Wrap(err, "error getting the item")
	}

	value, err := item.ValueCopy(nil)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "getting value error")
	}

	msg, err := r.loadMessage(value)
	if err != nil {
		return message.Message{}, errors.New("could not load a message")
	}

	return msg, nil
}

func (r ReceiveLogRepository) GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error) {
	bucket, err := r.createMessagesToSequencesBucket(ref)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	var sequences []common.ReceiveLogSequence

	if err := bucket.ForEach(func(item utils.Item) error {
		keyInBucket, err := bucket.KeyInBucket(item)
		if err != nil {
			return errors.Wrap(err, "unable to determine key in bucket")
		}

		sequence, err := r.unmarshalSequence(keyInBucket.Bytes())
		if err != nil {
			return errors.Wrap(err, "error unmarshaling sequence")
		}

		sequences = append(sequences, sequence)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "foreach error")
	}

	if len(sequences) == 0 {
		return nil, common.ErrReceiveLogEntryNotFound
	}

	return sequences, nil
}

func (r ReceiveLogRepository) loadMessage(value []byte) (message.Message, error) {
	id, err := r.unmarshalRef(value)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not create a message ref")
	}

	msg, err := r.messageRepository.Get(id)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not get the message")
	}

	return msg, nil
}

func (r ReceiveLogRepository) getReceiveLogSequence() (utils.Sequence, error) {
	bucket, err := r.createMetaBucket()
	if err != nil {
		return utils.Sequence{}, errors.Wrap(err, "error getting bucket")
	}
	return utils.NewSequence(
		bucket,
		utils.MustNewKeyComponent([]byte("sequence")),
	)
}

func (r ReceiveLogRepository) createMetaBucket() (utils.Bucket, error) {
	return utils.NewBucket(r.tx, r.metaBucketPath())
}

func (r ReceiveLogRepository) metaBucketPath() utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte("receive_log")),
		utils.MustNewKeyComponent([]byte("meta")),
	)
}

func (r ReceiveLogRepository) createSequencesToMessagesBucket() (utils.Bucket, error) {
	return utils.NewBucket(r.tx, r.sequencesToMessagesBucketPath())
}

func (r ReceiveLogRepository) sequencesToMessagesBucketPath() utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte("receive_log")),
		utils.MustNewKeyComponent([]byte("sequences_to_messages")),
	)
}

func (r ReceiveLogRepository) createMessagesToSequencesBucket(msgRef refs.Message) (utils.Bucket, error) {
	return utils.NewBucket(r.tx, r.messagesToSequencesBucketPath(msgRef))
}

func (r ReceiveLogRepository) messagesToSequencesBucketPath(msgRef refs.Message) utils.Key {
	return utils.MustNewKey(
		utils.MustNewKeyComponent([]byte("receive_log")),
		utils.MustNewKeyComponent([]byte("messages_to_sequences")),
		utils.MustNewKeyComponent([]byte(msgRef.String())),
	)
}

func (r ReceiveLogRepository) marshalSequence(seq common.ReceiveLogSequence) []byte {
	return itob(uint64(seq.Int()))
}

func (r ReceiveLogRepository) unmarshalSequence(key []byte) (common.ReceiveLogSequence, error) {
	return common.NewReceiveLogSequence(int(btoi(key)))
}

func (r ReceiveLogRepository) marshalRef(id refs.Message) []byte {
	return []byte(id.String())
}

func (r ReceiveLogRepository) unmarshalRef(value []byte) (refs.Message, error) {
	return refs.NewMessage(string(value))
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}
