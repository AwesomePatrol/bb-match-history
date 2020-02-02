package stats

import (
	"fmt"
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

var db *gorm.DB

func OpenDB(path string) {
	var err error
	db, err = gorm.Open("sqlite3", path)
	if err != nil {
		panic(fmt.Sprint("failed to connect database:", err))
	}
	db.AutoMigrate(&Player{})
	db.AutoMigrate(&Event{})
	db.AutoMigrate(&MVPplayer{})
	db.AutoMigrate(&Team{})
	db.AutoMigrate(&Match{})
}

func CloseDB() {
	db.Close()
}

func InsertMatch(match *Match) error {
	if len(match.Players) == 0 {
		log.Println("skipping empty match:", match)
		return nil
	}
	var n int
	db.Where("start = ?", match.Start).Find(new(Match)).Count(&n)
	if n > 0 {
		return fmt.Errorf("already in db")
	}
	return db.Create(match).Error
}

func QueryGlobalMVP(title string) (mvp []MVPquery, err error) {
	err = db.Table("mv_pplayers").Where("title = ?", title).Select("name, count(name) as stat").Group("name").Order("stat desc").Limit(10).Scan(&mvp).Error
	return
}

func queryMatchShort(id int) (match *Match, err error) {
	match = new(Match)
	matchDB := db.Preload("Players").First(match, id)
	if matchDB.Error != nil {
		return nil, matchDB.Error
	}

	match.North = new(Team)
	err = matchDB.Where("is_north = ?", true).Related(&match.North).Error
	if err != nil {
		return nil, err
	}

	match.South = new(Team)
	err = matchDB.Where("is_north = ?", false).Related(&match.South).Error
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

func isInTeam(name string, team []*Player) bool {
	for _, p := range team {
		if p.Name == name {
			return true
		}
	}
	return false
}

func QueryPlayerMatches(name string) ([]Match, error) {
	player := Player{Name: name}
	err := db.Preload("History", func(db *gorm.DB) *gorm.DB {
		return db.Order("matches.start DESC")
	}).First(&player).Error
	if err != nil {
		return nil, err
	}

	var falseV = false
	var trueV = true
	for i, match := range player.History {
		var teamW Team
		err = db.Preload("Players").Where("is_north = ?", match.NorthWon).Where("match_id = ?", match.ID).First(&teamW).Error
		if err != nil {
			return nil, err
		}
		var teamL Team
		err = db.Preload("Players").Where("is_north = ?", !match.NorthWon).Where("match_id = ?", match.ID).First(&teamL).Error
		if err != nil {
			return nil, err
		}
		switch {
		case isInTeam(name, teamW.Players):
			player.History[i].IsWinner = &trueV
		case isInTeam(name, teamL.Players):
			player.History[i].IsWinner = &falseV
		default:
			player.History[i].IsWinner = nil
		}
	}
	return player.History, nil
}
