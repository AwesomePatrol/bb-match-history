package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats"
	"github.com/awesomepatrol/bb-match-history/stats/const/difficulty"
	"github.com/awesomepatrol/bb-match-history/stats/const/science"
)

func NewMatch() (match *stats.Match) {
	match = new(stats.Match)
	match.South = new(stats.Team)
	match.North = new(stats.Team)
	match.South.IsNorth = false
	match.North.IsNorth = true
	match.South.TotalFeed = make([]int32, 7)
	match.North.TotalFeed = make([]int32, 7)
	match.Start = time.Now().UTC()
	match.Difficulty = difficulty.Normal
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
	return match, nil
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

func fixStartEnd(match *stats.Match, t time.Time) {
	if match.Start.IsZero() || match.Start.After(t) {
		match.Start = t
	}
	if match.End.IsZero() || match.End.Before(t) {
		match.End = t
	}
}

func ParseLineEmbed(match *stats.Match, line string, t time.Time) bool {
	fixStartEnd(match, t)
	switch {
	case strings.HasPrefix(line, ">> Map difficulty has changed to"):
		difficultyStr := strings.Replace(
			strings.Replace(line, ">> Map difficulty has changed to ", "", -1),
			" difficulty!", "", -1)

		// TODO: keep it in a map?
		diffConst, err := difficulty.StringToDifficulty(difficultyStr)
		if err != nil {
			log.Println("failed to convert:", difficultyStr, err)
			return false
		}
		match.Difficulty = diffConst
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.DifficultyChange, Payload: line, Timestamp: t})
	case strings.HasPrefix(line, "Server has"):
		switch line {
		case "Server has **started**":
			match.Start = t
		case "Server has **stopped**":
			match.End = t
			return true
		}
	case strings.Contains(line, "has won"):
		switch {
		case strings.Contains(line, "South has won"):
			match.Winner = stats.South
		case strings.Contains(line, "North has won"):
			match.Winner = stats.North
		default:
			log.Println("err: unknown team:", line)
		}
		match.Timeline = append(match.Timeline, &stats.Event{EventType: stats.WinnerAnnounce, Payload: line, Timestamp: t})
		return true
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
	case strings.Contains(line, "Evo:"):
		var teamName string
		var evo float32
		_, err := fmt.Sscanf(line, "%s Evo: %f%%", &teamName, &evo)
		if err != nil {
			log.Println("failed to parse team's final evo:", err)
			break
		}
		team, _ := getTeam(match, teamName)
		if team == nil {
			break
		}
		team.FinalEVO = evo
	case strings.Contains(line, "Threat:"):
		var teamName string
		var threat int
		_, err := fmt.Sscanf(line, "%s Threat: %d", &teamName, &threat)
		if err != nil {
			log.Println("failed to parse team's final threat:", err)
			break
		}
		team, _ := getTeam(match, teamName)
		if team == nil {
			break
		}
		team.FinalThreat = threat
	}
	return false
}

func removeFromTeam(name string, team *stats.Team) bool {
	for i, p := range team.Players {
		if p.Player.Name == name {
			team.Players[i] = team.Players[len(team.Players)-1]
			team.Players = team.Players[:len(team.Players)-1]
			return true
		}
	}
	return false
}

func findInTeam(players []*stats.GamePlayer, name string) *stats.GamePlayer {
	for _, p := range players {
		if p.Player.Name == name {
			return p
		}
	}
	return nil
}

func getTeam(match *stats.Match, name string) (*stats.Team, stats.Force) {
	switch name {
	case "North":
		fallthrough
	case "north":
		return match.North, stats.North
	case "South":
		fallthrough
	case "south":
		return match.South, stats.South
	}
	log.Println("unknown team name:", name)
	return nil, stats.Unknown
}

func processJoin(match *stats.Match, teamName string, player *stats.GamePlayer) {
	team, force := getTeam(match, teamName)
	if team == nil {
		return
	}
	if player.Force == force || findInTeam(team.Players, player.Player.Name) != nil {
		log.Println("ignoring join, player already linked to team:", player.Player.Name)
		return
	}

	// Make sure that player isn't on the opposite side already.
	switch force {
	case stats.North:
		if removeFromTeam(player.Player.Name, match.South) {
			log.Println("north player was on south", player)
		}
	case stats.South:
		if removeFromTeam(player.Player.Name, match.North) {
			log.Println("south player was on north", player)
		}
	}

	// Update force and add to the list
	player.Force = force
	team.Players = append(team.Players, player)
}

func ParseLine(match *stats.Match, line string, t time.Time) {
	fixStartEnd(match, t)

	var eventType stats.EventType
	switch {
	case strings.HasSuffix(line, "has joined the game"):
		// Not posted to discord on Raven's server
		log.Println("received unexpected join message:", line)
	case strings.HasSuffix(line, "has left the game"):
		// Not posted to discord on Raven's server
		log.Println("received unexpected leave message:", line)
	case strings.Contains(line, "has joined team"):
		var teamName, name string
		_, err := fmt.Sscanf(line, "%s has joined team %s", &name, &teamName)
		if err != nil {
			log.Println("failed to parse join:", err)
			return
		}
		teamName = strings.ReplaceAll(teamName, "!", "")

		p := findInTeam(match.Players, name)
		if p == nil {
			p = &stats.GamePlayer{Player: &stats.Player{Name: name}}
			match.Players = append(match.Players, p)
		}
		processJoin(match, teamName, p)
		eventType = stats.JoinTeam
	case strings.Contains(line, " was killed "):
		eventType = stats.PlayerDeath
	case strings.HasSuffix(line, "is spectating."):
		var name string
		_, err := fmt.Sscanf(line, "%s is spectating.", &name)
		if err != nil {
			log.Println("failed to parse spectate:", err)
			return
		}
		p := findInTeam(match.Players, name)
		if p == nil {
			log.Println("player not found")
			// TODO: add
			return
		}
		p.Force = stats.Spectator
		if !removeFromTeam(name, match.South) && !removeFromTeam(name, match.North) {
			log.Println("player not in team, but spectating:", name, match.North.Players, match.South.Players)
		}
		eventType = stats.LeaveTeam
	case strings.HasSuffix(line, "is no longer spectating."):
		var teamName, name string
		_, err := fmt.Sscanf(line, "Team %s player %s is no longer spectating.", &teamName, &name)
		if err != nil {
			log.Println("failed to parse no spectate:", err)
			return
		}
		p := findInTeam(match.Players, name)
		if p == nil {
			log.Println("player not found")
			// TODO: add
			return
		}
		processJoin(match, teamName, p)
		eventType = stats.JoinTeam
	case strings.Contains(line, " flasks of "):
		var teamName, name, scienceName string
		var amount int32
		_, err := fmt.Sscanf(line, "%s fed %d flasks of %s science to team %s biters!", &name, &amount, &scienceName, &teamName)
		if err != nil {
			log.Println("failed to parse feeding:", err)
			return
		}
		sc := science.NameToScience(scienceName)
		if sc == science.Unknown {
			log.Println("unknown science type")
			return
		}
		switch teamName {
		case "north":
			match.South.TotalFeed[int(sc)-1] += amount
		case "south":
			match.North.TotalFeed[int(sc)-1] += amount
		}
		eventType = stats.Feed
	default:
		// if no event detected, don't add it to the timeline
		goto no_save
	}
	match.Timeline = append(match.Timeline, &stats.Event{EventType: eventType, Payload: line, Timestamp: t})
no_save:
}
