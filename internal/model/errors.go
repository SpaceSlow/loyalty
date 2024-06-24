package model

import (
	"errors"
	"fmt"
)

var ErrInvalidPasswordHash = errors.New("invalid password hash layout")

type ErrUnknownHashAlg struct {
	Alg string
}

func (e *ErrUnknownHashAlg) Error() string {
	return fmt.Sprintf("unknown password hash algorithm: %s", e.Alg)
}
