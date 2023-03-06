package di

import (
	"github.com/google/wire"
	"github.com/planetary-social/scuttlego/service/app/commands"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
)

var contentSet = wire.NewSet(
	content.NewParser,
	wire.Bind(new(formats.ContentParser), new(*content.Parser)),
	wire.Bind(new(commands.ContentParser), new(*content.Parser)),

	blobs.NewScanner,
	wire.Bind(new(content.BlobScanner), new(*blobs.Scanner)),
)
