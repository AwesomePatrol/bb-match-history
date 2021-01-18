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
	EmptyModel
	Name    string  `gorm:"unique"`
	ELO     int     `gorm:"-"`
	History []Match `gorm:"many2many:player_match;" json:"-"`
}

type GamePlayer struct {
	EmptyModel
	PlayerID  uint `json:"-"`
	Player    Player
	Force     Force
	MatchID   int64 `json:"-"`
	BeforeELO int
	GainELO   int
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
	Players     []*GamePlayer `gorm:"many2many:player_team;"`
	MVPs        []*MVPplayer  `gorm:"many2many:mvp_team;" json:",omitempty"`
	AvgELO      float64
	TotalFeed   science.Feed `gorm:"type:integer[]"`
	FinalEVO    float32
	FinalThreat int
	IsNorth     bool  `gorm:"type:bool" json:"-"`
	MatchID     int64 `json:"-"`
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
	Unknown Force = iota
	Spectator
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

func (p Force) Opposite() Force {
	return North + (South - p)
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
	Players      []*GamePlayer `gorm:"many2many:player_match;" json:",omitempty"`
	South, North *Team         `gorm:"foreignkey:MatchID" json:",omitempty"`
	Start        time.Time     `gorm:"UNIQUE" json:",omitempty"`
	End          time.Time     `json:",omitempty"`
	Length       time.Duration
	Winner       Force
	Difficulty   difficulty.Difficulty `sql:"type:difficulty"`
	Timeline     []*Event              `gorm:"foreignkey:MatchID" json:",omitempty"`
	ChannelID    string                `json:"-"`
}

func (m *Match) String() string {
	return fmt.Sprintf("start: %v end: %v difficulty: %d players: %d", m.Start, m.End, m.Difficulty, len(m.Players))
}
