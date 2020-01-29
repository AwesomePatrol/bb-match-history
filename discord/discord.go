package discord

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/awesomepatrol/bb-match-history/parser"
	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/bwmarrin/discordgo"
)

const masterID = "213745524279345152"

var (
	match         *stats.Match
	bot           *discordgo.Session
	validChannels map[string]interface{}
)

func init() {
	match = parser.NewMatch()
	validChannels = make(map[string]interface{})
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
	// Process map restart
	if m.Content == `**\*\*\* Map is restarting! \*\*\***` {
		parser.FixPlayers(match)
		stats.InsertMatch(match)
		match = parser.NewMatch()

		ret, err := json.Marshal(match)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = s.ChannelMessageSend(m.ChannelID, string(ret))
		if err != nil {
			log.Println(err)
		}
		return
	}

	// Process bold messages
	if strings.HasPrefix(m.Content, "**") {
		line := strings.Trim(m.Content, "*")
		log.Println("parsing:", line)
		parser.ParseLine(match, line)
		return
	}

	// Process embed messages
	for _, e := range m.Embeds {
		line := e.Description
		log.Println("parsing:", line)
		parser.ParseLine(match, line)
	}
}

func processMasterCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == `!test` {
		sendReplyInDM(s, m.Author.ID, "ok")
	}
	if m.Content == `!addChannel` {
		validChannels[m.ChannelID] = nil
		sendReplyInDM(s, m.Author.ID, "channel "+m.ChannelID+" added")
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
