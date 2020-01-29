package stats

import (
	"encoding/json"
	"testing"
)

func TestInsertMatch(t *testing.T) {
	testdata := new(Match)
	err := json.Unmarshal([]byte(exampleJSON1), testdata)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	OpenDB(":memory:")
	//OpenDB("test.db")
	defer CloseDB()

	err = InsertMatch(testdata)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	returndata, err := QueryMatch(nil)
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	returnjson, _ := json.Marshal(returndata)
	t.Log(string(returnjson))
}
