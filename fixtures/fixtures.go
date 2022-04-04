package fixtures

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/logging"
	"github.com/planetary-social/go-ssb/service/domain/feeds/content"
	"github.com/planetary-social/go-ssb/service/domain/feeds/message"
	"github.com/planetary-social/go-ssb/service/domain/identity"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/planetary-social/go-ssb/service/domain/transport/rpc"
	"github.com/sirupsen/logrus"
	"go.etcd.io/bbolt"
)

func SomeLogger() logging.Logger {
	return logging.NewLogrusLogger(logrus.New(), "test", logging.LevelTrace)
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

func SomeInt() int {
	return rand.Int()
}

func SomeUint32() uint32 {
	return rand.Uint32()
}

func SomeRefMessage() refs.Message {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewMessage(fmt.Sprintf("%%%s.sha256", randomBase64(32)))
}

func SomeRefAuthor() refs.Identity {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewIdentity(fmt.Sprintf("@%s.ed25519", randomBase64(32)))
}

func SomeRefFeed() refs.Feed {
	// todo improve this by using some kind of a better constructor
	return refs.MustNewFeed(fmt.Sprintf("@%s.ed25519", randomBase64(32)))
}

func SomeTime() time.Time {
	// todo improve this by using some kind of a better constructor
	return time.Unix(rand.Int63(), 0)
}

func SomePublicIdentity() identity.Public {
	v, err := identity.NewPrivate()
	if err != nil {
		panic(err)
	}
	return v.Public()
}

func SomeContent() message.MessageContent {
	return content.MustNewUnknown(SomeBytes())
}

func SomeSequence() message.Sequence {
	return message.MustNewSequence(rand.Int())
}

func SomeBytes() []byte {
	r := make([]byte, 10+rand.Intn(100))
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
	return message.NewRawMessage([]byte(msg))
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
		SomeRefAuthor(),
		feed,
		SomeTime(),
		SomeContent(),
		SomeRawMessage(),
	)
}

func randomBase64(bytes int) string {
	r := make([]byte, bytes)
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return base64.StdEncoding.EncodeToString(r)
}

func File(t *testing.T) string {
	file, err := ioutil.TempFile("", "eggplant_test")
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

func Bolt(t *testing.T) *bbolt.DB {
	file := File(t)

	db, err := bbolt.Open(file, 0600, &bbolt.Options{Timeout: 5 * time.Second})
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
