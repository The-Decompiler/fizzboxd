package main

import (
	"testing"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

type Data struct {
	Title       string
	Description string
}

type Result struct {
	Poster  string
	Review  string
	Spoiler bool
}

type TestData struct {
	Data     Data
	Expected Result
}

var testDatas = []*TestData{
	{
		// [1] Poster: False; Review: True
		Data{"Chinatown, 1974 - ★★★★", " <p>amazing</p> "},
		Result{"", "amazing", false},
	},
	{
		// [1] Poster: False; Review: False
		Data{"Chinatown, 1974 - ★★★★", " <p>Watched on Monday January 1, 0001.</p> "},
		Result{"", "", false},
	},
	{
		// [2] Poster: False; Review: Spoiler
		Data{"Eureka, 2000 - ★★★★★ (contains spoilers)",
			" <p><em>This review may contain spoilers.</em></p> <p><i>We need some time to find ourselves.</i></p><p>Eureka is ... hard enough.</p> "},
		Result{"", "We need some time to find ourselves.   Eureka is ... hard enough.", true},
	},
	{
		// [2] Poster: True; Review: True
		Data{"Chinatown, 1974 - ★★★★", " <p><img src=\"https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188\"/></p> <p>some review or something.</p> "},
		Result{"https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188", "some review or something.", false},
	},
	{
		// [2] Poster: True; Review: False
		Data{"Chinatown, 1974 - ★★★★", " <p><img src=\"https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188\"/></p> <p>Watched on Monday January 1, 0001.</p> "},
		Result{"https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188", "", false},
	},
	{
		// [3] Poster: True; Review: Spoiler
		Data{"Eureka, 2000 - ★★★★★ (contains spoilers)", " <p><img src=\"https://a.ltrbxd.com/resized/film-poster/2/6/4/8/4/26484-eureka-0-500-0-750-crop.jpg?k=6a44c9e520\"/></p> <p><em>This review may contain spoilers.</em></p> <p><i>We need some time to find ourselves.</i></p><p>Eureka is ... hard enough.</p> "},
		Result{"https://a.ltrbxd.com/resized/film-poster/2/6/4/8/4/26484-eureka-0-500-0-750-crop.jpg?k=6a44c9e520", "We need some time to find ourselves.   Eureka is ... hard enough.", true},
	},
}

func TestHandleData(t *testing.T) {
	// stripHTMLTags
	var policy = bluemonday.StripTagsPolicy().AddSpaceWhenStrippingTag(true)

	for i := 0; i < len(testDatas); i++ {
		poster, review, spoiler, err := HandleData(testDatas[i].Data.Title, testDatas[i].Data.Description, time.Time{}, policy)
		if err != nil {
			t.Errorf("test failed: %v\n", err)
		}

		if poster != testDatas[i].Expected.Poster {
			t.Errorf("\nPoster Received: %v\nPoster Expected: %v", poster, testDatas[i].Expected.Poster)
		}
		if review != testDatas[i].Expected.Review {
			t.Errorf("\nReview Received: %v\nReview Expected: %v", review, testDatas[i].Expected.Review)
		}
		if spoiler != testDatas[i].Expected.Spoiler {
			t.Errorf("\nSpoiler Received: %v\nSpoiler Expected: %v", spoiler, testDatas[i].Expected.Spoiler)
		}
	}
}
