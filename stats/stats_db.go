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

func QueryMatch(query *Match) (*Match, error) {
	match := new(Match)
	matchDB := db.Preload("Players").First(&match)
	matchDB.Where("is_north = ?", true).Related(&match.North)
	matchDB.Where("is_north = ?", false).Related(&match.South)
	return match, matchDB.Error
}
