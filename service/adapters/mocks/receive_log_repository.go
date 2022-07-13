package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ReceiveLogRepositoryMock struct {
	GetMessageCalls []queries.ReceiveLogSequence

	sequenceToMessages  map[queries.ReceiveLogSequence]message.Message
	messagesToSequences map[string]queries.ReceiveLogSequence
}

func NewReceiveLogRepositoryMock() *ReceiveLogRepositoryMock {
	return &ReceiveLogRepositoryMock{
		sequenceToMessages:  map[queries.ReceiveLogSequence]message.Message{},
		messagesToSequences: map[string]queries.ReceiveLogSequence{},
	}
}

func (r ReceiveLogRepositoryMock) MockMessage(seq queries.ReceiveLogSequence, msg message.Message) {
	r.sequenceToMessages[seq] = msg
	r.messagesToSequences[msg.Id().String()] = seq
}

func (r ReceiveLogRepositoryMock) List(startSeq queries.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	return nil, nil
}

func (r *ReceiveLogRepositoryMock) GetMessage(seq queries.ReceiveLogSequence) (message.Message, error) {
	r.GetMessageCalls = append(r.GetMessageCalls, seq)

	v, ok := r.sequenceToMessages[seq]
	if !ok {
		return message.Message{}, errors.New("not found")
	}
	return v, nil
}

func (r ReceiveLogRepositoryMock) GetSequence(ref refs.Message) (queries.ReceiveLogSequence, error) {
	v, ok := r.messagesToSequences[ref.String()]
	if !ok {
		return queries.ReceiveLogSequence{}, errors.New("not found")
	}
	return v, nil
}
