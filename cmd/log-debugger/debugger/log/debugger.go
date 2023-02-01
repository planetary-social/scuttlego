package log

import (
	"bufio"
	"os"

	"github.com/boreq/errors"
)

type Log []Entry
type Entry map[string]string

const maxTokenLength = 1 * 1024 * 1024

func LoadLog(s string) (Log, error) {
	file, err := os.Open(s)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open the file")
	}

	var log Log

	scanner := bufio.NewScanner(file)
	scanner.Buffer(nil, maxTokenLength)

	for scanner.Scan() {
		line, err := loadLine(scanner.Bytes())
		if err != nil {
			return nil, errors.Wrap(err, "failed to load a line")
		}

		log = append(log, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanner error")
	}

	return log, nil
}

func loadLine(b []byte) (Entry, error) {
	l := newLexer(b)
	if err := l.lex(); err != nil {
		return nil, errors.Wrap(err, "lex failed")
	}

	if len(l.items)%2 != 0 {
		return nil, errors.New("malformed line")
	}

	result := make(Entry)

	for i := 0; i < len(l.items); i += 2 {
		label := l.items[i]
		value := l.items[i+1]

		if label.Type != labelItem {
			return nil, errors.New("expected a label")
		}

		if value.Type != valueItem {
			return nil, errors.New("expected a value")
		}

		result[label.Value] = value.Value
	}

	return result, nil
}
