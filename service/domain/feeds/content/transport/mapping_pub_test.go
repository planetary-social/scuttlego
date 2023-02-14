package transport_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/known"
	"github.com/planetary-social/scuttlego/service/domain/feeds/content/transport"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingPubUnmarshal(t *testing.T) {
	marshaler := newMarshaler(t)

	content := `
{
	"type": "pub",
	"address": {
		"host": "one.butt.nz",
		"port": 8008,
		"key": "@VJM7w1W19ZsKmG2KnfaoKIM66BRoreEkzaVm/J//wl8=.ed25519"
	}
}`

	msg, err := marshaler.Unmarshal(message.MustNewRawContent([]byte(content)))
	require.NoError(t, err)

	require.Equal(
		t,
		known.MustNewPub(
			refs.MustNewIdentity("@VJM7w1W19ZsKmG2KnfaoKIM66BRoreEkzaVm/J//wl8=.ed25519"),
			"one.butt.nz",
			8008,
		),
		msg,
	)
}

func TestMappingPubMarshal(t *testing.T) {
	marshaler := newMarshaler(t)

	msg := known.MustNewPub(
		refs.MustNewIdentity("@VJM7w1W19ZsKmG2KnfaoKIM66BRoreEkzaVm/J//wl8=.ed25519"),
		"one.butt.nz",
		8008,
	)

	raw, err := marshaler.Marshal(msg)
	require.NoError(t, err)

	require.Equal(
		t,
		`{"type":"pub","address":{"key":"@VJM7w1W19ZsKmG2KnfaoKIM66BRoreEkzaVm/J//wl8=.ed25519","host":"one.butt.nz","port":8008}}`,
		string(raw.Bytes()),
	)
}

func newMarshaler(t *testing.T) *transport.Marshaler {
	marshaler, err := transport.NewMarshaler(transport.DefaultMappings(), fixtures.SomeLogger())
	require.NoError(t, err)

	return marshaler
}
