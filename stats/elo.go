package stats

import (
	"log"
	"math"

	"gorm.io/gorm"
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
		d := calcChangeELO(float64(p.BeforeELO), avgOpponentELO, won)
		log.Printf("elo update: %20s %4d [%+2d]\n", p.Player.Name, p.BeforeELO, d)
		p.GainELO = d
	}
}

func (t *Team) setAvgELO() {
	sum := 0
	for _, p := range t.Players {
		if p.BeforeELO == 0 {
			p.BeforeELO = startELO
		}
		sum += p.BeforeELO
	}
	t.AvgELO = float64(sum) / float64(len(t.Players)) // smaller error
}

func (m *Match) UpdateMatchELO() {
	if len(m.South.Players) == 0 || len(m.North.Players) == 0 { // Don't update 1 v nothing
		return
	}
	m.North.setAvgELO()
	m.South.setAvgELO()
	m.North.updateTeamELO(m.South.AvgELO, m.Winner == North)
	m.South.updateTeamELO(m.North.AvgELO, m.Winner == South)
}

func FillPlayersWithELO(players []*GamePlayer) {
	for _, p := range players {
		if p.BeforeELO == 0 {
			qp, err := QueryPlayerByName(p.Player.Name)
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					continue
				}
				log.Println("failed to get player's elo:", err)
				continue
			}
			p.BeforeELO = qp.ELO
		}
	}
}
