package stats

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func OpenDB(path string) {
	var err error
	db, err = gorm.Open(sqlite.Open(path), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		panic(fmt.Sprint("failed to connect database:", err))
	}
	db.AutoMigrate(&Player{})
	db.AutoMigrate(&Event{})
	db.AutoMigrate(&MVPplayer{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Channel{})
	db.AutoMigrate(&Match{})
	db.AutoMigrate(&PlayerMatch{})
}

func CloseDB() {
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("failed to get DB interface:", err)
	}
	// Might as weel try it.
	sqlDB.Close()
}

func InsertMatch(match *Match) error {
	if len(match.Players) == 0 {
		log.Println("skipping empty match:", match)
		return nil
	}
	var n int64
	db.Where("start = ?", match.Start).Find(new(Match)).Count(&n)
	if n > 0 {
		return fmt.Errorf("already in db")
	}

	// Calculate ELO and update player's ELO values in db.
	FillPlayersWithELO(match.Players)
	FillPlayersWithELO(match.North.Players)
	FillPlayersWithELO(match.South.Players)
	match.UpdateMatchELO()
	updateTeamELO(match.North)
	updateTeamELO(match.South)

	return db.Create(match).Error
}

func QueryGlobalMVP(title string) (mvp []MVPquery, err error) {
	err = db.Table("mv_pplayers").Where("title = ?", title).Select("name, count(name) as stat, sum(stat) as total").Group("name").Order("stat desc").Limit(10).Scan(&mvp).Error
	return
}

func queryMatchShort(id int) (match *Match, err error) {
	match = new(Match)
	db.Preload("Players").First(match, id)
	if db.Error != nil {
		return nil, db.Error
	}

	match.North = new(Team)
	err = db.Preload("Players").Where("is_north = ?", true).Where("match_id = ?", match.ID).First(match.North).Error
	if err != nil {
		return nil, err
	}

	match.South = new(Team)
	err = db.Preload("Players").Where("is_north = ?", false).Where("match_id = ?", match.ID).First(match.South).Error
	if err != nil {
		return nil, err
	}
	return
}

func QueryMatchShort(id int) (m *Match, err error) {
	m, err = queryMatchShort(id)
	if err != nil {
		return nil, err
	}
	m.Players = nil
	return
}

func QueryMatchLong(id int) (match *Match, err error) {
	match, err = queryMatchShort(id)
	if err != nil {
		return nil, err
	}

	err = db.Where("match_id = ?", id).Find(&match.Timeline).Error
	return
}

func QueryMatchAll() (matches []Match, err error) {
	// TODO pagination?
	matches = make([]Match, 0, 128)
	err = db.Limit(128).Order("start desc").Find(&matches).Error
	return
}

func QueryPlayerByName(name string) (p *Player, err error) {
	p = new(Player)
	err = db.Where("name = ?", name).First(p).Error
	return p, err
}

func QueryTopPlayersByELO(limit int) (p []*Player, err error) {
	err = db.Order("elo desc").Limit(limit).Find(&p).Error
	return
}

func countScienceInEvents(timeline []*Event, name string) (cnt int) {
	// FIXME extremely inefficient, but should be rarely used
	for _, e := range timeline {
		if e.EventType == Feed && strings.Contains(e.Payload, name) {
			s := strings.Split(e.Payload, " ")
			if len(s) == 11 {
				i, err := strconv.Atoi(s[2])
				if err != nil {
					log.Println("failed to parse falsk count:", e.Payload)
					continue
				}
				cnt += i
			}
		}
	}
	return
}

func intSliceToStringSlice(a []int) (b []string) {
	b = make([]string, len(a))
	for i := range a {
		b[i] = strconv.Itoa(a[i])
	}
	return
}

func GetMatchWithFeedsAsCSV(writer io.Writer) (err error) {
	w := csv.NewWriter(writer)
	matches := make([]Match, 0, 128)
	err = db.Preload("Players").Order("id asc").Find(&matches).Error
	if err != nil {
		return
	}

	// header
	record := make([]string, 12)
	record = []string{"id", "len", "diff", "player_cnt", "north_won",
		"feed_red", "feed_green", "feed_grey", "feed_blue", "feed_yellow", "feed_purple", "feed_white"}
	err = w.Write(record)
	if err != nil {
		return
	}

	for _, m := range matches {
		timeline := make([]*Event, 0, 128)
		err = db.Where("match_id = ?", m.ID).Find(&timeline).Error
		if err != nil {
			return
		}

		record = intSliceToStringSlice([]int{
			int(m.ID), int(m.Length), int(m.Difficulty), len(m.Players), int(m.Winner),
			countScienceInEvents(timeline, "automation"),
			countScienceInEvents(timeline, "logistic"),
			countScienceInEvents(timeline, "military"),
			countScienceInEvents(timeline, "chemical"),
			countScienceInEvents(timeline, "production"),
			countScienceInEvents(timeline, "utility"),
			countScienceInEvents(timeline, "space"),
		})
		err = w.Write(record)
		if err != nil {
			return
		}
	}
	w.Flush()
	return w.Error()
}

func updatePlayerELO(p *Player) (err error) {
	err = db.Model(&Player{}).Where("name = ?", p.Name).Update("ELO", p.ELO).Error
	return
}

func updateTeamELO(t *Team) (err error) {
	for _, p := range t.Players {
		err = updatePlayerELO(p)
		if err != nil {
			log.Println("failed to update ELO for", p, ":", err)
			return
		}
	}
	return nil
}

func ShouldUpdateELO() (update bool, err error) {
	p := &Player{}
	err = db.Order("ELO desc").First(p).Error
	if err != nil {
		return
	}
	return p.ELO == 0, nil
}

func ResetELO() (err error) {
	err = db.Model(&Player{}).Update("elo", 0).Error
	return
}

// UpdateELO iterates over all matches (from oldest to newest) and updates players' ELO.
func UpdateELO() (err error) {
	matches := make([]Match, 0, 128)
	err = db.Order("start asc").Find(&matches).Error
	if err != nil {
		return
	}
	for _, mi := range matches {
		var m *Match
		// Team's players need to be queried here so that ELO values will be updated.
		m, err = queryMatchShort(int(mi.ID))
		if err != nil {
			return
		}
		m.UpdateMatchELO()
		log.Println("elo update for:", m.ID)
		updateTeamELO(m.North)
		updateTeamELO(m.South)
	}
	return nil
}

func (team *Team) IsPlayerInTeam(name string) bool {
	for _, p := range team.Players {
		if p.Name == name {
			return true
		}
	}
	return false
}

func QueryPlayerMatches(name string) ([]Match, error) {
	player := Player{Name: name}
	err := db.Preload("History.Match").Preload("History", func(db *gorm.DB) *gorm.DB {
		return db.Order("matches.start DESC")
	}).First(&player).Error
	if err != nil {
		return nil, err
	}

	for _, match := range player.History {
		var teamW Team
		err = db.Preload("Players").Where("is_north = ?", match.Winner == North).Where("match_id = ?", match.ID).First(&teamW).Error
		if err != nil {
			return nil, err
		}
		var teamL Team
		err = db.Preload("Players").Where("is_north = ?", match.Winner == South).Where("match_id = ?", match.ID).First(&teamL).Error
		if err != nil {
			return nil, err
		}
		// TODO: convert to PlayerMatch
	}
	return player.History, nil
}
