package difficulty

import (
	"errors"
	"fmt"
)

type Difficulty int64

const (
	Peaceful Difficulty = iota
	PieceOfCake
	Easy
	Normal
	Hard
	Nightmare
	UltraViolence
	FunAndFast
)

func (p *Difficulty) Scan(value interface{}) error {
	*p = Difficulty(value.(int64))
	return nil
}

func (p Difficulty) Value() (string, error) {
	return fmt.Sprint(p), nil
}

var str2diff = map[string]Difficulty{
	"Peaceful":             Peaceful,
	"I'm Too Young to Die": Peaceful,
	"Piece of Cake":        PieceOfCake,
	"Easy":                 Easy,
	"Normal":               Normal,
	"Hard":                 Hard,
	"Nightmare":            Nightmare,
	"Ultra-Violence":       UltraViolence,
	"Insane":               FunAndFast,
	"Fun and Fast":         FunAndFast,
}

var ErrNotFound = errors.New("Unknown difficulty")

func StringToDifficulty(s string) (d Difficulty, err error) {
	d, ok := str2diff[s]
	if !ok {
		return -1, ErrNotFound
	}
	return d, nil
}
