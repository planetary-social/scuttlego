package transport_test

import (
	"encoding/hex"
	"testing"

	"github.com/boreq/errors"
	"github.com/planetary-social/scuttlego/internal/fixtures"
	"github.com/planetary-social/scuttlego/service/domain/transport/rpc/transport"
	"github.com/stretchr/testify/require"
)

func TestMessageHeader(t *testing.T) {
	const validBodyLength = 10
	const validRequestNumber = 1
	const validResponseNumber = -1

	testCases := []struct {
		Name string

		Flags         transport.MessageHeaderFlags
		BodyLength    uint32
		RequestNumber int32

		ExpectedError error
	}{
		{
			Name:          "valid_request",
			Flags:         transport.MustNewMessageHeaderFlags(false, false, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: nil,
		},
		{
			Name:          "valid_request_termination",
			Flags:         transport.MustNewMessageHeaderFlags(false, true, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: nil,
		},
		{
			Name:          "valid_stream_request",
			Flags:         transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: nil,
		},
		{
			Name:          "valid_stream_request_termination",
			Flags:         transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: nil,
		},

		{
			Name:          "request_number_must_not_be_zero",
			Flags:         transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: 0,
			ExpectedError: errors.New("request number can not be set to zero"),
		},

		{
			Name:          "request_body_type_can_be_binary_in_duplex_streams",
			Flags:         transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeBinary),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: nil,
		},
		{
			Name:          "request_body_type_can_be_string_in_duplex_streams",
			Flags:         transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeString),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: nil,
		},
		{
			Name:          "request_body_type_can_not_be_binary_if_termination",
			Flags:         transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeBinary),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: errors.New("terminations should have body type set to JSON"),
		},
		{
			Name:          "request_body_type_can_not_be_string_if_termination",
			Flags:         transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeString),
			BodyLength:    validBodyLength,
			RequestNumber: validRequestNumber,
			ExpectedError: errors.New("terminations should have body type set to JSON"),
		},

		{
			Name:          "valid_response",
			Flags:         transport.MustNewMessageHeaderFlags(false, false, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validResponseNumber,
			ExpectedError: nil,
		},
		{
			Name:          "valid_response_termination",
			Flags:         transport.MustNewMessageHeaderFlags(false, true, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validResponseNumber,
			ExpectedError: nil,
		},
		{
			Name:          "valid_stream_response",
			Flags:         transport.MustNewMessageHeaderFlags(true, false, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validResponseNumber,
			ExpectedError: nil,
		},
		{
			Name:          "valid_stream_response_termination",
			Flags:         transport.MustNewMessageHeaderFlags(true, true, transport.MessageBodyTypeJSON),
			BodyLength:    validBodyLength,
			RequestNumber: validResponseNumber,
			ExpectedError: nil,
		},

		{
			Name:          "header_flag_can_not_be_a_zero_value",
			Flags:         transport.MessageHeaderFlags{},
			BodyLength:    validBodyLength,
			RequestNumber: validResponseNumber,
			ExpectedError: errors.New("zero value of flags"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			h, err := transport.NewMessageHeader(testCase.Flags, testCase.BodyLength, testCase.RequestNumber)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)

				headerBytes, err := h.Bytes()
				require.NoError(t, err)

				t.Log(hex.EncodeToString(headerBytes))

				headerFromBytes, err := transport.NewMessageHeaderFromBytes(headerBytes)
				require.NoError(t, err)

				require.Equal(t, h, headerFromBytes)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}
}

func TestNewMessageHeaderFromBytes(t *testing.T) {
	testCases := []struct {
		Name             string
		HeaderBytesInHex string
		ExpectedHeader   transport.MessageHeader
	}{
		{
			Name:             "request",
			HeaderBytesInHex: "020000000a00000001",
			ExpectedHeader: transport.MustNewMessageHeader(
				transport.MustNewMessageHeaderFlags(false, false, transport.MessageBodyTypeJSON),
				10,
				1,
			),
		},
		{
			Name:             "response",
			HeaderBytesInHex: "020000000affffffff",
			ExpectedHeader: transport.MustNewMessageHeader(
				transport.MustNewMessageHeaderFlags(false, false, transport.MessageBodyTypeJSON),
				10,
				-1,
			),
		},
	}

	for _, testCase := range testCases {
		headerBytes, err := hex.DecodeString(testCase.HeaderBytesInHex)
		require.NoError(t, err)

		header, err := transport.NewMessageHeaderFromBytes(headerBytes)
		require.NoError(t, err)

		require.Equal(t, testCase.ExpectedHeader, header)
	}
}

func TestNewMessageHeaderFromBytes_DoesNotPanicOnInvalidHeaderLength(t *testing.T) {
	_, err := transport.NewMessageHeaderFromBytes(nil)
	require.EqualError(t, err, "invalid header length 0")
}

func TestMessageHeaderFlags(t *testing.T) {
	testCases := []struct {
		Name string

		BodyType transport.MessageBodyType

		ExpectedError error
	}{
		{
			Name:          "valid",
			BodyType:      transport.MessageBodyTypeString,
			ExpectedError: nil,
		},
		{
			Name:          "zero_value_of_message_body_type",
			BodyType:      transport.MessageBodyType{},
			ExpectedError: errors.New("zero value of message body type"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			_, err := transport.NewMessageHeaderFlags(fixtures.SomeBool(), fixtures.SomeBool(), testCase.BodyType)
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
		})
	}

}
