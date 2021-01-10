package science

import "fmt"

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
