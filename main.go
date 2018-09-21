package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bottleneckco/radio-clerk/commands"
	"github.com/bottleneckco/radio-clerk/util"
	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	dg, err := discordgo.New(fmt.Sprintf("Bot %s", os.Getenv("DISCORD_TOKEN")))
	if err != nil {
		log.Panic(err)
	}

	dg.AddHandler(func(s *discordgo.Session, m *discordgo.MessageCreate) {
		// Ignore all messages created by the bot itself
		if m.Author.ID == s.State.User.ID {
			return
		}

		log.Printf("[MESSAGE] '%s' - '%s'\n", m.Content, m.Author.Username)
		parts := strings.Split(m.Content, " ")

		if strings.HasPrefix(parts[0], os.Getenv("BOT_COMMAND_PREFIX")) {
			if handler, ok := commands.CommandsMap[parts[0][1:]]; ok {
				log.Printf("[COMMAND] Processing command '%s'\n", parts[0][1:])
				m.Content = strings.Join(parts[1:], " ")
				handler(s, m)
			}
		}
	})

	dg.AddHandler(func(s *discordgo.Session, vsu *discordgo.VoiceStateUpdate) {
		if commands.VoiceConnection == nil {
			return
		}
		channel, err := s.Channel(commands.VoiceConnection.ChannelID)
		if err != nil {
			log.Println(err)
			return
		}
		if channel.ID == commands.VoiceConnection.ChannelID && len(util.GetUsersInVoiceChannel(s, commands.VoiceConnection.ChannelID)) == 1 {
			// Only bot left
			log.Println("Leaving, only me left in voice channel.")
			commands.VoiceConnection.Disconnect()
			commands.VoiceConnection = nil
		}
	})

	commands.GameUpdateFunc = func(game string) {
		dg.UpdateStatus(0, game)
	}

	err = dg.Open()
	if err != nil {
		log.Panic(err)
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	if commands.VoiceConnection != nil {
		commands.VoiceConnection.Disconnect()
	}

	// Cleanly close down the Discord session.
	dg.Close()
}
