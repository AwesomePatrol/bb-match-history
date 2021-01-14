package science

import (
	"database/sql/driver"
	"fmt"

	"github.com/lib/pq"
)

type Science int64

const (
	Automation Science = iota
	Logistic
	Military
	Chemical
	Production
	Utility
	Space
)

func (p *Science) Scan(value interface{}) error {
	*p = Science(value.(int64))
	return nil
}

func (p Science) Value() (string, error) {
	return fmt.Sprint(p), nil
}

type Feed []int32

func (p *Feed) Scan(value interface{}) error {
	return pq.Array(p).Scan(value)
}

func (p Feed) Value() (driver.Value, error) {
	return pq.Array(p).Value()
}
