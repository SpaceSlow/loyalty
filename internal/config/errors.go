package config

import (
	"errors"
	"fmt"
)

type ErrIncorrectPort struct {
	Port string
}

func (e *ErrIncorrectPort) Error() string {
	return fmt.Sprintf("error occurred when parsing an incorrect port. The port requires a decimal number in the range 0-65535. Got: %s", e.Port)
}

var (
	ErrIncorrectNetAddress = errors.New("need address in a form host:port")
	ErrEmptyDSN            = errors.New("flag error: needed DSN. check specification")
	ErrEmptyAccrualAddress = errors.New("flag error: needed accrual system address. check specification")
)
