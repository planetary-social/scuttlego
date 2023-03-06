package messages_test

import (
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal"
	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/messages"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc"
	"github.com/stretchr/testify/require"
)

func TestNewBlobsGet(t *testing.T) {
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
			Size: internal.Ptr(blobs.MustNewSize(123)),
			Max:  nil,

			ExpectedJSON: `[{"hash":"\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256","size":123}]`,
		},
		{
			Name: "hash_and_max",

			Hash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			Size: nil,
			Max:  internal.Ptr(blobs.MustNewSize(123)),

			ExpectedJSON: `[{"hash":"\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256","max":123}]`,
		},
		{
			Name: "hash_and_size_and_max",

			Hash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			Size: internal.Ptr(blobs.MustNewSize(123)),
			Max:  internal.Ptr(blobs.MustNewSize(456)),

			ExpectedJSON: `[{"hash":"\u0026uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256","size":123,"max":456}]`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			args, err := messages.NewBlobsGetArguments(testCase.Hash, testCase.Size, testCase.Max)
			require.NoError(t, err)

			req, err := messages.NewBlobsGet(args)
			require.NoError(t, err)

			require.Equal(t, rpc.MustNewProcedureName([]string{"blobs", "get"}), req.Name())
			require.Equal(t, rpc.ProcedureTypeSource, req.Type())
			require.Equal(t, testCase.ExpectedJSON, string(req.Arguments()))
		})
	}
}

func TestNewBlobsGetArgumentsFromBytesString(t *testing.T) {
	args, err := messages.NewBlobsGetArgumentsFromBytes([]byte(`["&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"]`))
	require.NoError(t, err)

	require.Equal(t, "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", args.Hash().String())
}

func TestNewBlobsGetArgumentsFromBytesObject(t *testing.T) {
	testCases := []struct {
		Name          string
		Payload       string
		ExpectedHash  refs.Blob
		ExpectedSize  *blobs.Size
		ExpectedMax   *blobs.Size
		ExpectedError error
	}{
		{
			Name:         "everything",
			Payload:      `[{"hash": "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", "size": 161699, "max": 200000}]`,
			ExpectedHash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			ExpectedSize: internal.Ptr(blobs.MustNewSize(161699)),
			ExpectedMax:  internal.Ptr(blobs.MustNewSize(200000)),
		},
		{
			Name:         "nil_size",
			Payload:      `[{"hash": "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", "max": 200000}]`,
			ExpectedHash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			ExpectedSize: nil,
			ExpectedMax:  internal.Ptr(blobs.MustNewSize(200000)),
		},
		{
			Name:         "nil_size_nil_max",
			Payload:      `[{"hash": "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"}]`,
			ExpectedHash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			ExpectedSize: nil,
			ExpectedMax:  nil,
		},
		{
			Name:         "nil_max",
			Payload:      `[{"hash": "&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256", "size": 161699}]`,
			ExpectedHash: refs.MustNewBlob("&uaGieSQDJcHfUp6hjIcIq55GoZh4Ug7tNmgaohoxrpw=.sha256"),
			ExpectedSize: internal.Ptr(blobs.MustNewSize(161699)),
			ExpectedMax:  nil,
		},
		{
			Name:         "key_instead_of_hash",
			Payload:      `[{"key":"&eb3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256","max":5242880}]`,
			ExpectedHash: refs.MustNewBlob("&eb3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256"),
			ExpectedSize: nil,
			ExpectedMax:  internal.Ptr(blobs.MustNewSize(5242880)),
		},
		{
			Name:         "identical_key_and_hash",
			Payload:      `[{"hash":"&eb3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256","key":"&eb3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256","max":5242880}]`,
			ExpectedHash: refs.MustNewBlob("&eb3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256"),
			ExpectedSize: nil,
			ExpectedMax:  internal.Ptr(blobs.MustNewSize(5242880)),
		},
		{
			Name:          "different_key_and_hash",
			Payload:       `[{"hash":"&2b3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256","key":"&eb3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQjA=.sha256","max":5242880}]`,
			ExpectedError: errors.New("2 errors occurred:\n\t* error unmarshaling arguments as string: json unmarshal failed: []string: ReadString: expects \" or n, but found {, error found in #2 byte of ...|[{\"hash\":\"&2|..., bigger context ...|[{\"hash\":\"&2b3zi3R00MZ6X+9jXgZMCS6/N1W1PGM2leOEKvpKQ|...\n\t* error unmarshaling arguments as object: could not create a blob ref: key and hash are set but have different values\n\n"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			args, err := messages.NewBlobsGetArgumentsFromBytes([]byte(testCase.Payload))
			if testCase.ExpectedError != nil {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			} else {
				require.NoError(t, err)

				require.Equal(t, testCase.ExpectedHash, args.Hash())

				size, ok := args.Size()
				if testCase.ExpectedSize != nil {
					require.True(t, ok)
					require.Equal(t, *testCase.ExpectedSize, size)
				} else {
					require.False(t, ok)
				}

				max, ok := args.Max()
				if testCase.ExpectedMax != nil {
					require.True(t, ok)
					require.Equal(t, *testCase.ExpectedMax, max)
				} else {
					require.False(t, ok)
				}
			}
		})
	}

}
