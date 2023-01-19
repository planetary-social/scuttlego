package debugger

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/boreq/errors"
)

type itemType string

const (
	labelItem itemType = "label"
	valueItem itemType = "value"
)

type item struct {
	Type  itemType
	Value string
}

type stateFunc func(l *lexer) (stateFunc, error)

type lexer struct {
	s       string
	pos     int
	prevPos int
	current item
	items   []item
}

func newLexer(s string) *lexer {
	return &lexer{
		s: s,
	}
}

func (l *lexer) lex() error {
	var state stateFunc = lexLabel

	for {
		if state == nil {
			return nil
		}

		var err error
		state, err = state(l)
		if err != nil {
			return errors.Wrap(err, "state func returned an error")
		}
	}
}

func (l *lexer) next() (rune, bool) {
	r, size := utf8.DecodeRune([]byte(l.s[l.pos:]))

	if r == utf8.RuneError {
		return 0, false
	}

	l.prevPos = l.pos
	l.pos += size
	l.current.Value += string(r)
	return r, true
}

func (l *lexer) previous() rune {
	r, _ := utf8.DecodeLastRune([]byte(l.s[:l.prevPos]))
	return r
}

func (l *lexer) backtrack() {
	length := l.pos - l.prevPos
	l.pos = l.prevPos
	l.current.Value = l.current.Value[:len(l.current.Value)-length]
}

// drop last element form current value
func (l *lexer) skip() {
	_, length := utf8.DecodeLastRune([]byte(l.current.Value))
	l.current.Value = l.current.Value[:len(l.current.Value)-length]
}

func (l *lexer) peek() rune {
	next, _ := l.next()
	l.backtrack()
	return next
}

func (l *lexer) emit() {
	l.items = append(l.items, l.current)
}

func lexLabel(l *lexer) (stateFunc, error) {
	l.current = item{Type: labelItem}

	for {
		r, ok := l.next()
		if !ok {
			return nil, errors.New("next returned false")
		}

		if r == '=' {
			l.skip()
			l.emit()
			return lexValue, nil
		}

		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsPunct(r) {
			return nil, fmt.Errorf("expected end of label but got '%s' (%s)", string(r), l.current)
		}
	}
}

func lexValue(l *lexer) (stateFunc, error) {
	l.current = item{Type: valueItem}

	r, ok := l.next()
	if !ok {
		return nil, errors.New("next returned false")
	}

	if r == '"' {
		l.skip()
		return lexInsideQuotes, nil
	}

	if unicode.IsSpace(r) {
		l.skip()
		l.emit()
		return lexLabel, nil
	}

	return lexWithoutQuotes, nil
}

func lexInsideQuotes(l *lexer) (stateFunc, error) {
	for {
		r, ok := l.next()
		if !ok {
			return nil, errors.New("next returned false")
		}

		if r == '\\' && l.peek() == '"' {
			l.skip()
			continue
		}

		if r == '"' && l.previous() != '\\' {
			l.skip()
			l.emit()
			return lexEndOfValue, nil
		}
	}
}

func lexWithoutQuotes(l *lexer) (stateFunc, error) {
	for {
		r, ok := l.next()
		if !ok {
			l.emit()
			return nil, nil
		}

		if unicode.IsSpace(r) {
			l.backtrack()
			l.emit()
			return lexEndOfValue, nil
		}
	}
}

func lexEndOfValue(l *lexer) (stateFunc, error) {
	for {
		r, ok := l.next()
		if !ok {
			return nil, nil
		}

		if unicode.IsSpace(r) {
			l.skip()
		}

		return lexLabel, nil
	}
}
