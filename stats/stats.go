package stats

import (
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

type Model struct {
	ID        uint       `gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

type EmptyModel struct {
	ID        uint       `gorm:"primary_key" json:"-"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"-"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}

type Player struct {
	Name    string  `gorm:"PRIMARY_KEY"`
	History []Match `gorm:"many2many:player_match;" json:"-"`
}

type MVPplayer struct {
	EmptyModel
	Name  string
	Title string
	Stat  int
}

type Team struct {
	EmptyModel
	Players []*Player    `gorm:"many2many:player_team;"`
	MVPs    []*MVPplayer `gorm:"many2many:mvp_team;" json:",omitempty"`
	IsNorth bool         `json:"-"`
	MatchID int64        `json:"-"`
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
	EmptyModel
	Timestamp time.Time
	EventType
	Payload string
	MatchID int64 `json:"-"`
}

type Difficulty int64

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
	Model
	Players      []*Player `gorm:"many2many:player_match;" json:",omitempty"`
	South, North *Team     `gorm:"foreignkey:MatchID" json:",omitempty"`
	Start, End   time.Time `json:",omitempty"`
	Length       time.Duration
	NorthWon     bool
	Difficulty   `sql:"type:difficulty"`
	Timeline     []*Event `gorm:"foreignkey:MatchID" json:",omitempty"`
}
