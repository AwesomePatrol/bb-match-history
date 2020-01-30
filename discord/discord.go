package discord

import (
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
	currentMatch *stats.Match
	bot          *discordgo.Session
)
var validChannels = map[string]interface{}{"671815098427244567": nil, "493470400336887811": nil}

const comfylatronID = "493392617258876948"

func init() {
	currentMatch = parser.NewMatch()
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
func processMatchMessages(s *discordgo.Session, m *discordgo.Message, match *stats.Match) bool {
	t, _ := discordgo.SnowflakeTimestamp(m.ID)
	return _processMatchMessages(s, m, match, t)
}

func _processMatchMessages(s *discordgo.Session, m *discordgo.Message, match *stats.Match, t time.Time) bool {
	for _, line := range strings.Split(m.Content, "\n") {
		if len(line) < 4 {
			continue
		}
		// Process map restart
		if strings.Contains(line, "Map is restarting") {
			log.Println("GAME RESTART")
			parser.FixPlayers(match)

			// Server stops after match restart
			if match.End.IsZero() || match.Start.Before(match.End) {
				match.End = t
			}

			err := stats.InsertMatch(match)
			if err != nil {
				log.Println("will be skipped:", err)
			}

			log.Println(match)
			return true
		}

		// Process bold messages
		if strings.HasPrefix(line, "**") {
			line = strings.ReplaceAll(line, "\\", "")
			line := strings.Trim(line, "*")
			log.Println("parsing:", line)
			parser.ParseLine(match, line, t)
			return false
		}
	}

	// Process embed messages
	for _, e := range m.Embeds {
		line := e.Description
		line = strings.ReplaceAll(line, "\\", "")
		log.Println("parsing:", line)
		parser.ParseLineEmbed(match, line, t)
	}
	return false
}

func getRelevantHistory(s *discordgo.Session, chanID string, t time.Time) (lines []*discordgo.Message) {
	//TODO: filter by Author.ID
	ch, err := s.Channel(chanID)
	if err != nil {
		log.Println("failed to get chan info:", err)
		return
	}
	beforeID := ch.LastMessageID
	for {
		msgs, err := s.ChannelMessages(chanID, 64, beforeID, "", "")
		if err != nil {
			log.Println("failed to get messages from history:", err)
			return
		}
		if len(msgs) == 0 {
			log.Println("history: no more messages")
			break
		}
		if ts, _ := discordgo.SnowflakeTimestamp(msgs[0].ID); ts.Before(t) {
			log.Println("history: messages are too old")
			break
		} else {
			log.Println("oldest message:", ts)
		}
		beforeID = msgs[len(msgs)-1].ID

		for _, msg := range msgs { // From newest to oldest
			if msg.Author.ID == s.State.User.ID { // Ignore own
				continue
			}
			if strings.HasPrefix(msg.Content, "**") || len(msg.Embeds) > 0 {
				lines = append(lines, msg)
			}
		}
	}
	return
}

func parseHistory(s *discordgo.Session, chanID string, t time.Time) {
	lines := getRelevantHistory(s, chanID, t)
	historyMatch := parser.NewMatch()
	for i := len(lines) - 1; i >= 0; i-- { // switch order
		if processMatchMessages(s, lines[i], historyMatch) {
			historyMatch = parser.NewMatch()
		}
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
	if strings.HasPrefix(m.Content, "!parseHistory") {
		var str, chanID string
		_, err := fmt.Sscanf(m.Content, "!parseHistory %s %s", &str, &chanID)
		if err != nil {
			log.Println("parseHistory command failed: scan:", err)
			return
		}
		t, err := time.Parse("2006-01-02", str)
		if err != nil {
			log.Println("parseHistory command failed: timestamp:", err)
			return
		}
		sendReplyInDM(s, m.Author.ID, "parsing history from "+chanID+" started")
		go parseHistory(s, chanID, t)
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore own messages (shouldn't happen often anyway)
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Process only added channels
	if _, ok := validChannels[m.ChannelID]; ok && m.Author.ID == comfylatronID {
		log.Println(*m.Message, m.Author.ID)
		if processMatchMessages(s, m.Message, currentMatch) {
			currentMatch = parser.NewMatch()
		}
	}

	// Commands for master only
	if m.Author.ID == masterID {
		processMasterCommands(s, m)
	}
}
