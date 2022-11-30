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

var errReceiveLogEntryNotFound = errors.New("receive log entry not found")

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
	messagesToSequences, err := r.createMessagesToSequencesBucket(r.tx)
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

	seqBytes := r.marshalSequence(receiveLogSeq)
	refBytes := r.marshalRef(id)

	if err := sequencesToMessages.Put(seqBytes, refBytes); err != nil {
		return errors.Wrap(err, "failed to put into sequences to messages bucket")
	}

	if err := messagesToSequences.Put(refBytes, seqBytes); err != nil {
		return errors.Wrap(err, "failed to put into messages to sequences bucket")
	}

	return nil
}

func (r ReceiveLogRepository) PutUnderSpecificSequence(id refs.Message, sequence common.ReceiveLogSequence) error {
	// todo
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
		return message.Message{}, errReceiveLogEntryNotFound
	}

	value := bucket.Get(r.marshalSequence(seq))
	if value == nil {
		return message.Message{}, errReceiveLogEntryNotFound
	}

	return r.loadMessage(value)
}

func (r ReceiveLogRepository) GetSequence(ref refs.Message) (common.ReceiveLogSequence, error) {
	bucket, err := r.getMessagesToSequencesBucket(r.tx)
	if err != nil {
		return common.ReceiveLogSequence{}, errors.Wrap(err, "could not create a bucket")
	}

	if bucket == nil {
		return common.ReceiveLogSequence{}, errReceiveLogEntryNotFound
	}

	value := bucket.Get(r.marshalRef(ref))
	if value == nil {
		return common.ReceiveLogSequence{}, errReceiveLogEntryNotFound
	}

	return r.unmarshalSequence(value)
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

func (r ReceiveLogRepository) createMessagesToSequencesBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return utils.CreateBucket(tx, r.messagesToSequencesBucketPath())
}

func (r ReceiveLogRepository) getMessagesToSequencesBucket(tx *bbolt.Tx) (*bbolt.Bucket, error) {
	return utils.GetBucket(tx, r.messagesToSequencesBucketPath())
}

func (r ReceiveLogRepository) messagesToSequencesBucketPath() []utils.BucketName {
	return []utils.BucketName{
		utils.BucketName("receive_log"),
		utils.BucketName("messages_to_sequences"),
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

func (r ReadReceiveLogRepository) GetSequence(ref refs.Message) (common.ReceiveLogSequence, error) {
	var result common.ReceiveLogSequence

	if err := r.db.View(func(tx *bbolt.Tx) error {
		r, err := r.factory(tx)
		if err != nil {
			return errors.Wrap(err, "could not call the factory")
		}

		seq, err := r.ReceiveLog.GetSequence(ref)
		if err != nil {
			return errors.Wrap(err, "failed to call the repository")
		}

		result = seq
		return nil
	}); err != nil {
		return common.ReceiveLogSequence{}, errors.Wrap(err, "transaction failed")
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
