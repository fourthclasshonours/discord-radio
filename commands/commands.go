package commands

import (
	"github.com/andersfylling/disgord"
	"github.com/bottleneckco/discord-radio/session"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

var (
	// PrimaryCommandMap a map of all the primary command handlers
	PrimaryCommandMap = make(map[string]func(disgord.Session, *disgord.MessageCreate))

	// SecondaryCommandMap a map of all the secondary command handlers
	SecondaryCommandMap = make(map[string]func(disgord.Session, *disgord.MessageCreate))

	// GuildSessionMap a map of all the guild sessions
	GuildSessionMap = make(map[disgord.Snowflake]*session.GuildSession)
)

func createGuildSession(guildID disgord.Snowflake, guildName string) session.GuildSession {
	return session.GuildSession{
		GuildID:   guildID,
		GuildName: guildName,
		RWMutex:   sync.RWMutex{},
		MusicPlayer: session.MusicPlayer{
			Control:       make(chan session.MusicPlayerAction),
			PlaybackState: session.PlaybackStateStopped,
		},
	}
}

func findOrCreateGuildSession(s disgord.Session, guildID disgord.Snowflake) *session.GuildSession {
	if session, ok := GuildSessionMap[guildID]; ok {
		return session
	}
	var guildName string
	guild, err := s.Guild(guildID).Get()
	if err == nil {
		guildName = guild.Name
	}
	session := createGuildSession(guildID, guildName)
	GuildSessionMap[guildID] = &session
	return &session
}

func init() {
	godotenv.Load()

	PrimaryCommandMap["ping"] = ping
	PrimaryCommandMap["q"] = queue
	PrimaryCommandMap["queue"] = queue
	PrimaryCommandMap["play"] = play
	PrimaryCommandMap["suicide"] = suicide
	PrimaryCommandMap["skip"] = skip
	PrimaryCommandMap["join"] = join
	PrimaryCommandMap["pause"] = pause
	PrimaryCommandMap["resume"] = resume
	PrimaryCommandMap["help"] = help
	PrimaryCommandMap["leave"] = leave
	PrimaryCommandMap["status"] = status

	SecondaryCommandMap["play"] = playSecondaryHandler
}

func deleteMessageDelayed(s disgord.Session, msg *disgord.Message) {
	time.Sleep(20 * time.Second)

	s.Channel(msg.ChannelID).DeleteMessages(&disgord.DeleteMessagesParams{
		Messages: []disgord.Snowflake{
			msg.ID,
		},
	})
}
