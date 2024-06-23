package server

import "fmt"

type ErrInvalidOrderNumber struct {
	orderNumber int
}

func (e *ErrInvalidOrderNumber) Error() string {
	return fmt.Sprintf("incorrect format number for %v order", e.orderNumber)
}

