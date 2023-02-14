package blobs_test

import (
	"testing"

	"github.com/planetary-social/scuttlego/service/domain/blobs"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestScanner(t *testing.T) {
	testCases := []struct {
		Name         string
		Content      string
		ExpectedRefs []refs.Blob
	}{
		{
			Name: "about_with_image",
			Content: `{
				"type": "about",
				"about": "@G/zUdqlPMsdgsdg8yXIfMjx1676ApAOghwgc=.ed25519",
				"image": "&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256",
				"name": "Hilo",
				"description": "extrem klug"
			}`,
			ExpectedRefs: []refs.Blob{
				refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256"),
			},
		},
		{
			Name: "about_with_blank_string_image",
			Content: `{
				"type": "about",
				"about": "@G/zUdqlPMsdgsdg8yXIfMjx1676ApAOghwgc=.ed25519",
				"image": "",
				"name": "Hilo",
				"description": "extrem klug"
			}`,
			ExpectedRefs: nil,
		},
		{
			Name: "about_with_missing_image",
			Content: `{
				"type": "about",
				"about": "@G/zUdqlPMsdgsdg8yXIfMjx1676ApAOghwgc=.ed25519",
				"name": "Hilo",
				"description": "extrem klug"
			}`,
			ExpectedRefs: nil,
		},
		{
			Name: "about_with_complex_image",
			Content: `{
				"type": "about",
				"image":{
					"size": 112320,
					"width": 1000,
					"height": 1000,
					"type": "image/jpeg",
					"link": "&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256"
				},
				"about": "@RhjO4sEkb0mKcuO/QvAxbsOLY7k260EjQN6kDAkA6Sk=.ed25519",
				"name": "Coach Tony"
			}`,
			ExpectedRefs: []refs.Blob{
				refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256"),
			},
		},

		{
			Name: "post_without_mentions",
			Content: `{
				"type": "post",
				"text": "YES WE CAN! :heart: :smiley_cat:"
			}`,
			ExpectedRefs: nil,
		},
		{
			Name: "post_with_complex_image_mention",
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
			ExpectedRefs: []refs.Blob{
				refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256"),
			},
		},
		{
			Name: "post_with_duplicate_complex_image_mentions",
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
					},
					{
						"link": "&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256",
						"type": "image/jpeg",
						"size": 195993
					}
				]
			}`,
			ExpectedRefs: []refs.Blob{
				refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256"),
			},
		},
		{
			Name: "post_with_non_image_mention",
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
			ExpectedRefs: nil,
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
			ExpectedRefs: []refs.Blob{
				refs.MustNewBlob("&O0h21NiGLLmjCF1kD2xWllvPExwe6t5P+F7YK3HAX4g=.sha256"),
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			scanner := blobs.NewScanner()

			result, err := scanner.Scan(message.MustNewRawMessageContent([]byte(testCase.Content)))
			require.NoError(t, err)
			require.Equal(t,
				testCase.ExpectedRefs,
				result,
			)
		})
	}
}
