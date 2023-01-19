package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/planetary-social/scuttlego/logging"
	"github.com/planetary-social/scuttlego/service/app/common"
	"github.com/planetary-social/scuttlego/service/domain/bans"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/identity"
	"github.com/planetary-social/scuttlego/service/domain/invites"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/rooms/aliases"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/sirupsen/logrus"
)

func SomeLogger() logging.Logger {
	return logging.NewLogrusLogger(logrus.New(), "test", logging.LevelTrace)
}

func TestLogger(t *testing.T) logging.Logger {
	return newTestingLogger(t.Name(), t)
}

func TestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	return ctx
}

func SomeProcedureName() rpc.ProcedureName {
	return rpc.MustNewProcedureName([]string{randomBase64(10)})
}

func SomeProcedureType() rpc.ProcedureType {
	return rpc.ProcedureTypeAsync
}

func SomeBool() bool {
	return rand.Int()%2 == 0
}

func SomeNonNegativeInt() int {
	return rand.Int()
}

func SomePositiveInt() int {
	return 1 + rand.Intn(math.MaxInt-1)
}

func SomeUint32() uint32 {
	return rand.Uint32()
}

func SomePositiveInt32() int32 {
	return int32(rand.Intn(math.MaxInt32))
}

func SomeNegativeInt32() int32 {
	return int32(-rand.Intn(-math.MinInt32))
}

func SomeRefMessage() refs.Message {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewMessage(fmt.Sprintf("%%%s.sha256", randomBase64(32)))
}

func SomeRefIdentity() refs.Identity {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewIdentity(fmt.Sprintf("@%s.ed25519", randomBase64(32)))
}

func SomeRefFeed() refs.Feed {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewFeed(fmt.Sprintf("@%s.ed25519", randomBase64(32)))
}

func SomeRefBlob() refs.Blob {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewBlob(fmt.Sprintf("&%s.sha256", randomBase64(32)))
}

func SomeTime() time.Time {
	// todo improve this by using some kind of a better constructor
	return time.Unix(rand.Int63(), 0)
}

func SomeDuration() time.Duration {
	return time.Duration(time.Duration(SomePositiveInt32()) * time.Second)
}

func SomePublicIdentity() identity.Public {
	return SomePrivateIdentity().Public()
}

func SomePrivateIdentity() identity.Private {
	v, err := identity.NewPrivate()
	if err != nil {
		panic(err)
	}
	return v
}

func SomeContent() content.KnownMessageContent {
	return content.MustNewUnknown(SomeBytes())
}

func SomeMessageBodyType() transport.MessageBodyType {
	if rand.Int()%2 == 0 {
		return transport.MessageBodyTypeJSON
	} else {
		return transport.MessageBodyTypeBinary
	}
}

func SomeSequence() message.Sequence {
	return message.MustNewSequence(rand.Int())
}

func SomeReceiveLogSequence() common.ReceiveLogSequence {
	return common.MustNewReceiveLogSequence(rand.Int())
}

func SomeString() string {
	return strconv.Itoa(SomeNonNegativeInt())
}

func SomeAlias() aliases.Alias {
	return aliases.MustNewAlias("alias" + SomeString())
}

func SomeBytes() []byte {
	r := make([]byte, 10+rand.Intn(100))
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return r
}

func SomeBytesOfLength(n int) []byte {
	r := make([]byte, n)
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return r
}

func SomeJSON() []byte {
	return []byte(`{"key":"value"}`)
}

func SomeRawMessage() message.RawMessage {
	msg := `
{
	"previous": null,
	"author": "@FCX/tsDLpubCPKKfIrw4gc+SQkHcaD17s7GI6i/ziWY=.ed25519",
	"sequence": 1,
	"timestamp": 1514517067954,
	"hash": "sha256",
	"content": {
	"type": "post",
		"text": "This is the first post!"
	},
	"signature": "QYOR/zU9dxE1aKBaxc3C0DJ4gRyZtlMfPLt+CGJcY73sv5abKKKxr1SqhOvnm8TY784VHE8kZHCD8RdzFl1tBA==.sig.ed25519"
}
`
	return message.MustNewRawMessage([]byte(msg))
}

func SomeRawMessageContent() message.RawMessageContent {
	return message.MustNewRawMessageContent(SomeBytes())
}

func SomeMessage(seq message.Sequence, feed refs.Feed) message.Message {
	var previous *refs.Message
	if !seq.IsFirst() {
		tmp := SomeRefMessage()
		previous = &tmp
	}

	return message.MustNewMessage(
		SomeRefMessage(),
		previous,
		seq,
		SomeRefIdentity(),
		feed,
		SomeTime(),
		SomeContent(),
		SomeRawMessage(),
	)
}

func SomeMessageWithUniqueRawMessage(seq message.Sequence, feed refs.Feed) message.Message {
	var previous *refs.Message
	if !seq.IsFirst() {
		tmp := SomeRefMessage()
		previous = &tmp
	}

	return message.MustNewMessage(
		SomeRefMessage(),
		previous,
		seq,
		SomeRefIdentity(),
		feed,
		SomeTime(),
		SomeContent(),
		message.MustNewRawMessage(SomeBytes()),
	)
}

func SomeConnectionId() rpc.ConnectionId {
	return rpc.NewConnectionId(SomeNonNegativeInt())
}

func SomeWantDistance() blobs.WantDistance {
	return blobs.MustNewWantDistance(SomeNonNegativeInt())
}

func SomeSize() blobs.Size {
	return blobs.MustNewSize(int64(SomePositiveInt32()))
}

func SomeError() error {
	return fmt.Errorf("some error: %d", rand.Int())
}

func SomeBanListHash() bans.Hash {
	r := make([]byte, 32)
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return bans.MustNewHash(r)
}

func SomeInvite() invites.Invite {
	return invites.MustNewInviteFromString("one.planetary.pub:8008:@CIlwTOK+m6v1hT2zUVOCJvvZq7KE/65ErN6yA2yrURY=.ed25519~KVvak/aZeQJQUrn1imLIvwU+EVTkCzGW8TJWTmK8lOk=")
}

func randomBase64(bytes int) string {
	r := make([]byte, bytes)
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(r)
}

func Directory(t *testing.T) string {
	name, err := os.MkdirTemp("", "scuttlego-test")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := os.RemoveAll(name)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(cleanup)

	return name
}

func File(t *testing.T) string {
	file, err := os.CreateTemp("", "scuttlego-test")
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := os.Remove(file.Name())
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(cleanup)

	return file.Name()
}

func Badger(t *testing.T) *badger.DB {
	dir := Directory(t)

	db, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		t.Fatal(err)
	}

	cleanup := func() {
		err := db.Close()
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(cleanup)

	return db
}

func TestFileRelativePath(elem ...string) string {
	_, filename, _, _ := runtime.Caller(1)
	elems := []string{filepath.Dir(filename)}
	elems = append(elems, elem...)
	return filepath.Join(elems...)
}

type testingLogger struct {
	name string
	t    *testing.T
	log  func(args ...any)
}

func newTestingLogger(name string, t *testing.T) testingLogger {
	return testingLogger{name: name, t: t, log: t.Log}
}

func (t testingLogger) New(name string) logging.Logger {
	return testingLogger{name: t.name + "." + name, t: t.t, log: t.log}
}

func (t testingLogger) WithError(err error) logging.Logger {
	return t.withField("err", err)
}

func (t testingLogger) WithField(key string, v any) logging.Logger {
	t.t.Helper()
	return t.withField(key, v)
}

func (t testingLogger) Error(message string) {
	t.t.Helper()
	t.withField("level", "error").log(message)
}

func (t testingLogger) Debug(message string) {
	t.t.Helper()
	t.withField("level", "debug").log(message)
}

func (t testingLogger) Trace(message string) {
	t.t.Helper()
	t.withField("level", "trace").log(message)
}

func (t testingLogger) withField(key string, v any) testingLogger {
	prev := t.log
	return testingLogger{name: t.name, t: t.t, log: func(args ...any) {
		t.t.Helper()
		tmp := []any{fmt.Sprintf("%s=%s", key, v)}
		tmp = append(tmp, args...)
		prev(tmp)
	}}
}
