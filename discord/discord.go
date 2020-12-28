package discord

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/awesomepatrol/bb-match-history/parser"
	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/bwmarrin/discordgo"
)

const masterID = "213745524279345152"

var (
	casualServer     = "781628243223642123"
	tournamentServer = ""
	testServer       = "671815098427244567"
	bot              *discordgo.Session
	currentMatch     = map[string]*stats.Match{
		testServer:   nil,
		casualServer: nil,
	}
	trustedBotIDs = map[string]string{
		"493392617258876948": "comfylatron",
		"785636002739388426": "factoriodiscordbot",
	}
	mux sync.Mutex
)

func init() {
	for key := range currentMatch {
		NewMatch(key)
	}
}

func _getCurrentMatch(id string) *stats.Match {
	mux.Lock()
	defer mux.Unlock()
	if id == "" {
		return nil
	}
	m := currentMatch[id]
	parser.FixPlayers(m)
	return m
}

func GetCurrentCasual() *stats.Match {
	return _getCurrentMatch(casualServer)
}

func GetCurrentTournament() *stats.Match {
	return _getCurrentMatch(tournamentServer)
}

// NewMatch initializes match structure and sets it
func NewMatch(chanID string) (m *stats.Match) {
	m = parser.NewMatch()
	m.ChannelID = chanID
	currentMatch[chanID] = m
	return
}

// OpenBot connects to discord with given token.
// It will panic if any error occurs.
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

	go parseCurrent(bot, casualServer, time.Now().AddDate(0, 0, -1))
}

// CloseBot disconnects from discord.
func CloseBot() {
	bot.Close()
}

// sendReplyInDM sends message directly instead of posting it to a channel.
func sendReplyInDM(s *discordgo.Session, recipientID string, content string) {
	log.Println("reply in DM:", content)
	if recipientID == "" {
		return
	}
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

// processMatchMessages retreives timestamp from message for _processMatchMessages.
func processMatchMessages(s *discordgo.Session, m *discordgo.Message, match *stats.Match, skip bool) bool {
	t, _ := discordgo.SnowflakeTimestamp(m.ID)
	return _processMatchMessages(s, m, match, t, skip)
}

// processMatchEnd fills missing values and fixes common issues in match structure.
// If skip is set to true, the match isn't put into the database.
func processMatchEnd(match *stats.Match, t time.Time, skip bool) bool {
	log.Println("GAME RESTART")
	parser.FixPlayers(match)

	// Server stops after match restart
	if match.End.IsZero() || match.Start.After(match.End) {
		match.End = t
	}

	log.Println(match)

	if skip {
		log.Println("do not insert a partial match")
		return true
	}
	err := stats.InsertMatch(match)
	if err != nil {
		log.Println("will be skipped:", err)
	}
	return true
}

// _processMatchMessages parses a single discord message m for match assuming time t.
// If skip is set to true, the match isn't put into the database.
func _processMatchMessages(s *discordgo.Session, m *discordgo.Message, match *stats.Match, t time.Time, skip bool) bool {
	for _, line := range strings.Split(m.Content, "\n") {
		// Skip extremely short lines
		if len(line) < 4 {
			continue
		}

		// Process map restart
		if strings.Contains(line, "Map is restarting") || strings.Contains(line, "Server is shutting down") {
			return processMatchEnd(match, t, skip)
		}

		// Process bold messages
		if strings.HasPrefix(line, "**") {
			line = strings.ReplaceAll(line, "\\", "")
			line = strings.Trim(line, "*")
			log.Println("parsing bold:", line)
			parser.ParseLine(match, line, t)
			return false
		}
	}

	// Process embed messages
	for _, e := range m.Embeds {
		line := e.Description
		line = strings.ReplaceAll(line, "\\", "")
		log.Println("parsing embed:", line)
		if strings.Contains(line, "MVP") {
			parser.ParseMVP(match, line)
		}
		if parser.ParseLineEmbed(match, line, t) {
			return processMatchEnd(match, t, skip)
		}
	}
	return false
}

// getRelevantHistory reads all messages in chanID that are newer than time t and returns them as
// lines. If current is set to true, it will stop on the first map restart.
func getRelevantHistory(s *discordgo.Session, chanID string, t time.Time, current bool) (lines []*discordgo.Message) {
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
			if current && strings.Contains(msg.Content, "Map is restarting") {
				return
			}
			lines = append(lines, msg)
		}
	}
	return
}

// parseHistory reads all messages in chanID that are newer than time t and then inserts all parsed
// matches in the database.
func parseHistory(s *discordgo.Session, chanID string, t time.Time) {
	lines := getRelevantHistory(s, chanID, t, false)
	historyMatch := parser.NewMatch()
	historyMatch.ChannelID = chanID
	skip := true
	for i := len(lines) - 1; i >= 0; i-- { // switch order
		if processMatchMessages(s, lines[i], historyMatch, skip) {
			skip = false
			historyMatch = parser.NewMatch()
			historyMatch.ChannelID = chanID
		}
	}
}

// parseCurrent reads messages in chanID until the first match start (but no older than time t) and
// then processes it into currentMatch.
func parseCurrent(s *discordgo.Session, chanID string, t time.Time) {
	lines := getRelevantHistory(s, chanID, t, true)
	mux.Lock()
	defer mux.Unlock()
	for i := len(lines) - 1; i >= 0; i-- { // switch order
		if processMatchMessages(s, lines[i], currentMatch[chanID], false) {
			log.Println("shouldn't have ended")
			NewMatch(chanID)
		}
	}
}

// scanChannels lists channels on all guilds and adds them to watched lists when they match a
// pattern. authorID is informed when a matching channels is found.
func scanChannels(s *discordgo.Session, authorID string) {
	for _, guild := range s.State.Guilds {
		channels, err := s.GuildChannels(guild.ID)
		if err != nil {
			log.Println("failed to get guild channels:", err)
			return
		}
		for _, c := range channels {
			if c.Type != discordgo.ChannelTypeGuildText {
				continue
			}
			if _, ok := currentMatch[c.ID]; ok {
				log.Println("channel already added:", c.ID)
				continue
			}
			if c.Name == "bb-server-chat" || (strings.HasPrefix(c.Name, "s") && strings.Contains(c.Name, "biter-battle")) {
				sendReplyInDM(s, authorID, "Adding channel "+c.ID+" with name: "+c.Name)
				NewMatch(c.ID)
				go parseCurrent(s, c.ID, time.Now().AddDate(0, 0, -1))
			}
		}
	}
	sendReplyInDM(s, authorID, "Done.")
}

// processMasterCommands acts on messages from masterID.
func processMasterCommands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Content == `!test` {
		sendReplyInDM(s, m.Author.ID, "ok")
	}
	if m.Content == `!scanChannels` {
		go scanChannels(s, m.Author.ID)
	}
	if m.Content == `!addChannel` {
		NewMatch(m.ChannelID)
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
	if strings.HasPrefix(m.Content, "!parseCurrent") {
		var str, chanID string
		_, err := fmt.Sscanf(m.Content, "!parseCurrent %s %s", &str, &chanID)
		if err != nil {
			log.Println("parseCurrent command failed: scan:", err)
			return
		}
		t, err := time.Parse("2006-01-02", str)
		if err != nil {
			log.Println("parseCurrent command failed: timestamp:", err)
			return
		}
		NewMatch(chanID)
		sendReplyInDM(s, m.Author.ID, "parsing current match from "+chanID+" started")
		go parseCurrent(s, chanID, t)
	}
}

// messageCreate is a handler for new messages that bot receives.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore own messages (shouldn't happen often anyway)
	if m.Author.ID == s.State.User.ID {
		return
	}

	mux.Lock()
	defer mux.Unlock()

	// Process only added channels
	if _, ok := currentMatch[m.ChannelID]; ok {
		// Process only trusted authors (bots)
		if _, ok := trustedBotIDs[m.Author.ID]; ok {
			log.Println(*m.Message, m.Author.ID)
			if processMatchMessages(s, m.Message, currentMatch[m.ChannelID], false) {
				// Match ended so a new one should be created
				NewMatch(m.ChannelID)
			}
		}
	}

	// Commands for master only
	if m.Author.ID == masterID {
		processMasterCommands(s, m)
	}
}
