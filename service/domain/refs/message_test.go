package refs_test

import (
	"strings"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMessage(t *testing.T) {
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
			Ref:           "%oW1jOoJUL5NIYNpMxDYC1np8rT9Nn4h0h80tdJ95NsY=.sha256",
			ExpectedError: nil,
		},
		{
			Name:          "too_short",
			Ref:           "%abc=.sha256",
			ExpectedError: errors.New("invalid hash length '2'"),
		},
		{
			Name:          "too_long",
			Ref:           "%" + strings.Repeat("a", 1000) + ".sha256",
			ExpectedError: errors.New("invalid hash length '750'"),
		},
		{
			Name:          "invalid_base64",
			Ref:           "%notbase64.sha256",
			ExpectedError: errors.New("invalid base64: illegal base64 data at input byte 8"),
		},
		{
			Name:          "missing_prefix",
			Ref:           "oW1jOoJUL5NIYNpMxDYC1np8rT9Nn4h0h80tdJ95NsY=.sha256",
			ExpectedError: errors.New("invalid prefix"),
		},
		{
			Name:          "missing_suffix",
			Ref:           "%oW1jOoJUL5NIYNpMxDYC1np8rT9Nn4h0h80tdJ95NsY=",
			ExpectedError: errors.New("invalid suffix"),
		},
		{
			Name:          "invalid_suffix",
			Ref:           "%oW1jOoJUL5NIYNpMxDYC1np8rT9Nn4h0h80tdJ95NsY=.md5",
			ExpectedError: errors.New("invalid suffix"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := refs.NewMessage(testCase.Ref)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}
