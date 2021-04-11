package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
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
	Spoiler     bool
}

var iconUrl = "https://cdn.discordapp.com/attachments/530814994204590097/794205173358395422/image0.png"

// Fetches a user's RSS feed, returning an array of 50 FeedEntrys with parsed values
func GetFeed(username string) (userFeed Feed, error error) {
	fp := gofeed.NewParser()
	feed, err := fp.ParseURL("https://letterboxd.com/" + username + "/rss/")

	if err != nil {
		return Feed{}, fmt.Errorf("failed to fetch feed: %v\n", err)
	}

	entries := []*FeedEntry{}
	for _, item := range feed.Items {
		if strings.HasPrefix(item.GUID, "letterboxd-list-") {
			continue
		}
		entry, err := parseEntry(item)
		if err != nil {
			continue
		}
		entries = append(entries, entry)
	}

	return Feed{
		Username:    username,
		DisplayName: handleDisplayName(feed.Title),
		IconURL:     iconUrl,
		Entries:     entries,
	}, nil
}

func parseEntry(entry *gofeed.Item) (*FeedEntry, error) {
	watchedDate := handleWatchedDate(entry.Extensions["letterboxd"]["watchedDate"])
	poster, review, spoiler, err := HandleData(entry.Title, entry.Description, watchedDate)

	if err != nil {
		return &FeedEntry{}, err
	}

	return &FeedEntry{
		ID:          entry.GUID,
		URL:         entry.Link,
		Title:       handleFilmTitle(entry.Extensions["letterboxd"]["filmTitle"]),
		Year:        handleYear(entry.Extensions["letterboxd"]["filmYear"]),
		Rating:      handleRating(entry.Extensions["letterboxd"]["memberRating"]),
		WatchedDate: watchedDate,
		Rewatch:     handleRewatch(entry.Extensions["letterboxd"]["rewatch"]),
		Poster:      poster,
		Review:      review,
		Spoiler:     spoiler,
	}, nil
}

// fullTitle expected to return `Letterboxd - DisplayName`
func handleDisplayName(fullTitle string) string {
	// feed.Items[0].Author.Name not used in case of empty feed
	displayName := strings.Replace(fullTitle, "Letterboxd - ", " ", 1)
	return displayName
}

func handleFilmTitle(title []ext.Extension) string {
	if len(title) == 0 {
		return ""
	}
	return title[0].Value
}

func handleYear(year []ext.Extension) string {
	if len(year) == 0 {
		return ""
	}
	return year[0].Value
}

func handleRating(rating []ext.Extension) int {
	if len(rating) == 0 {
		return -1
	}

	i, err := strconv.Atoi(string(rating[0].Value[0]))
	if err != nil {
		return -1
	}

	j, err := strconv.Atoi(string(rating[0].Value[2]))
	if err != nil {
		return -1
	}

	return (i * 10) + j
}

func handleWatchedDate(date []ext.Extension) time.Time {
	if len(date) == 0 {
		return time.Time{}
	}

	t, err := time.Parse("2006-01-02", date[0].Value)

	if err != nil {
		return time.Time{}
	}

	return t
}

func handleRewatch(rewatch []ext.Extension) bool {
	if len(rewatch) == 0 {
		return false
	}

	if rewatch[0].Value == "Yes" {
		return true
	}
	return false
}

func HandleData(title, description string, watchedDate time.Time) (poster, review string, spoiler bool, err error) {
	watchedDateString := watchedDate.Format("Monday January 2, 2006")

	// Poster
	rePoster, err := regexp.Compile(`^ <p><img src="https://a.ltrbxd.com/resized/(?P<image_url>.+?)"/></p> `)
	if err != nil {
		return "", "", false, err
	}

	if img := rePoster.FindStringSubmatch(description); len(img) > 0 {
		poster = "https://a.ltrbxd.com/resized/" + img[1]
		description = rePoster.ReplaceAllString(description, " ") // Remove, preventing injection
	} else {
		poster = ""
	}

	// Is spoiler
	if strings.HasSuffix(title, " (contains spoilers)") {
		description = strings.Replace(description, " <p><em>This review may contain spoilers.</em></p> ", " ", 1)
		spoiler = true
	}

	if description == fmt.Sprintf(" <p>Watched on %s.</p> ", watchedDateString) {
		review = ""
		spoiler = false
	} else {
		stripHTMLTags := bluemonday.StripTagsPolicy().AddSpaceWhenStrippingTag(true)
		review = stripHTMLTags.Sanitize(description)
	}
	review = strings.TrimSpace(review)

	return poster, review, spoiler, nil
}
