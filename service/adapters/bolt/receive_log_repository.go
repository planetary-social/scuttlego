package bolt

import (
	"encoding/binary"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/adapters/bolt/utils"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"go.etcd.io/bbolt"
)

type ReceiveLogRepository struct {
	tx                *bbolt.Tx
	messageRepository *MessageRepository
}

func NewReceiveLogRepository(tx *bbolt.Tx, messageRepository *MessageRepository) *ReceiveLogRepository {
	return &ReceiveLogRepository{
		tx:                tx,
		messageRepository: messageRepository,
	}
}

func (r ReceiveLogRepository) Put(id refs.Message) error {
	messagesToSequences, err := r.createMessagesToSequencesBucket(r.tx, id)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	sequencesToMessages, err := r.createSequencesToMessagesBucket(r.tx)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	seq, err := sequencesToMessages.NextSequence()
	if err != nil {
		return errors.Wrap(err, "could not get the next sequence")
	}

	receiveLogSeq, err := common.NewReceiveLogSequence(int(seq - 1)) // NextSequence starts with 1 while our log is 0 indexed
	if err != nil {
		return errors.Wrap(err, "failed to create receive log sequence")
	}

	if err := r.put(id, receiveLogSeq, sequencesToMessages, messagesToSequences); err != nil {
		return errors.Wrap(err, "put failed")
	}

	return nil
}

func (r ReceiveLogRepository) PutUnderSpecificSequence(id refs.Message, sequence common.ReceiveLogSequence) error {
	messagesToSequences, err := r.createMessagesToSequencesBucket(r.tx, id)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	sequencesToMessages, err := r.createSequencesToMessagesBucket(r.tx)
	if err != nil {
		return errors.Wrap(err, "could not create a bucket")
	}

	if err := r.put(id, sequence, sequencesToMessages, messagesToSequences); err != nil {
		return errors.Wrap(err, "put failed")
	}

	// todo advance counter I think

	return nil
}

func (r ReceiveLogRepository) put(id refs.Message, receiveLogSeq common.ReceiveLogSequence, sequencesToMessages *bbolt.Bucket, messagesToSequences *bbolt.Bucket) error {
	seqBytes := r.marshalSequence(receiveLogSeq)
	refBytes := r.marshalRef(id)

	if err := sequencesToMessages.Put(seqBytes, refBytes); err != nil {
		return errors.Wrap(err, "failed to put into sequences to messages bucket")
	}

	if err := messagesToSequences.Put(seqBytes, nil); err != nil {
		return errors.Wrap(err, "failed to put into messages to sequences bucket")
	}

	return nil
}

func (r ReceiveLogRepository) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	if limit <= 0 {
		return nil, errors.New("limit must be positive")
	}

	bucket, err := r.getSequencesToMessagesBucket(r.tx)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return nil, nil
	}

	var result []queries.LogMessage

	c := bucket.Cursor()
	for key, value := c.Seek(itob(uint64(startSeq.Int()))); key != nil; key, value = c.Next() {
		receiveLogSequence, err := r.unmarshalSequence(key)
		if err != nil {
			return nil, errors.New("could not load the key")
		}

		msg, err := r.loadMessage(value)
		if err != nil {
			return nil, errors.New("could not load a message")
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
	bucket, err := r.getSequencesToMessagesBucket(r.tx)
	if err != nil {
		return message.Message{}, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return message.Message{}, common.ErrReceiveLogEntryNotFound
	}

	value := bucket.Get(r.marshalSequence(seq))
	if value == nil {
		return message.Message{}, common.ErrReceiveLogEntryNotFound
	}

	return r.loadMessage(value)
}

func (r ReceiveLogRepository) GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error) {
	bucket, err := r.getMessagesToSequencesBucket(r.tx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return nil, common.ErrReceiveLogEntryNotFound
	}

	var sequences []common.ReceiveLogSequence

	if err := bucket.ForEach(func(k, v []byte) error {
		sequence, err := r.unmarshalSequence(k)
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
		return message.Message{}, errors.New("could not create a message ref")
	}

	msg, err := r.messageRepository.Get(id)
	if err != nil {
		return message.Message{}, errors.New("could not get the message")
	}

	return msg, nil
}

func (r ReceiveLogRepository) createSequencesToMessagesBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return utils.CreateBucket(tx, r.sequencesToMessagesBucketPath())
}

func (r ReceiveLogRepository) getSequencesToMessagesBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return utils.GetBucket(tx, r.sequencesToMessagesBucketPath())
}

func (r ReceiveLogRepository) sequencesToMessagesBucketPath() []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName("receive_log"),
		utils.BucketName("sequences_to_messages"),
	}
}

func (r ReceiveLogRepository) createMessagesToSequencesBucket(tx *bbolt.Tx, msgRef refs.Message) (*bbolt.Bucket, error) {
	return utils.CreateBucket(tx, r.messagesToSequencesBucketPath(msgRef))
}

func (r ReceiveLogRepository) getMessagesToSequencesBucket(tx *bbolt.Tx, msgRef refs.Message) (*bbolt.Bucket, error) {
	return utils.GetBucket(tx, r.messagesToSequencesBucketPath(msgRef))
}

func (r ReceiveLogRepository) messagesToSequencesBucketPath(msgRef refs.Message) []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName("receive_log"),
		utils.BucketName("messages_to_sequences"),
		utils.BucketName(msgRef.String()),
	}
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

type ReadReceiveLogRepository struct {
	db      *bbolt.DB
	factory TxRepositoriesFactory
}

func NewReadReceiveLogRepository(db *bbolt.DB, factory TxRepositoriesFactory) *ReadReceiveLogRepository {
	return &ReadReceiveLogRepository{
		db:      db,
		factory: factory,
	}
}

func (r ReadReceiveLogRepository) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	var result []queries.LogMessage

	if err := r.db.View(func(tx *bbolt.Tx) error {
		r, err := r.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		msgs, err := r.ReceiveLog.List(startSeq, limit)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = msgs
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (r ReadReceiveLogRepository) GetMessage(seq common.ReceiveLogSequence) (message.Message, error) {
	var result message.Message

	if err := r.db.View(func(tx *bbolt.Tx) error {
		r, err := r.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		msg, err := r.ReceiveLog.GetMessage(seq)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = msg
		return nil
	}); err != nil {
		return message.Message{}, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func (r ReadReceiveLogRepository) GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error) {
	var result []common.ReceiveLogSequence

	if err := r.db.View(func(tx *bbolt.Tx) error {
		r, err := r.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		seq, err := r.ReceiveLog.GetSequences(ref)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = seq
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	return result, nil
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}

func btoi(v []byte) uint64 {
	return binary.BigEndian.Uint64(v)
}
