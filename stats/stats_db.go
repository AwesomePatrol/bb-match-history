package stats

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"strconv"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	db.AutoMigrate(&GamePlayer{})
	db.AutoMigrate(&Event{})
	db.AutoMigrate(&MVPplayer{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Channel{})
	db.AutoMigrate(&Match{})
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
	// TODO: use a transaction?
	var n int64
	db.Where("start = ?", match.Start).Find(new(Match)).Count(&n)
	if n > 0 {
		return fmt.Errorf("already in db")
	}

	// Calculate ELO and update player's ELO values in db.
	FillPlayersWithELO(match.Players) // Assume all players are present in this slice.
	match.UpdateMatchELO()

	// Insert
	err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(match).Error
	if err != nil {
		return err
	}
	// Update Players' ELO
	for _, p := range match.Players {
		err = db.Model(p.Player).Update("elo", p.Player.ELO).Error
		if err != nil {
			log.Println("failed to update player's elo:", err)
			continue
		}
	}
	return nil
}

func QueryGlobalMVP(title string) (mvp []MVPquery, err error) {
	err = db.Table("mv_pplayers").Where("title = ?", title).Select("name, count(name) as stat, sum(stat) as total").Group("name").Order("stat desc").Limit(10).Scan(&mvp).Error
	return
}

func queryMatchShort(id int) (match *Match, err error) {
	match = new(Match)
	err = db.Preload("Players").First(match, id).Error
	if err != nil {
		return
	}

	match.North = new(Team)
	err = db.Where("is_north = ?", true).Where("match_id = ?", match.ID).First(match.North).Error
	if err != nil {
		return nil, err
	}
	err = db.Preload("Player").Where("match_id = ?", match.ID).Where("force = ?", North).Find(&match.North.Players).Error
	if err != nil {
		return nil, err
	}

	match.South = new(Team)
	err = db.Where("is_north = ?", false).Where("match_id = ?", match.ID).First(match.South).Error
	if err != nil {
		return nil, err
	}
	err = db.Preload("Player").Where("match_id = ?", match.ID).Where("force = ?", South).Find(&match.South.Players).Error
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

func intSliceToStringSlice(a []int32) (b []string) {
	b = make([]string, len(a))
	for i := range a {
		b[i] = strconv.Itoa(int(a[i]))
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
	record := []string{"id", "len", "diff", "player_cnt", "north_won",
		"north_feed_red", "north_feed_green", "north_feed_grey", "north_feed_blue", "north_feed_yellow", "north_feed_purple", "north_feed_white",
		"south_feed_red", "south_feed_green", "south_feed_grey", "south_feed_blue", "south_feed_yellow", "south_feed_purple", "south_feed_white",
	}
	err = w.Write(record)
	if err != nil {
		return
	}

	for _, m := range matches {
		m.North, m.South = new(Team), new(Team)
		err = db.Where("is_north = ?", true).Where("match_id = ?", m.ID).First(m.North).Error
		if err != nil {
			return err
		}
		err = db.Where("is_north = ?", false).Where("match_id = ?", m.ID).First(m.South).Error
		if err != nil {
			return err
		}

		base := []int32{
			int32(m.ID), int32(m.Length), int32(m.Difficulty), int32(len(m.Players)), int32(m.Winner),
		}
		record = intSliceToStringSlice(append(append(base, []int32(m.North.TotalFeed)...), []int32(m.South.TotalFeed)...))
		err = w.Write(record)
		if err != nil {
			return
		}
	}
	w.Flush()
	return w.Error()
}

func updateTeamELO(t *Team) (err error) {
	for _, p := range t.Players {
		err = db.Model(&Player{}).Where("name = ?", p.Player.Name).Update("ELO", p.BeforeELO+p.GainELO).Error
		if err != nil {
			log.Println("failed to update ELO for", p.Player.Name, ":", err)
			return nil // Ignore error
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
		if p.Player.Name == name {
			return true
		}
	}
	return false
}

func QueryPlayerMatchesShort(name string) (gp []*GamePlayer, err error) {
	player := new(Player)
	err = db.Where("name = ?", name).First(player).Error
	if err != nil {
		return nil, err
	}
	err = db.Preload("Match").Omit("Player").Order("id desc").Where("player_id = ?", player.ID).Find(&gp).Error
	return
}

func FillPlayersWithELO(players []*GamePlayer) {
	for _, p := range players {
		if p.PlayerID != 0 { // Skip already associated GamePlayers.
			continue
		}
		if p.BeforeELO == 0 {
			err := db.Where("name", p.Player.Name).Attrs("elo", startELO).FirstOrInit(&p.Player).Error
			if err != nil {
				log.Println("failed to query/create player in db:", err)
				continue
			}
			p.BeforeELO = p.Player.ELO
			p.PlayerID = p.Player.ID
		}
	}
}
