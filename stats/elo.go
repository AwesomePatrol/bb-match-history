package stats

import (
	"log"
	"math"

	"github.com/jinzhu/gorm"
)

const (
	k   float64 = 32
	dev float64 = 400

	startELO int = 800
)

func expectedScore(a, b float64) float64 {
	d := b - a
	if d > dev { // clamp d values
		d = dev
	} else if d < -dev {
		d = -dev
	}
	return 1 / (1 + math.Pow(10, d/dev))
}

func calcChangeELO(playerELO, avgOpponentELO float64, won bool) int {
	change := -expectedScore(playerELO, avgOpponentELO)
	if won {
		change += 1
	}
	return int(math.Round(change * k))
}

func (t *Team) updateTeamELO(avgOpponentELO float64, won bool) {
	for _, p := range t.Players {
		d := calcChangeELO(float64(p.ELO), avgOpponentELO, won)
		p.ELO += d
	}
}

func (t *Team) getAvgELO() float64 {
	sum := 0
	for _, p := range t.Players {
		if p.ELO == 0 {
			p.ELO = startELO
		}
		sum += p.ELO
	}
	return float64(sum) / float64(len(t.Players)) // smaller error
}

func (m *Match) UpdateMatchELO() {
	if len(m.South.Players) == 0 || len(m.North.Players) == 0 { // Don't update 1 v nothing
		return
	}
	southAvg, northAvg := m.South.getAvgELO(), m.North.getAvgELO()
	m.North.updateTeamELO(southAvg, m.NorthWon)
	m.South.updateTeamELO(northAvg, !m.NorthWon)
}

func FillPlayersWithELO(players []*Player) {
	for _, p := range players {
		if p.ELO == 0 {
			elo, err := QueryPlayerELO(p.Name)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				log.Println("failed to get player's elo:", err)
				continue
			}
			p.ELO = elo
		}
	}
}
