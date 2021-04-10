package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var db *DB

func main() {
	discordToken := os.Getenv("DISCORD_TOKEN")
	if discordToken == "" {
		log.Fatalln("No $DISCORD_TOKEN given.")
	}

	discord, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		log.Println("error creating Discord session,", err)
		return
	}
	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers
	discord.State.MaxMessageCount = 100
	discord.AddHandler(messageCreate)

	log.Println("Connecting to Discord")
	err = discord.Open()
	if err != nil {
		log.Fatalf("failed to open discord connection: %v\n", err)
		return
	}

	log.Println("Opening DB")
	db, err = OpenSQLDB("sqlite3", "fizzboxd.db")
	if err != nil {
		log.Fatalf("failed to open database: %v\n", err)
	}

	log.Println("Bot is now running. Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	log.Println("Closing Discord")
	discord.Close()

	log.Println("Closing DB")
	db.Close()

	log.Println("bye")
}
