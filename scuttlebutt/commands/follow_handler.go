package commands

import (
	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/logging"
)

type Follow struct {
}

type FollowHandler struct {
	logger logging.Logger
}

func NewFollowHandler(logger logging.Logger) *FollowHandler {
	return &FollowHandler{
		logger: logger.New("follow_handler"),
	}
}

func (h *FollowHandler) Handle(cmd Follow) error {
	return errors.New("not implemented")
}
