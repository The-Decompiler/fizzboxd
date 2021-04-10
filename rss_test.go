package main

import (
	"testing"
	"time"
)

func TestHandleData(t *testing.T) {
	testDatas := [6][2][2]string{
		// [1] Poster: False; Review: True
		{{ "Chinatown, 1974 - ★★★★", " <p>amazing</p> " }, { "", "amazing" }},
		// [1] Poster: False; Review: False
		{{ "Chinatown, 1974 - ★★★★", " <p>Watched on Monday January 1, 0001.</p> " }, { "", "" }},
		// [2] Poster: False; Review: Spoiler
		{{ "Eureka, 2000 - ★★★★★ (contains spoilers)", " <p><em>This review may contain spoilers.</em></p> <p><i>We need some time to find ourselves.</i></p><p>Eureka is ... hard enough.</p> " }, { "", "This review may contain spoilers."}},
		// [2] Poster: True; Review: True
		{{ "Chinatown, 1974 - ★★★★", " <p><img src=\"https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188\"/></p> <p>some review or something.</p> " }, { "https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188", "some review or something." }},
		// [2] Poster: True; Review: False
		{{ "Chinatown, 1974 - ★★★★", " <p><img src=\"https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188\"/></p> <p>Watched on Monday January 1, 0001.</p> " }, { "https://a.ltrbxd.com/resized/sm/upload/vc/vi/qv/i7/2WEog45eMBUFRgxhIbxx2nOaaMZ-0-500-0-750-crop.jpg?k=1ad1e0a188", "" }},
		// [3] Poster: True; Review: Spoiler
		{{ "Eureka, 2000 - ★★★★★ (contains spoilers)", " <p><img src=\"https://a.ltrbxd.com/resized/film-poster/2/6/4/8/4/26484-eureka-0-500-0-750-crop.jpg?k=6a44c9e520\"/></p> <p><em>This review may contain spoilers.</em></p> <p><i>We need some time to find ourselves.</i></p><p>Eureka is ... hard enough.</p> " }, { "https://a.ltrbxd.com/resized/film-poster/2/6/4/8/4/26484-eureka-0-500-0-750-crop.jpg?k=6a44c9e520", "This review may contain spoilers." }},
	}

	for i := 0; i < len(testDatas); i++ {
		poster, review := handleData(testDatas[i][0][0], testDatas[i][0][1], time.Time{})
		if poster != testDatas[i][1][0] {
			t.Errorf("Poster Received: %v", poster)
			t.Errorf("Poster Expected: %v", testDatas[i][1][0])
		}
		if review != testDatas[i][1][1] {
			t.Errorf("Review Received: %v", review)
			t.Errorf("Review Expected: %v", testDatas[i][1][1])
		}
	}
}
