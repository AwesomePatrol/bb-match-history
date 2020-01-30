package discord

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/awesomepatrol/bb-match-history/parser"
	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/bwmarrin/discordgo"
)

const masterID = "213745524279345152"

var (
	match *stats.Match
	bot   *discordgo.Session
)
var validChannels = map[string]interface{}{"671815098427244567": nil}

func init() {
	match = parser.NewMatch()
}

func OpenBot(token string) {
	var err error
	bot, err = discordgo.New("Bot " + token)
	if err != nil {
		panic(fmt.Errorf("error creating Discord session: %s", err))
	}

	bot.AddHandler(messageCreate)

	err = bot.Open()
	if err != nil {
		panic(fmt.Errorf("error opening connection: %s", err))
	}
}

func CloseBot() {
	bot.Close()
}

func sendReplyInDM(s *discordgo.Session, recipientID string, content string) {
	ch, err := s.UserChannelCreate(recipientID)
	if err != nil {
		log.Println(err)
		return
	}
	_, err = s.ChannelMessageSend(ch.ID, content)
	if err != nil {
		log.Println(err)
	}
}
func processMatchMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	t, _ := discordgo.SnowflakeTimestamp(m.ID)
	_processMatchMessages(s, m, t)
}

func _processMatchMessages(s *discordgo.Session, m *discordgo.MessageCreate, t time.Time) {
	// Process map restart
	if m.Content == `**\*\*\* Map is restarting! \*\*\***` {
		parser.FixPlayers(match)
		stats.InsertMatch(match)
		match.End = t

		ret, err := json.Marshal(match)
		match = parser.NewMatch()
		if err != nil {
			log.Println(err)
			return
		}
		match.Start = t
		log.Println(err)
		return
	}

	// Process bold messages
	if strings.HasPrefix(m.Content, "**") {
		line := strings.Trim(m.Content, "*")
		log.Println("parsing:", line)
		parser.ParseLine(match, line, t)
		return
	}

	// Process embed messages
	for _, e := range m.Embeds {
		line := e.Description
		log.Println("parsing:", line)
		parser.ParseLineEmbed(match, line, t)
	}
}

func parseHistory(s *discordgo.Session, chanID string, t time.Time) {
	log.Println("done")
}

func processMasterCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == `!test` {
		sendReplyInDM(s, m.Author.ID, "ok")
	}
	if m.Content == `!addChannel` {
		validChannels[m.ChannelID] = nil
		sendReplyInDM(s, m.Author.ID, "channel "+m.ChannelID+" added")
	}
	if strings.HasPrefix(m.Content, "!parseHistory") {
		var str string
		_, err := fmt.Sscanf(m.Content, "!parseHistory %s", &str)
		if err != nil {
			log.Println("parseHistory command failed: scan:", err)
			return
		}
		t, err := time.Parse("2006-01-02", str)
		if err != nil {
			log.Println("parseHistory command failed: timestamp:", err)
			return
		}
		sendReplyInDM(s, m.Author.ID, "parsing history from "+m.ChannelID+" started")
		go parseHistory(s, m.ChannelID, t)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore own messages (shouldn't happen often anyway)
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Process only added channels
	if _, ok := validChannels[m.ChannelID]; ok {
		log.Println(*m.Message)
		processMatchMessages(s, m)
	}

	// Commands for master only
	if m.Author.ID == masterID {
		processMasterCommands(s, m)
	}
}
