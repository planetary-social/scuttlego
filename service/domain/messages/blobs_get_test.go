package messages_test

import (
	"testing"

	"github.com/planetary-social/go-ssb/service/domain/blobs"
	"github.com/planetary-social/go-ssb/service/domain/messages"
	"github.com/planetary-social/go-ssb/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestNewBlobsMarshal(t *testing.T) {
	testCases := []struct {
		Name string

		Hash refs.Blob
		Size *blobs.Size
		Max  *blobs.Size

		ExpectedJSON string
	}{
		{
			Name: "only_hash",

			Hash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			Size: nil,
			Max:  nil,

			ExpectedJSON: `["\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"]`,
		},
		{
			Name: "hash_and_size",

			Hash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			Size: sizePtr(blobs.MustNewSize(123)),
			Max:  nil,

			ExpectedJSON: `[{"hash":"\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256","size":123}]`,
		},
		{
			Name: "hash_and_max",

			Hash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			Size: nil,
			Max:  sizePtr(blobs.MustNewSize(123)),

			ExpectedJSON: `[{"hash":"\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256","max":123}]`,
		},
		{
			Name: "hash_and_size_and_max",

			Hash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			Size: sizePtr(blobs.MustNewSize(123)),
			Max:  sizePtr(blobs.MustNewSize(456)),

			ExpectedJSON: `[{"hash":"\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256","size":123,"max":456}]`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			args, err := messages.NewBlobsGetArguments(testCase.Hash, testCase.Size, testCase.Max)
			require.NoError(t, err)

			j, err := args.MarshalJSON()
			require.NoError(t, err)

			require.Equal(t, testCase.ExpectedJSON, string(j))
		})
	}
}

func TestNewBlobsGetArgumentsFromBytesString(t *testing.T) {
	args, err := messages.NewBlobsGetArgumentsFromBytes([]byte(`["&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"]`))
	require.NoError(t, err)

	require.Equal(t, "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", args.Hash().String())
}

func TestNewBlobsGetArgumentsFromBytesObject(t *testing.T) {
	args, err := messages.NewBlobsGetArgumentsFromBytes([]byte(`[{"hash": "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", "size": 161699, "max": 200000}]`))
	require.NoError(t, err)

	require.Equal(t, "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", args.Hash().String())

	size, ok := args.Size()
	require.True(t, ok)
	require.Equal(t, blobs.MustNewSize(161699), size)

	max, ok := args.Max()
	require.True(t, ok)
	require.Equal(t, blobs.MustNewSize(200000), max)
}

func sizePtr(s blobs.Size) *blobs.Size {
	return &s
}
