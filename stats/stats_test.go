package stats

import (
	"encoding/json"
	"testing"
)

func TestInsertMatch(t *testing.T) {
	t.Skip()
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
}
