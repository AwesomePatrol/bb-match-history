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

var (
	match *stats.Match
	bot   *discordgo.Session
)

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

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == s.State.User.ID {
		return
	}

	log.Println(m.Content)

	if m.Content == `**\*\*\* Map is restarting! \*\*\***` {
		parser.FixPlayers(match)
		stats.InsertMatch(match)

		ret, err := json.Marshal(match)
		if err != nil {
			log.Println(err)
		}
		_, err = s.ChannelMessageSend(m.ChannelID, string(ret))
		if err != nil {
			log.Println(err)
		}
		match = parser.NewMatch()
		return
	}

	if strings.HasPrefix(m.Content, "**") {
		line := strings.Trim(m.Content, "*")
		log.Println("parsing:", line)
		parser.ParseLine(match, line)
	}
}
