package session

import (
	"fmt"
	"github.com/andersfylling/disgord"
	"github.com/bottleneckco/discord-radio/models"
	"github.com/bottleneckco/discord-radio/youtube"
	"log"
	"os"
	"sync"
	"time"
	"unicode"

	"github.com/chrisport/go-lang-detector/langdet"
)

// GuildSession represents a guild voice session
type GuildSession struct {
	GuildID         disgord.Snowflake
	GuildName       string
	RWMutex         sync.RWMutex
	Queue           []models.QueueItem // current item = index 0
	VoiceConnection *disgord.VoiceConnection
	VoiceChannelID  disgord.Snowflake
	History         []string // Youtube IDs
	MusicPlayer     MusicPlayer
}

var (
	isAutoPlaylistEnabled = len(os.Getenv("BOT_AUTO_PLAYLIST")) > 0
)

// Loop session management loop
func (guildSession *GuildSession) Loop() {
	var err error
	go guildSession.OpusLoop()

	for {
		if guildSession.VoiceConnection == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		if guildSession.MusicPlayer.PlaybackState != PlaybackStateStopped {
			log.Println("[SCP] currently playing something!")
			time.Sleep(1 * time.Second)
			continue
		}
		if len(guildSession.Queue) == 0 && !isAutoPlaylistEnabled {
			log.Println("[SCP] no items in queue")
			time.Sleep(1 * time.Second)
			continue
		}
		if len(guildSession.Queue) == 0 {
			log.Println("[SCP] Getting from auto playlist")
			playlistItem, err := youtube.GenerateAutoPlaylistQueueItem(guildSession.History)
			if err != nil {
				log.Printf("[SCP] Error generating auto playlist item: %s\n", err)
				time.Sleep(1 * time.Second)
				continue
			}

			// Clear history automatically
			if len(guildSession.History) >= youtube.GetAutoPlaylistCacheLength() {
				guildSession.History = make([]string, 0)
			}

			queueItem := models.ConvertYouTubePlaylistItem(playlistItem)
			guildSession.RWMutex.Lock()
			guildSession.Queue = append(guildSession.Queue, queueItem)
			guildSession.RWMutex.Unlock()
		}
		guildSession.RWMutex.RLock()
		var song = guildSession.Queue[0]
		guildSession.RWMutex.RUnlock()

		// Announce music title
		// songTitle := util.SanitiseSongTitleTTS(song.Title)

		detector := langdet.NewDetector()
		clc := langdet.UnicodeRangeLanguageComparator{
			Name:       "zh-TW",
			RangeTable: unicode.Han,
		}
		jlc := langdet.UnicodeRangeLanguageComparator{
			Name:       "ja",
			RangeTable: unicode.Katakana,
		}
		klc := langdet.UnicodeRangeLanguageComparator{
			Name:       "ko",
			RangeTable: unicode.Hangul,
		}
		eng := langdet.UnicodeRangeLanguageComparator{
			Name:       "en",
			RangeTable: unicode.ASCII_Hex_Digit,
		}
		detector.AddLanguageComparators(&clc, &jlc, &klc, &eng)

		// if ttsMsgURL, err := googletts.GetTTSURL(fmt.Sprintf("Music: %s", songTitle), detector.GetLanguages(songTitle)[0].Name); err == nil {
		// 	log.Printf("[PLAYER] Announcing upcoming song title: '%s'\n", songTitle)

		// 	err = guildSession.MusicPlayer.PlayURL(ttsMsgURL)
		// 	if err != nil {
		// 		log.Println("Playback error", err)
		// 	}
		// }
		log.Printf("[PLAYER] Playing '%s'\n", song.Title)

		guildSession.History = append(guildSession.History, song.VideoID)

		var voiceConnection = *guildSession.VoiceConnection

		err = voiceConnection.StartSpeaking()
		if err != nil {
			log.Println(err)
			return
		}

		// NOTE: Only YouTube is supported for now
		var err = guildSession.MusicPlayer.PlayYouTubeVideo(fmt.Sprintf("https://www.youtube.com/watch?v=%s", song.VideoID))
		if err != nil {
			log.Println("Playback error", err)
		}

		err = voiceConnection.StopSpeaking()
		if err != nil {
			log.Println(err)
			return
		}

		guildSession.RWMutex.Lock()
		if len(guildSession.Queue) > 0 {
			guildSession.Queue = guildSession.Queue[1:]
		}
		guildSession.RWMutex.Unlock()
	}
}

func (guildSession *GuildSession) OpusLoop() {
	var err error

	for {
		if guildSession.VoiceConnection == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		var voiceConnection = *guildSession.VoiceConnection

		var bts = <-guildSession.MusicPlayer.PlaybackChannel

		err = voiceConnection.SendOpusFrame(bts)
		if err != nil {
			log.Println(err)
		}
	}

}
