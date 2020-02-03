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
	match.Difficulty = stats.Normal
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

func ParseMVP(match *stats.Match, content string) {
	lines := strings.Split(content, "\n")
	if len(lines) != 9 {
		log.Println("unexpected number of lines in MVP", len(lines))
		return
	}

	var team *stats.Team
	switch {
	case strings.Contains(lines[0], "NORTH"):
		team = match.North
	case strings.Contains(lines[0], "SOUTH"):
		team = match.South
	default:
		log.Println("unknown team name:", lines[0])
		return
	}

	for _, i := range []int{1, 4, 7} {
		mvp := new(stats.MVPplayer)
		_, err := fmt.Sscanf(lines[i], "MVP %s", &mvp.Title)
		if err != nil {
			log.Println("couldn't parse title:", lines[i], mvp.Title)
			continue
		}
		mvp.Title = strings.ReplaceAll(mvp.Title, ":", "")
		var format string
		switch mvp.Title {
		case "Defender":
			format = "%s with a score of %d"
		case "Builder":
			format = "%s built %d things"
		case "Deaths":
			format = "%s died %d times"
		default:
			log.Println("unrecognized title", mvp.Title)
			continue
		}
		_, err = fmt.Sscanf(lines[i+1], format, &mvp.Name, &mvp.Stat)
		if err != nil {
			log.Println("couldn't parse stats:", lines[i+1])
			continue
		}
		team.MVPs = append(team.MVPs, mvp)
	}
}

func ParseLineEmbed(match *stats.Match, line string, t time.Time) {
	if match.Start.IsZero() {
		match.Start = t
	}
	switch {
	case strings.HasPrefix(line, ">> Map difficulty has changed to"):
		var difficulty string

		var diffConst stats.Difficulty
		var err error
		if line == ">> Map difficulty has changed to Piece of cake difficulty!" {
			diffConst = stats.PieceOfCake
			goto done
		}
		_, err = fmt.Sscanf(line, ">> Map difficulty has changed to %s difficulty!", &difficulty)
		if err != nil {
			log.Println("failed to parse difficulty:", err)
			break
		}

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
	done:
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

func removeFromTeam(name string, team *stats.Team) bool {
	for i, p := range team.Players {
		if p.Name == name {
			team.Players[i] = team.Players[len(team.Players)-1]
			team.Players = team.Players[:len(team.Players)-1]
			return true
		}
	}
	return false
}

func processJoin(match *stats.Match, teamName string, player *stats.Player) {
	var team *stats.Team
	switch teamName {
	case "north":
		team = match.North
	case "south":
		team = match.South
	default:
		log.Println("unknown team name:", teamName)
		return
	}
	team.Players = append(team.Players, player)
}

func ParseLine(match *stats.Match, line string, t time.Time) {
	switch {
	case strings.HasSuffix(line, "has joined the game"):
		event := new(stats.Event)
		event.EventType = stats.JoinGame
		event.Timestamp = t
		_, err := fmt.Sscanf(line, "%s has joined the game", &event.Payload)
		if err != nil {
			log.Println("failed to parse game join:", err)
			break
		}
		match.Players = append(match.Players, &stats.Player{Name: event.Payload})
		match.Timeline = append(match.Timeline, event)
	case strings.HasSuffix(line, "has left the game"):
		event := new(stats.Event)
		event.EventType = stats.LeaveGame
		event.Timestamp = t
		_, err := fmt.Sscanf(line, "%s has left the game", &event.Payload)
		if err != nil {
			log.Println("failed to parse game leave:", err)
			break
		}
		match.Timeline = append(match.Timeline, event)
	case strings.Contains(line, "has joined team"):
		var teamName, name string
		_, err := fmt.Sscanf(line, "%s has joined team %s", &name, &teamName)
		if err != nil {
			log.Println("failed to parse join:", err)
			break
		}
		teamName = strings.ReplaceAll(teamName, "!", "")
		processJoin(match, teamName, &stats.Player{Name: name})
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.JoinTeam, Payload: line, Timestamp: t})
	case strings.Contains(line, " was killed "):
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.PlayerDeath, Payload: line, Timestamp: t})
	case strings.HasSuffix(line, "is spectating."):
		var name string
		_, err := fmt.Sscanf(line, "%s is spectating.", &name)
		if err != nil {
			log.Println("failed to parse spectate:", err)
			break
		}
		if !removeFromTeam(name, match.South) && !removeFromTeam(name, match.North) {
			log.Println("player not in team, but spectating:", name, match.North.Players, match.South.Players)
		}
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.LeaveTeam, Payload: line, Timestamp: t})
	case strings.HasSuffix(line, "is no longer spectating."):
		var teamName, name string
		_, err := fmt.Sscanf(line, "Team %s player %s is no longer spectating.", &teamName, &name)
		if err != nil {
			log.Println("failed to parse no spectate:", err)
			break
		}
		processJoin(match, teamName, &stats.Player{Name: name})
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.JoinTeam, Payload: line, Timestamp: t})
	case strings.Contains(line, " flasks of "):
		var teamName, name, scienceName string
		var amount int
		_, err := fmt.Sscanf(line, "%s fed %d flasks of %s science to team %s biters!", &name, &amount, &scienceName, &teamName)
		if err != nil {
			log.Println("failed to parse feeding:", err)
			break
		}
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.Feed, Payload: line, Timestamp: t})
	}
}
