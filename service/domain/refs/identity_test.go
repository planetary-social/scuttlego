package refs_test

import (
	"strings"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/fixtures"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestIdentityAndFeedFromString(t *testing.T) {
	testCases := []struct {
		Name          string
		Ref           string
		ExpectedError error
	}{
		{
			Name:          "empty_string",
			Ref:           "",
			ExpectedError: errors.New("invalid prefix"),
		},
		{
			Name:          "valid",
			Ref:           "@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519",
			ExpectedError: nil,
		},
		{
			Name:          "too_short",
			Ref:           "@abc=.ed25519",
			ExpectedError: errors.New("could not create a public identity: invalid public key length"),
		},
		{
			Name:          "too_long",
			Ref:           "@" + strings.Repeat("a", 1000) + ".ed25519",
			ExpectedError: errors.New("could not create a public identity: invalid public key length"),
		},
		{
			Name:          "invalid_base64",
			Ref:           "@notbase64.ed25519",
			ExpectedError: errors.New("invalid base64: illegal base64 data at input byte 8"),
		},
		{
			Name:          "missing_prefix",
			Ref:           "qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519",
			ExpectedError: errors.New("invalid prefix"),
		},
		{
			Name:          "missing_suffix",
			Ref:           "@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=",
			ExpectedError: errors.New("invalid suffix"),
		},
		{
			Name:          "invalid_suffix",
			Ref:           "@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ggfeed-v1",
			ExpectedError: errors.New("invalid suffix"),
		},
	}

	types := []struct {
		Name string
		Func func(s string) error
	}{
		{
			Name: "identity",
			Func: func(s string) error {
				_, err := refs.NewIdentity(s)
				return err
			},
		},
		{
			Name: "feed",
			Func: func(s string) error {
				_, err := refs.NewFeed(s)
				return err
			},
		},
	}

	for _, typ := range types {
		t.Run(typ.Name, func(t *testing.T) {
			for _, testCase := range testCases {
				t.Run(testCase.Name, func(t *testing.T) {
					err := typ.Func(testCase.Ref)
					if testCase.ExpectedError == nil {
						require.NoError(t, err)
					} else {
						require.EqualError(t, err, testCase.ExpectedError.Error())
					}
				})
			}
		})
	}
}

func TestNewIdentityFromPublic(t *testing.T) {
	public := fixtures.SomePublicIdentity()

	i, err := refs.NewIdentityFromPublic(public)
	require.NoError(t, err)

	require.True(t, strings.HasPrefix(i.String(), "@"))
	require.True(t, strings.HasSuffix(i.String(), ".ed25519"))
}

func TestMainFeed(t *testing.T) {
	const ref = "@qFtLJ6P5Eh9vKxnj7Rsh8SkE6B6Z36DVLP7ZOKNeQ/Y=.ed25519"

	i, err := refs.NewIdentity(ref)
	require.NoError(t, err)

	f := i.MainFeed()
	require.Equal(t, ref, f.String(), "main feed should the same as the identity")
}
