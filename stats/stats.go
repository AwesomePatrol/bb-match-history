package stats

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Player struct {
	gorm.Model
	Name    string
	History []*Match `gorm:"many2many:player_match;"`
}

type MVPplayer struct {
	Player *Player
	Stat   int
}

type Team struct {
	Players     []*Player `gorm:"many2many:player_match;"`
	MVPdefender MVPplayer
	MVPbuilder  MVPplayer
	MVPdeaths   MVPplayer
}

type EventType int

const (
	GameStart EventType = iota
	GameEnd
	JoinGame
	LeaveGame
	JoinTeam
	LeaveTeam
	PlayerDeath
	DifficultyChange
	WinnerAnnounce
	GameTimeAnnounce
)

type Event struct {
	gorm.Model
	Timestamp time.Time
	EventType
	Payload string
}

type Match struct {
	gorm.Model
	South, North Team
	Start, End   time.Time
	Length       time.Duration
	NorthWon     bool
	Timeline     []*Event
}
