package mocks

import (
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
)

type ContentParser struct {
}

func NewContentParser() *ContentParser {
	return &ContentParser{}
}

func (c ContentParser) Parse(raw message.RawContent) (message.Content, error) {
	return message.NewContent(raw, nil, nil)
}
