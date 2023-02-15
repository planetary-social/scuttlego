package mocks

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/app/queries"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

type ReceiveLogRepositoryMock struct {
	GetMessageCalls               []common.ReceiveLogSequence
	GetSequencesCalls             []refs.Message
	PutUnderSpecificSequenceCalls []ReceiveLogRepositoryPutUnderSpecificSequenceCall
	ReserveSequencesUpToCalls     []ReceiveLogRepositoryReserveSequencesUpToCall

	sequenceToMessages  map[common.ReceiveLogSequence]message.Message
	messagesToSequences map[string][]common.ReceiveLogSequence
}

func NewReceiveLogRepositoryMock() *ReceiveLogRepositoryMock {
	return &ReceiveLogRepositoryMock{
		sequenceToMessages:  map[common.ReceiveLogSequence]message.Message{},
		messagesToSequences: map[string][]common.ReceiveLogSequence{},
	}
}

func (r ReceiveLogRepositoryMock) MockMessage(seq common.ReceiveLogSequence, msg message.Message) {
	r.sequenceToMessages[seq] = msg
	r.messagesToSequences[msg.Id().String()] = append(r.messagesToSequences[msg.Id().String()], seq)
}

func (r ReceiveLogRepositoryMock) List(startSeq common.ReceiveLogSequence, limit int) ([]queries.LogMessage, error) {
	return nil, nil
}

func (r *ReceiveLogRepositoryMock) GetMessage(seq common.ReceiveLogSequence) (message.Message, error) {
	r.GetMessageCalls = append(r.GetMessageCalls, seq)

	v, ok := r.sequenceToMessages[seq]
	if !ok {
		return message.Message{}, common.ErrReceiveLogEntryNotFound
	}
	return v, nil
}

func (r *ReceiveLogRepositoryMock) GetSequences(ref refs.Message) ([]common.ReceiveLogSequence, error) {
	r.GetSequencesCalls = append(r.GetSequencesCalls, ref)

	sequences, ok := r.messagesToSequences[ref.String()]
	if !ok {
		return nil, errors.New("not found")
	}
	return sequences, nil
}

func (r *ReceiveLogRepositoryMock) PutUnderSpecificSequence(id refs.Message, sequence common.ReceiveLogSequence) error {
	r.PutUnderSpecificSequenceCalls = append(r.PutUnderSpecificSequenceCalls, ReceiveLogRepositoryPutUnderSpecificSequenceCall{
		Id:       id,
		Sequence: sequence,
	})
	return nil
}

func (r *ReceiveLogRepositoryMock) ReserveSequencesUpTo(sequence common.ReceiveLogSequence) error {
	r.ReserveSequencesUpToCalls = append(r.ReserveSequencesUpToCalls, ReceiveLogRepositoryReserveSequencesUpToCall{
		Sequence: sequence,
	})
	return nil
}

type ReceiveLogRepositoryPutUnderSpecificSequenceCall struct {
	Id       refs.Message
	Sequence common.ReceiveLogSequence
}

type ReceiveLogRepositoryReserveSequencesUpToCall struct {
	Sequence common.ReceiveLogSequence
}
