package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats"
)

func makeUnique(players []*stats.Player) []*stats.Player {
	m := make(map[string]struct{})
	for _, p := range players {
		m[p.Name] = struct{}{}
	}
	iter := 0
	for k := range m {
		players[iter].Name = k
		iter++
	}
	return players[:iter]
}

func NewMatch() (match *stats.Match) {
	match = new(stats.Match)
	match.South = new(stats.Team)
	match.North = new(stats.Team)
	match.South.IsNorth = false
	match.North.IsNorth = true
	match.Start = time.Now().UTC()
	return
}

func ParseSingleMatch(reader io.Reader) (*stats.Match, error) {
	scanner := bufio.NewScanner(reader)
	match := NewMatch()
	ongoing := true
	for scanner.Scan() && ongoing {
		line := scanner.Text()
		log.Println(line)
		// TODO: check if bold
		if line == "*** Map is restarting!  ***" {
			ongoing = false
			break
		}
		ParseLine(match, line, time.Now())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	FixPlayers(match)
	return match, nil
}

func FixPlayers(match *stats.Match) {
	match.Players = makeUnique(match.Players)
	match.North.Players = makeUnique(match.North.Players)
	match.South.Players = makeUnique(match.South.Players)
}

func ParseLineEmbed(match *stats.Match, line string, t time.Time) {
	if match.Start.IsZero() {
		match.Start = t
	}
	switch {
	case strings.HasPrefix(line, ">> Map difficulty has changed to"):
		var difficulty string

		_, err := fmt.Sscanf(line, ">> Map difficulty has changed to %s difficulty!", &difficulty)
		if err != nil {
			break
		}

		var diffConst stats.Difficulty
		switch difficulty {
		case "Peaceful":
			diffConst = stats.Peaceful
		case "Piece of cake":
			diffConst = stats.PieceOfCake
		case "Easy":
			diffConst = stats.Easy
		case "Normal":
			diffConst = stats.Normal
		case "Hard":
			diffConst = stats.Hard
		case "Nightmare":
			diffConst = stats.Nightmare
		case "Insane":
			diffConst = stats.Insane
		default:
			log.Println("unknown difficulty:", difficulty)
		}
		match.Difficulty = diffConst
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.DifficultyChange, Payload: line, Timestamp: t})
	case strings.HasPrefix(line, "Server has"):
		switch line {
		case "Server has **started**":
			match.Start = t
		case "Server has **stopped**":
			match.End = t
		}
	case strings.HasSuffix(line, "has won!"):
		switch line {
		case "Team South has won!":
			match.NorthWon = false
		case "Team North has won!":
			match.NorthWon = true
		default:
			log.Println("err: unknown team:", line)
		}
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.WinnerAnnounce, Payload: line, Timestamp: t})
	case strings.HasPrefix(line, "Time - "):
		var hours, minutes int
		_, err := fmt.Sscanf(line, "Time - %d hours and %d minutes", &hours, &minutes)
		if err != nil {
			hours = 0
			_, err := fmt.Sscanf(line, "Time - %d minutes", &minutes)
			if err != nil {
				break
			}
		}
		match.Length = time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes)
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.GameTimeAnnounce, Payload: line, Timestamp: t})
	}
}

func ParseLine(match *stats.Match, line string, t time.Time) {
	switch {
	case strings.HasSuffix(line, "has joined the game"):
		event := new(stats.Event)
		event.EventType = stats.JoinGame
		event.Timestamp = t
		_, err := fmt.Sscanf(line, "%s has joined the game", &event.Payload)
		if err != nil {
			break
		}
		match.Timeline = append(match.Timeline, event)
	case strings.HasSuffix(line, "has left the game"):
		event := new(stats.Event)
		event.EventType = stats.LeaveGame
		event.Timestamp = t
		_, err := fmt.Sscanf(line, "%s has left the game", &event.Payload)
		if err != nil {
			break
		}
		match.Timeline = append(match.Timeline, event)
	case strings.HasSuffix(line, "has joined team north!"):
		player := new(stats.Player)
		_, err := fmt.Sscanf(line, "%s has joined team north!", &player.Name)
		if err != nil {
			break
		}
		match.Players = append(match.Players, player)
		match.North.Players = append(match.North.Players, player)
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.JoinTeam, Payload: line, Timestamp: t})
	case strings.HasSuffix(line, "has joined team south!"):
		player := new(stats.Player)
		_, err := fmt.Sscanf(line, "%s has joined team south!", &player.Name)
		if err != nil {
			break
		}
		match.Players = append(match.Players, player)
		match.South.Players = append(match.South.Players, player)
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.JoinTeam, Payload: line, Timestamp: t})
	case strings.Contains(line, " was killed "):
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.PlayerDeath, Payload: line, Timestamp: t})
	}
}
