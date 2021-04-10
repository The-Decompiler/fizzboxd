package main

import (
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func CmdFollow(args []string, channel, guild string) (string, error) {
	if len(args) == 0 {
		return "Usage: `!follow <username>`", nil
	}

	username := strings.ToLower(args[0])

	exists, err := db.FollowExists(username, channel)
	if err != nil {
		return "", fmt.Errorf("failed to check if username '%s' exists in channel '%s': %v\n", username, channel, err)
	}

	if exists {
		return fmt.Sprintf("Already following %s in this channel.", username), nil
	}

	err = db.Follow(username, channel, guild)
	if err != nil {
		return "", fmt.Errorf("failed to follow username '%s' in channel '%s' in guild '%s': %v\n", username, channel, guild, err)
	}

	return fmt.Sprintf("Now following %s in this channel.", username), nil
}

func CmdUnfollow(args []string, channel string) (string, error) {
	if len(args) == 0 {
		return "Usage: `!unfollow <username>`", nil
	}

	username := strings.ToLower(args[0])

	exists, err := db.FollowExists(username, channel)
	if err != nil {
		return "", fmt.Errorf("failed to check if username '%s' exists in channel '%s': %v\n", username, channel, err)
	}

	if !exists {
		return fmt.Sprintf("Can't unfollow %s, username not in the list of followed users in this channel.", username), nil
	}

	err = db.Unfollow(username, channel)
	if err != nil {
		return "", fmt.Errorf("failed to unfollow username '%s' in channel '%s': %v\n", username, channel, err)
	}

	return fmt.Sprintf("%s is no longer being followed in this channel.", username), nil
}

func CmdFollowing(channel string) (string, error) {
	following, err := db.Following(channel)
	if err != nil {
		return "", fmt.Errorf("failed to get list of followed users for channel '%s': %v\n", channel, err)
	}

	if len(following) == 0 {
		return "Not following anyone in this channel.", nil
	}

	sort.Strings(following)
	usernames := strings.Join(following, ", ")

	return fmt.Sprintf("Following the following Letterboxd usernames in this channel: %s", usernames), nil
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Author.Bot {
		return
	}

	if len(m.Content) == 0 {
		return
	}

	if m.Content[0] != '!' {
		return
	}

	say := func(text string) {
		s.ChannelMessageSend(m.ChannelID, text)
	}

	msg := strings.Fields(m.Content)
	cmd := strings.ToLower(msg[0])
	args := msg[1:]

	var isAdmin bool
	perms, err := s.State.MessagePermissions(m.Message)
	if err != nil {
		log.Printf("failed to get message permissions: %v\n", err)
		isAdmin = false
	} else {
		isAdmin = perms&discordgo.PermissionAdministrator != 0
	}

	switch {
	case cmd == "!follow" && isAdmin:
		resp, err := CmdFollow(args, m.ChannelID, m.GuildID)

		if err != nil {
			log.Printf("failed to execute CmdFollow: %v\n", err)
		}

		if resp != "" {
			say(resp)
		}

	case cmd == "!following":
		resp, err := CmdFollowing(m.ChannelID)

		if err != nil {
			log.Printf("failed to execute CmdFollowing: %v\n", err)
		}

		if resp != "" {
			say(resp)
		}

	case cmd == "!help":
		help := `**!follow <username>** - follows a user in this channel
**!unfollow <username>** - unfollows a user in this channel
**!following** - shows the list of currently followed users in this channel
**!help** - shows this help message`
		say(help)

	case cmd == "!unfollow" && isAdmin:
		resp, err := CmdUnfollow(args, m.ChannelID)

		if err != nil {
			log.Printf("failed to execute CmdUnfollow: %v\n", err)
		}

		if resp != "" {
			say(resp)
		}

	}
}
