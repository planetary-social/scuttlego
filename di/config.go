package di

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/dgraph-io/badger/v3/options"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/domain"
	"github.com/planetary-social/scuttlego/service/domain/feeds/formats"
	"github.com/planetary-social/scuttlego/service/domain/graph"
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

	// LoggingSystem is the logger used for logging by this library.
	// Optional, defaults to logging.NewDevNullLoggingSystem().
	LoggingSystem logging.LoggingSystem

	// PeerManagerConfig specifies the config for the peer manager which is responsible for establishing new
	// connections and managing existing connections.
	PeerManagerConfig domain.PeerManagerConfig

	// Hops specifies how far away the feeds which are automatically replicated
	// based on contact messages can be in the social graph.
	// Optional, defaults to 2 (followees of your followees).
	Hops *graph.Hops

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

	if c.LoggingSystem == nil {
		c.LoggingSystem = logging.NewDevNullLoggingSystem()
	}

	if c.Hops == nil {
		c.Hops = internal.Ptr(graph.MustNewHops(2))
	}
}

type BadgerOptions interface {
	SetNumGoroutines(val int)
	SetNumCompactors(val int)
	SetCompression(val options.CompressionType)
	SetLogger(val badger.Logger)
	SetValueLogFileSize(val int64)
	SetBlockCacheSize(val int64)
	SetIndexCacheSize(val int64)
	SetSyncWrites(val bool)
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

func (b BadgerOptionsAdapter) SetBlockCacheSize(val int64) {
	b.options.BlockCacheSize = val
}

func (b BadgerOptionsAdapter) SetIndexCacheSize(val int64) {
	b.options.IndexCacheSize = val
}

func (b BadgerOptionsAdapter) SetSyncWrites(val bool) {
	b.options.SyncWrites = val
}
