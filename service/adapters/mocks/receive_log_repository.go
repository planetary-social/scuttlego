package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ReceiveLogRepositoryPutUnderSpecificSequenceCall struct {
	Id       refs.Message
	Sequence common.ReceiveLogSequence
}

type ReceiveLogRepositoryMock struct {
	GetMessageCalls               []common.ReceiveLogSequence
	PutUnderSpecificSequenceCalls []ReceiveLogRepositoryPutUnderSpecificSequenceCall

	sequenceToMessages  map[common.ReceiveLogSequence]message.Message
	messagesToSequences map[string]common.ReceiveLogSequence
}

func NewReceiveLogRepositoryMock() *ReceiveLogRepositoryMock {
	return &ReceiveLogRepositoryMock{
		sequenceToMessages:  map[common.ReceiveLogSequence]message.Message{},
		messagesToSequences: map[string]common.ReceiveLogSequence{},
	}
}

func (r ReceiveLogRepositoryMock) MockMessage(seq common.ReceiveLogSequence, msg message.Message) {
	r.sequenceToMessages[seq] = msg
	r.messagesToSequences[msg.Id().String()] = seq
}

func (r ReceiveLogRepositoryMock) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	return nil, nil
}

func (r *ReceiveLogRepositoryMock) GetMessage(seq common.ReceiveLogSequence) (message.Message, error) {
	r.GetMessageCalls = append(r.GetMessageCalls, seq)

	v, ok := r.sequenceToMessages[seq]
	if !ok {
		return message.Message{}, errors.New("not found")
	}
	return v, nil
}

func (r ReceiveLogRepositoryMock) GetSequence(ref refs.Message) (common.ReceiveLogSequence, error) {
	v, ok := r.messagesToSequences[ref.String()]
	if !ok {
		return common.ReceiveLogSequence{}, errors.New("not found")
	}
	return v, nil
}

func (r *ReceiveLogRepositoryMock) PutUnderSpecificSequence(id refs.Message, sequence common.ReceiveLogSequence) error {
	r.PutUnderSpecificSequenceCalls = append(r.PutUnderSpecificSequenceCalls, ReceiveLogRepositoryPutUnderSpecificSequenceCall{
		Id:       id,
		Sequence: sequence,
	})
	return nil
}
