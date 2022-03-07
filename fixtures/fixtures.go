package fixtures

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/planetary-social/go-ssb/refs"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/content"
	"github.com/planetary-social/go-ssb/scuttlebutt/feeds/message"
	bolt "go.etcd.io/bbolt"
)

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

func SomeContent() message.MessageContent {
	return content.MustNewUnknown(SomeBytes())
}

func SomeBytes() []byte {
	r := make([]byte, 10+rand.Intn(100))
	_, err := rand.Read(r)
	if err != nil {
		panic(err)
	}
	return r
}

func SomeRawMessage() message.RawMessage {
	return message.NewRawMessage(SomeBytes())
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

func Bolt(t *testing.T) *bolt.DB {
	file := File(t)

	db, err := bolt.Open(file, 0600, &bolt.Options{Timeout: 5 * time.Second})
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
