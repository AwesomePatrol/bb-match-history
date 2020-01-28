package stats

import (
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Player struct {
	Name    string   `gorm:"PRIMARY_KEY"`
	History []*Match `gorm:"many2many:player_match;"`
}

type MVPplayer struct {
	Player *Player
	Stat   int
}

type Team struct {
	gorm.Model
	Players     []*Player `gorm:"many2many:player_team;"`
	MVPdefender MVPplayer
	MVPbuilder  MVPplayer
	MVPdeaths   MVPplayer
	IsNorth     bool
	MatchID     int64
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
	MatchID int64
}

type Difficulty int

const (
	Peaceful Difficulty = iota
	PieceOfCake
	Easy
	Normal
	Hard
	Nightmare
	Insane
)

func (p *Difficulty) Scan(value interface{}) error {
	*p = Difficulty(value.(int64))
	return nil
}

func (p Difficulty) Value() (string, error) {
	return string(p), nil
}

type Match struct {
	gorm.Model
	Players      []*Player `gorm:"many2many:player_match;"`
	South, North Team      `gorm:"foreignkey:MatchID"`
	Start, End   time.Time
	Length       time.Duration
	NorthWon     bool
	Difficulty   `sql:"type:difficulty"`
	Timeline     []*Event `gorm:"foreignkey:MatchID"`
}
