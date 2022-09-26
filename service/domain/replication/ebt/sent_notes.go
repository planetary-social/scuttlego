package ebt

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/replication"
)

type SentNotes struct {
	prevNotes map[string]messages.EbtReplicateNote
}

func NewSentNotes() *SentNotes {
	return &SentNotes{
		make(map[string]messages.EbtReplicateNote),
	}
}

func (w *SentNotes) Update(contacts []replication.Contact) (messages.EbtReplicateNotes, error) {
	var notesToSend []messages.EbtReplicateNote

	missing := w.getMissing(contacts)

	for _, contact := range contacts {
		note, err := w.contactToNote(contact)
		if err != nil {
			return messages.EbtReplicateNotes{}, errors.Wrap(err, "could not create a note")
		}

		if w.shouldSend(note) {
			notesToSend = append(notesToSend, note)
		}

		w.prevNotes[note.Ref().String()] = note
	}

	for refString, note := range missing {
		note, err := w.cancellationNote(note.Ref())
		if err != nil {
			return messages.EbtReplicateNotes{}, errors.Wrap(err, "could not create a note")
		}

		delete(w.prevNotes, refString)
		notesToSend = append(notesToSend, note)
	}

	return messages.NewEbtReplicateNotes(notesToSend)
}

func (w *SentNotes) getMissing(newContacts []replication.Contact) map[string]messages.EbtReplicateNote {
	missing := make(map[string]messages.EbtReplicateNote)

	new := make(map[string]struct{})
	for _, newContact := range newContacts {
		new[newContact.Who.String()] = struct{}{}
	}

	for refString, note := range w.prevNotes {
		if _, ok := new[refString]; !ok {
			missing[refString] = note
		}
	}

	return missing
}

func (w *SentNotes) feedStateAsInt(feedState replication.FeedState) int {
	sequence, ok := feedState.Sequence()
	if ok {
		return sequence.Int()
	}
	return 0
}

func (w *SentNotes) cancellationNote(ref refs.Feed) (messages.EbtReplicateNote, error) {
	return messages.NewEbtReplicateNote(ref, false, false, -1)
}

func (w *SentNotes) contactToNote(contact replication.Contact) (messages.EbtReplicateNote, error) {
	seq := w.feedStateAsInt(contact.FeedState)
	return messages.NewEbtReplicateNote(contact.Who, true, true, seq)
}

func (w *SentNotes) shouldSend(note messages.EbtReplicateNote) bool {
	prevNote, ok := w.prevNotes[note.Ref().String()]
	if ok {
		if prevNote.Equal(note) {
			return false
		}

	}

	return true
}
