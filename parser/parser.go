package parser

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/awesomepatrol/bb-match-history/stats"
)

func ParseSingleMatch(reader io.Reader) (*stats.Match, error) {
	scanner := bufio.NewScanner(reader)
	match := new(stats.Match)
	for scanner.Scan() {
		line := scanner.Text()
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
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return match, nil
}
