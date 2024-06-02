package store

import "fmt"

type ErrNoUser struct {
	username string
}

func (e *ErrNoUser) Error() string {
	return fmt.Sprintf("%s is not existed", e.username)
}
