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

func ParseSingleMatch(reader io.Reader) (*stats.Match, error) {
	scanner := bufio.NewScanner(reader)
	match := new(stats.Match)
	ongoing := true
	for scanner.Scan() && ongoing {
		line := scanner.Text()
		log.Println(line)
		// TODO: check if bold
		switch {
		case strings.HasPrefix(line, "Status:"):
		case strings.HasSuffix(line, "has joined the game"):
			event := new(stats.Event)
			event.EventType = stats.JoinGame
			_, err := fmt.Sscanf(line, "%s has joined the game", &event.Payload)
			if err != nil {
				break
			}
			match.Timeline = append(match.Timeline, event)
		case strings.HasSuffix(line, "has left the game"):
			event := new(stats.Event)
			event.EventType = stats.LeaveGame
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
			match.North.Players = append(match.North.Players, player)
			match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.JoinTeam, Payload: line})
		case strings.HasSuffix(line, "has joined team south!"):
			player := new(stats.Player)
			_, err := fmt.Sscanf(line, "%s has joined team south!", &player.Name)
			if err != nil {
				break
			}
			match.South.Players = append(match.South.Players, player)
			match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.JoinTeam, Payload: line})
		case strings.HasSuffix(line, "has won!"):
			switch line {
			case "Team South has won!":
				match.NorthWon = false
			case "Team North has won!":
				match.NorthWon = true
			default:
				log.Println("err: unknown team:", line)
			}
			match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.WinnerAnnounce, Payload: line})
		case strings.Contains(line, " was killed "):
			match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.PlayerDeath, Payload: line})
		case strings.HasPrefix(line, "Time - "):
			var hours, minutes int
			_, err := fmt.Sscanf(line, "Time - %d hours and %d minutes", &hours, &minutes)
			if err != nil {
				break
			}
			match.Length = time.Hour*time.Duration(hours) + time.Minute*time.Duration(minutes)
			match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.GameTimeAnnounce, Payload: line})
		case line == "*** Map is restarting!  ***":
			ongoing = false
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return match, nil
}
