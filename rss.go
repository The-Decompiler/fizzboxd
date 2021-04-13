package main

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
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

func PostFeeds(db *DB, discord *discordgo.Session, p *bluemonday.Policy) error {
	users, err := db.GetFollows()
	if err != nil {
		return err
	}

	for username, follows := range users {
		feed, err := GetFeed(username, p)
		if err != nil {
			log.Printf("failed to get feed for username '%s': %v\n", username, err)
			continue
		}

		// Done this way so that not multiple requests are made to LB for
		// someone that is being followed in multiple channels.
		for _, f := range follows {
			filteredFeed := feed.FilterEntries(f.History, 4)
			if len(filteredFeed.Entries) == 0 {
				continue
			}
			embed := filteredFeed.GenerateEmbded()

			// To avoid spamming when first following someone
			if len(f.History) != 0 {
				if _, err := discord.ChannelMessageSendEmbed(f.Channel, embed); err != nil {
					log.Printf("failed to send embed message '%v': %v\n", *embed, err)
					continue
				}
			}

			if err := db.UpdateHistory(username, f.Channel, feed.GetHistory()); err != nil {
				log.Printf("failed to update history: %v\n", err)
				continue
			}
		}
	}

	return nil
}

func (f *Feed) GetHistory() []string {
	history := []string{}
	for _, e := range f.Entries {
		history = append(history, e.ID)
	}
	return history
}

// Keep n amount of entries excluding those in the history.
func (f Feed) FilterEntries(history []string, numOfEntries int) Feed {
	entries := []*FeedEntry{}
	count := 0
	for _, e := range f.Entries {
		if count >= numOfEntries {
			break
		}
		if stringInSlice(history, e.ID) {
			continue
		}
		entries = append(entries, e)
		count++
	}

	f.Entries = entries
	return f
}

func (f *Feed) GenerateEmbded() *discordgo.MessageEmbed {
	description := ""
	for _, e := range f.Entries {
		var url string
		if e.URL == "" {
			url = "https://letterboxd.com/"
		} else {
			url = e.URL
		}

		var watchedDate string
		if e.WatchedDate.IsZero() {
			watchedDate = ""
		} else {
			watchedDate = e.WatchedDate.Format("2006-01-02")
		}

		var rating string
		if e.Rating == -1 {
			rating = ""
		} else {
			rating = strings.Repeat("★", e.Rating/10)
			if e.Rating%10 == 5 {
				rating += "½"
			}
		}

		var rewatch string
		if e.Rewatch {
			rewatch = "↺"
		} else {
			rewatch = ""
		}

		var review string
		if e.Spoiler {
			review = "This review may contain spoilers."
		} else if len(e.Review) > 300 {
			review = e.Review[:300] + "..."
		} else {
			review = e.Review
		}

		if review != "" {
			review = fmt.Sprintf("```%s```", review)
		}

		description += fmt.Sprintf("**[%s (%s)](%s)**\n", e.Title, e.Year, url)
		description += fmt.Sprintf("**%s** %s %s\n", watchedDate, rating, rewatch)
		description += fmt.Sprintf("%s\n", review)
	}

	var poster string
	for _, e := range f.Entries {
		if e.Poster != "" {
			poster = e.Poster
			break
		}
	}

	embed := &discordgo.MessageEmbed{
		Author: &discordgo.MessageEmbedAuthor{
			URL:     fmt.Sprintf("https://letterboxd.com/%s/films/diary/", f.Username),
			Name:    f.DisplayName,
			IconURL: f.IconURL,
		},
		Color:       0xd8b437,
		Description: description,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: poster,
		},
	}

	return embed
}

// Fetches a user's RSS feed, returning an array of 50 FeedEntrys with parsed values
func GetFeed(username string, policy *bluemonday.Policy) (Feed, error) {
	var iconUrl = "https://cdn.discordapp.com/attachments/530814994204590097/794205173358395422/image0.png"

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
		entry, err := parseEntry(item, policy)
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

func parseEntry(entry *gofeed.Item, policy *bluemonday.Policy) (*FeedEntry, error) {
	watchedDate := handleWatchedDate(entry.Extensions["letterboxd"]["watchedDate"])
	poster, review, spoiler, err := HandleData(entry.Title, entry.Description, watchedDate, policy)

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

func HandleData(title, description string, watchedDate time.Time, policy *bluemonday.Policy) (poster, review string, spoiler bool, err error) {
	watchedDateString := watchedDate.Format("Monday January 2, 2006")

	// Poster
	rePoster, err := regexp.Compile(`^ <p><img src="https://a.ltrbxd.com/resized/(?P<image_url>.+?)"/></p> `)
	if err != nil {
		return "", "", false, err
	}

	if img := rePoster.FindStringSubmatch(description); len(img) > 0 {
		// Ensures that all images will only come from this link
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
		review = policy.Sanitize(description)
	}
	review = strings.TrimSpace(review)

	return poster, review, spoiler, nil
}
