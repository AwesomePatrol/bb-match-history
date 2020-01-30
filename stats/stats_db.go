package stats

import (
	"fmt"

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
	return db.Create(match).Error
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

func QueryMatchShort(id int) (*Match, error) {
	return queryMatchShort(id)
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
	err = db.Limit(128).Order("id desc").Find(&matches).Error
	return
}

func isInWinningTeam(name string, team []*Player) bool {
	for _, p := range team {
		if p.Name == name {
			return true
		}
	}
	return false
}

func QueryPlayerMatches(name string) ([]Match, error) {
	player := Player{Name: name}
	err := db.Preload("History").First(&player).Error
	if err != nil {
		return nil, err
	}

	var falseV = false
	var trueV = true
	for i, match := range player.History {
		var team Team
		err = db.Preload("Players").Where("is_north = ?", match.NorthWon).Where("match_id = ?", match.ID).First(&team).Error
		if err != nil {
			return nil, err
		}
		if isInWinningTeam(name, team.Players) {
			player.History[i].IsWinner = &trueV
		} else {
			player.History[i].IsWinner = &falseV
		}
	}
	return player.History, nil
}
