package parser

import (
	"strings"
	"testing"
)

func TestParseSingleMatch(t *testing.T) {
	match, err := ParseSingleMatch(strings.NewReader(example1))
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log("North")
	for _, player := range match.North.Players {
		t.Log(player)
	}
	t.Log("South")
	for _, player := range match.South.Players {
		t.Log(player)
	}
	t.Log("Timeline")
	for _, event := range match.Timeline {
		t.Log(event)
	}
	t.Log("winner:", match.NorthWon)
	t.Log(match.Length)
	t.Log(match.Difficulty)
}
