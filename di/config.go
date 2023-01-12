package di

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/transport/boxstream"
)

type Config struct {
	// DataDirectory specifies where the primary database and other data
	// will be stored.
	DataDirectory string

	// GoSSBDataDirectory specifies the location of the old data directory used
	// by go-ssb. This is needed for data migrations.
	GoSSBDataDirectory string

	// ListenAddress for the TCP listener in the format accepted by the
	// standard library.
	// Optional, defaults to ":8008".
	ListenAddress string

	// Setting NetworkKey is mainly useful for test networks.
	// Optional, defaults to boxstream.NewDefaultNetworkKey().
	NetworkKey boxstream.NetworkKey

	// Setting MessageHMAC is mainly useful for test networks.
	// Optional, defaults to formats.NewDefaultMessageHMAC().
	MessageHMAC formats.MessageHMAC

	// Logger is the logger used for logging by this library. It is most
	// likely useful to configure at least to log errors.
	// Optional, defaults to logging.NewDevNullLogger().
	Logger logging.Logger

	// PeerManagerConfig specifies the config for the peer manager which is responsible for establishing new
	// connections and managing existing connections.
	PeerManagerConfig domain.PeerManagerConfig

	// ModifyBadgerOptions allows you to specify a function allowing you to modify certain Badger options.
	// Optional, this value is ignored if not set.
	ModifyBadgerOptions func(options BadgerOptions)
}

func (c *Config) SetDefaults() {
	if c.ListenAddress == "" {
		c.ListenAddress = ":8008"
	}

	if c.NetworkKey.IsZero() {
		c.NetworkKey = boxstream.NewDefaultNetworkKey()
	}

	if c.MessageHMAC.IsZero() {
		c.MessageHMAC = formats.NewDefaultMessageHMAC()
	}

	if c.Logger == nil {
		c.Logger = logging.NewDevNullLogger()
	}
}

type BadgerOptions interface {
	SetNumGoroutines(val int)
	SetNumCompactors(val int)
	SetCompression(val options.CompressionType)
	SetLogger(val badger.Logger)
	SetValueLogFileSize(val int64)
}

type BadgerOptionsAdapter struct {
	options *badger.Options
}

func NewBadgerOptionsAdapter(options *badger.Options) BadgerOptionsAdapter {
	return BadgerOptionsAdapter{options: options}
}

func (b BadgerOptionsAdapter) SetNumGoroutines(val int) {
	b.options.NumGoroutines = val
}

func (b BadgerOptionsAdapter) SetNumCompactors(val int) {
	b.options.NumCompactors = val
}

func (b BadgerOptionsAdapter) SetCompression(val options.CompressionType) {
	b.options.Compression = val
}

func (b BadgerOptionsAdapter) SetLogger(val badger.Logger) {
	b.options.Logger = val
}

func (b BadgerOptionsAdapter) SetValueLogFileSize(val int64) {
	b.options.ValueLogFileSize = val
}
