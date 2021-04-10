package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"time"

	"github.com/mmcdole/gofeed"
	ext "github.com/mmcdole/gofeed/extensions"
)

type Feed struct {
		Username    string
		DisplayName string
		IconURL     string
		Entries     []*FeedEntry
}

type FeedEntry struct {
		ID          string
		URL         string
		Title       string
		Year        string
		Rating      int
		WatchedDate time.Time
		Rewatch     bool
		Poster      string
		Review      string
}

// Fetches a user's RSS feed, returning an array of 50 FeedEntrys with parsed values
func GetFeed(username string) (Feed) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://letterboxd.com/" + username + "/rss/")

	if err != nil {
		log.Fatalf("failed to fetch feed: %v\n", err)
	}

	entries := []*FeedEntry{}
	for i := 0; i < 50; i++ {
		entries = append(entries, parseEntry(feed.Items[i]))
	}

	return Feed{
		username,
		feed.Items[0].Author.Name,
		"iconURL",
		entries,
	}
}

func parseEntry(entry *gofeed.Item) (*FeedEntry) {
	watchedDate := handleWatchedDate(entry.Extensions["letterboxd"]["watchedDate"])
	poster, review := handleData(entry.Title, entry.Description, watchedDate)

	return &FeedEntry{
		entry.GUID,
		entry.Link,
		handleFilmTitle(entry.Extensions["letterboxd"]["filmTitle"]),
		handleYear(entry.Extensions["letterboxd"]["filmYear"]),
		handleRating(entry.Extensions["letterboxd"]["memberRating"]),
		watchedDate,
		handleRewatch(entry.Extensions["letterboxd"]["rewatch"]),
		poster,
		review,
	}
}

func handleFilmTitle(title []ext.Extension) (string) {
	if len(title) == 0 {
		return ""
	}
	return title[0].Value
}

func handleYear(year []ext.Extension) (string) {
	if len(year) == 0 {
		return ""
	}
	return year[0].Value
}

func handleRating(rating []ext.Extension) (int) {
	if len(rating) == 0 {
		return -1
	}

	i, err := strconv.Atoi(string(rating[0].Value[0]))
	j, err := strconv.Atoi(string(rating[0].Value[2]))

	if err != nil {
		return -1
	}

	return (i * 10) + j
}

func handleWatchedDate(date []ext.Extension) (time.Time) {
	if len(date) == 0 {
		return time.Time{}
	}

	t, err := time.Parse("2006-01-02", date[0].Value)

	if err != nil {
		return time.Time{}
	}

	return t
}

func handleRewatch(rewatch []ext.Extension) (bool) {
	if len(rewatch) == 0 {
		return false
	}

	if rewatch[0].Value == "Yes" {
		return true
	}
	return false
}

func handleData(title, description string, watchedDate time.Time) (poster, review string) {
	watchedDateString := watchedDate.Format("Monday January 2, 2006")

	// Poster
	rePoster, _ := regexp.Compile(`^ <p><img src="https://a.ltrbxd.com/resized/(?P<image_url>.+?)"/></p> `)
	if img := rePoster.FindStringSubmatch(description); len(img) > 0 {
		poster = "https://a.ltrbxd.com/resized/" + img[1]
		description = rePoster.ReplaceAllString(description, " ") // Remove, preventing injection
	} else {
		poster = ""
	}

	// If spoiler, elseif no review, else review
	if r, _ := regexp.MatchString(` \(contains spoilers\)$`, title); r {
		review = "This review may contain spoilers."
	} else if r, _ := regexp.MatchString(fmt.Sprintf(` <p>Watched on %s.</p> `, watchedDateString), description); r {
		review = ""
	} else {
		reReview, _ := regexp.Compile(` <p>(?P<review>.+)</p> `)
		if rer := reReview.FindStringSubmatch(description); len(rer) > 0 {
			review = rer[1]
		} else {
			review = ""
		}
	}

	return poster, review
}
