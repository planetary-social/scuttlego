package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/adapters/bolt"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/feeds"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/transport"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
)

//nolint:deadcode,varcheck
var formatsSet = wire.NewSet(
	newFormats,

	formats.NewScuttlebutt,

	transport.NewMarshaler,
	wire.Bind(new(formats.Marshaler), new(*transport.Marshaler)),

	transport.DefaultMappings,

	formats.NewRawMessageIdentifier,
	wire.Bind(new(commands.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
	wire.Bind(new(bolt.RawMessageIdentifier), new(*formats.RawMessageIdentifier)),
)

func newFormats(
	s *formats.Scuttlebutt,
) []feeds.FeedFormat {
	return []feeds.FeedFormat{
		s,
	}
}
