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

func QueryMatchShort(id int) (*Match, error) {
	match := new(Match)
	return match, db.First(match, id).Error
}

func QueryMatchLong(id int) (match *Match, err error) {
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

	err = db.Where("match_id = ?", id).Find(&match.Timeline).Error
	return
}

func QueryMatchAll() (matches []Match, err error) {
	// TODO pagination?
	matches = make([]Match, 0, 128)
	err = db.Limit(128).Order("id desc").Find(&matches).Error
	return
}
