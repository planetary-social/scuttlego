package transport_test

import (
	"testing"

	msgcontents "github.com/planetary-social/scuttlego/service/domain/feeds/content"
	"github.com/planetary-social/scuttlego/service/domain/feeds/message"
	"github.com/planetary-social/scuttlego/service/domain/refs"
	"github.com/stretchr/testify/require"
)

func TestMappingAboutUnmarshal(t *testing.T) {
	imageRef := refs.MustNewBlob("&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256")
	about := msgcontents.MustNewAbout(&imageRef)
	aboutWithoutImage := msgcontents.MustNewAbout(nil)

	testCases := []struct {
		Name          string
		Content       string
		ExpectedAbout msgcontents.About
	}{
		{
			Name: "populated_string",
			Content: `{
				"type": "about",
				"about": "@G/zUdqlPMsdgsdg8yXIfMjx1676ApAOghwgc=.ed25519",
				"image": "&Uv38ByGCZU8WP18PmmIdcpVmx00QA3xNe7sEB9Hixkk=.sha256",
				"name": "Hilo",
				"description": "extrem klug"
			}`,
			ExpectedAbout: about,
		},
		{
			Name: "empty_string",
			Content: `{
				"type": "about",
				"about": "@G/zUdqlPMsdgsdg8yXIfMjx1676ApAOghwgc=.ed25519",
				"image": "",
				"name": "Hilo",
				"description": "extrem klug"
			}`,
			ExpectedAbout: aboutWithoutImage,
		},
		{
			Name: "missing_string",
			Content: `{
				"type": "about",
				"about": "@G/zUdqlPMsdgsdg8yXIfMjx1676ApAOghwgc=.ed25519",
				"name": "Hilo",
				"description": "extrem klug"
			}`,
			ExpectedAbout: aboutWithoutImage,
		},
		{
			Name: "complex",
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
			ExpectedAbout: about,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			marshaler := newMarshaler(t)

			msg, err := marshaler.Unmarshal(message.MustNewRawMessageContent([]byte(testCase.Content)))
			require.NoError(t, err)

			require.Equal(
				t,
				testCase.ExpectedAbout,
				msg,
			)
		})
	}
}
