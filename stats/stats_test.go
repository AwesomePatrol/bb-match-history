package stats

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInsertMatch(t *testing.T) {
	testdata := new(Match)
	err := json.Unmarshal([]byte(exampleJSON1), testdata)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	testdata.North.IsNorth = true
	OpenDB(":memory:")
	//OpenDB("test.db")
	defer CloseDB()

	err = InsertMatch(testdata)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	returndata, err := QueryMatchShort(1)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	returnjson, _ := json.Marshal(returndata)
	t.Log(string(returnjson))

	// try again
	testdata.ID = 0
	err = InsertMatch(testdata)
	if err == nil {
		t.Log("should fail to insert same match twice")
		t.FailNow()
	}

	// try again, but keep it unique
	testdata.Start = time.Now()
	err = InsertMatch(testdata)
	if err != nil {
		t.Log("should allow same match but on different time")
		t.FailNow()
	}

	// check db contents
	m, err := QueryMatchAll()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if len(m) != 2 {
		t.Log(m)
		t.Log("expected 2 matches to be present in db")
		t.FailNow()
	}

	players := make([]Player, 0, 128)
	err = db.Find(&players).Error
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	if len(players) != 9 {
		t.Log(players)
		t.Log("expected 9 players to be present in db")
		t.FailNow()
	}
}
