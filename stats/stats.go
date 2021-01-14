package stats

import (
	"fmt"
	"time"

	"github.com/awesomepatrol/bb-match-history/stats/const/difficulty"
	"github.com/awesomepatrol/bb-match-history/stats/const/science"
)

type Model struct {
	ID uint `gorm:"primaryKey"`
}

type EmptyModel struct {
	ID uint `gorm:"primaryKey" json:"-"`
}

type Player struct {
	Name    string `gorm:"PRIMARY_KEY"`
	ELO     int
	History []Match `gorm:"many2many:player_match;" json:"-"`
}

type GamePlayer struct {
	Player
	Force Force
}

type Channel struct {
	ID      string  `gorm:"PRIMARY_KEY"`
	History []Match `gorm:"foreignkey:ChannelID"`
}

type MVPquery struct {
	Name  string
	Stat  int
	Total int
}

type MVPplayer struct {
	EmptyModel
	Name  string
	Title string
	Stat  int
}

type Team struct {
	EmptyModel
	Players   []*Player    `gorm:"many2many:player_team;"`
	MVPs      []*MVPplayer `gorm:"many2many:mvp_team;" json:",omitempty"`
	AvgELO    int
	TotalFeed science.Feed `gorm:"type:integer[]"`
	IsNorth   bool         `gorm:"type:bool" json:"-"`
	MatchID   int64        `json:"-"`
}

type EventType int64

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
	Feed
)

func (p *EventType) Scan(value interface{}) error {
	*p = EventType(value.(int64))
	return nil
}

func (p EventType) Value() (string, error) {
	return fmt.Sprint(p), nil
}

type Force int64

const (
	Spectator Force = iota
	North
	South
)

func (p *Force) Scan(value interface{}) error {
	*p = Force(value.(int64))
	return nil
}

func (p Force) Value() (string, error) {
	return fmt.Sprint(p), nil
}

type Event struct {
	EmptyModel
	Timestamp time.Time
	EventType EventType
	Payload   string
	MatchID   int64 `json:"-"`
}

type Match struct {
	Model
	Players      []*Player `gorm:"many2many:player_match;" json:",omitempty"`
	South, North *Team     `gorm:"foreignkey:MatchID" json:",omitempty"`
	Start        time.Time `gorm:"UNIQUE" json:",omitempty"`
	End          time.Time `json:",omitempty"`
	Length       time.Duration
	Winner       Force
	Difficulty   difficulty.Difficulty `sql:"type:difficulty"`
	Timeline     []*Event              `gorm:"foreignkey:MatchID" json:",omitempty"`
	ChannelID    string                `json:"-"`
}

func (m *Match) String() string {
	return fmt.Sprintf("start: %v end: %v difficulty: %d players: %d", m.Start, m.End, m.Difficulty, len(m.Players))
}

type PlayerMatch struct {
	EmptyModel
	Match     *Match `gorm:"-"`
	IsWinner  *bool  // IsWinner is a pointer to indicate situtation when player is just a spectator.
	FlasksFed science.Feed
	BeforeELO int
	GainELO   int
}
