package blobs

import (
	"bytes"
	"encoding/base64"
	"strings"
	"unicode/utf8"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
)

const (
	standardBase64Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

type Scanner struct {
}

func NewScanner() *Scanner {
	return &Scanner{}
}

func (s *Scanner) Scan(content message.RawContent) ([]refs.Blob, error) {
	if !bytes.Contains(content.Bytes(), []byte(refs.BlobPrefix)) {
		return nil, nil
	}

	l := newLexer(content.Bytes())

	if err := l.lex(); err != nil {
		return nil, errors.Wrap(err, "error lexing the message content")
	}

	results := make(map[string]refs.Blob)

	for _, item := range l.items {
		switch item.Type {
		case itemBlobRef:
			justBase64 := strings.TrimSuffix(strings.TrimPrefix(item.Value, refs.BlobPrefix), refs.BlobSuffix)

			b, err := base64.StdEncoding.DecodeString(justBase64)
			if err != nil {
				continue
			}

			if l := len(b); l != refs.BlobHashLength {
				continue
			}

			blobRef, err := refs.NewBlob(item.Value)
			if err != nil {
				return nil, errors.Wrapf(err, "error creating a blob ref from '%s'", item.Value)
			}

			results[blobRef.String()] = blobRef
		default:
			return nil, errors.New("unknown item type")
		}
	}

	return s.toSlice(results), nil
}

func (s *Scanner) toSlice(mapResults map[string]refs.Blob) []refs.Blob {
	var result []refs.Blob
	for _, ref := range mapResults {
		result = append(result, ref)
	}
	return result
}

type itemType string

const (
	itemBlobRef itemType = "blob_ref"
)

type item struct {
	Type  itemType
	Value string
}

type stateFunc func(l *lexer) (stateFunc, error)

type lexer struct {
	b       []byte
	pos     int
	prevPos int
	current item
	items   []item
}

func newLexer(b []byte) *lexer {
	return &lexer{
		b: b,
	}
}

func (l *lexer) lex() error {
	var state stateFunc = lexUnknown

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
	r, size := utf8.DecodeRune(l.b[l.pos:])
	if r == utf8.RuneError {
		return 0, false
	}

	l.prevPos = l.pos
	l.pos += size
	l.current.Value += string(r)
	return r, true
}

func (l *lexer) backtrack() {
	length := l.pos - l.prevPos
	l.pos = l.prevPos
	l.current.Value = l.current.Value[:len(l.current.Value)-length]
}

func (l *lexer) emit() {
	l.items = append(l.items, l.current)
}

func lexUnknown(l *lexer) (stateFunc, error) {
	firstPrefixRune, _ := utf8.DecodeRuneInString(refs.BlobPrefix)
	if firstPrefixRune == utf8.RuneError {
		return nil, errors.New("error decoding first rune in prefix")
	}

	for {
		r, ok := l.next()
		if !ok {
			return nil, nil
		}

		switch r {
		case firstPrefixRune:
			l.backtrack()
			return lexBlobRefPrefix, nil
		default:
			continue
		}
	}
}

func lexBlobRefPrefix(l *lexer) (stateFunc, error) {
	l.current = item{Type: itemBlobRef}

	for _, expectedR := range refs.BlobPrefix {
		r, ok := l.next()
		if !ok {
			return nil, nil
		}

		if r != expectedR {
			l.backtrack()
			return lexUnknown, nil
		}
	}

	return lexBlobRefBase64, nil
}

func lexBlobRefBase64(l *lexer) (stateFunc, error) {
	for {
		r, ok := l.next()
		if !ok {
			return nil, nil
		}

		if !strings.ContainsRune(standardBase64Alphabet, r) && r != base64.StdPadding {
			l.backtrack()
			return lexBlobRefSuffix, nil
		}
	}
}

func lexBlobRefSuffix(l *lexer) (stateFunc, error) {
	for _, expectedR := range refs.BlobSuffix {
		r, ok := l.next()
		if !ok {
			return nil, nil
		}

		if r != expectedR {
			l.backtrack()
			return lexUnknown, nil
		}
	}

	l.emit()
	return lexUnknown, nil
}
