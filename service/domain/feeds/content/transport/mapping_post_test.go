package transport_test

import (
	"testing"

	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingPostUnmarshal(t *testing.T) {
	imageRef := refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256")
	post := msgcontents.MustNewPost([]refs.Blob{imageRef})
	postWithoutBlobs := msgcontents.MustNewPost(nil)

	testCases := []struct {
		Name         string
		Content      string
		ExpectedPost msgcontents.Post
	}{
		{
			Name: "no_mentions",
			Content: `{
				"type": "post",
				"text": "YES WE CAN! :heart: :smiley_cat:"
			}`,
			ExpectedPost: postWithoutBlobs,
		},
		{
			Name: "complex",
			Content: `{
				"type": "post",
				"text": "YES WE CAN! :heart: :smiley_cat:",
				"root": "%Yx6/snCfur1NHd9fov8H359DfqTyncxuh93uZKnLQI8=.sha256",
				"branch": "%X8PLQuBhdUA+WF5VANRfG5iAKMNAeBXxlAtd9SKDAtM=.sha256",
				"channel": "patchfox",
				"mentions": [
					{
						"link": "&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256",
						"type": "image/jpeg",
						"size": 195993
					}
				]
			}`,
			ExpectedPost: post,
		},
		{
			Name: "link_which_is_not_a_blob",
			Content: `{
				"type": "post",
				"text": "YES WE CAN! :heart: :smiley_cat:",
				"root": "%Yx6/snCfur1NHd9fov8H359DfqTyncxuh93uZKnLQI8=.sha256",
				"branch": "%X8PLQuBhdUA+WF5VANRfG5iAKMNAeBXxlAtd9SKDAtM=.sha256",
				"channel": "patchfox",
				"mentions": [
					{
						"link": "#channel"
					}
				]
			}`,
			ExpectedPost: postWithoutBlobs,
		},
		{
			Name: "mentions_which_are_a_map",
			Content: `{
				"type": "post",
				"text": "a new photo from #dweb-camp 2022! ![photo.bmp](&O0h21NiGLLmjCF1kD2xWllvPExwe6t5P+F7YK3HAX4g=.sha256)",
				"mentions": {
					"0":{"name":"photo.bmp","type": "image/bmp"}
				}
			}`,
			ExpectedPost: postWithoutBlobs, // todo probably actually scan the markdown
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			marshaler := newMarshaler(t)

			msg, err := marshaler.Unmarshal(message.MustNewRawMessageContent([]byte(testCase.Content)))
			require.NoError(t, err)

			require.Equal(
				t,
				testCase.ExpectedPost,
				msg,
			)
		})
	}
}
