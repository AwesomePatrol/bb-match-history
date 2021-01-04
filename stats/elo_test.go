package stats

import (
	"log"
	"math"
	"testing"
)

func testExpectedScore(t *testing.T, a, b, expected float64) {
	res := expectedScore(a, b)
	if math.Abs(res-expected) > 0.01 {
		log.Println(a, "vs", b, ":", res, "expected:", expected)
		t.Fail()
	}
}

func TestELO_Update(t *testing.T) {
	testExpectedScore(t, 800, 800, 0.5)
	testExpectedScore(t, 600, 800, 0.24)
	testExpectedScore(t, 800, 600, 1-0.24)
}
