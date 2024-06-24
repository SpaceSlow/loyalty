package server

import (
	"errors"
	"fmt"
)

var ErrGeneratePasswordHash = errors.New("error occurred when generating password")

type ErrInvalidOrderNumber struct {
	orderNumber int
}

func (e *ErrInvalidOrderNumber) Error() string {
	return fmt.Sprintf("incorrect format number for %v order", e.orderNumber)
}

type ErrNotEnoughLoyaltyPoints struct {
	current float64
}

func (e *ErrNotEnoughLoyaltyPoints) Error() string {
	return fmt.Sprintf("not enough loyalty points. current balance: %0.1f", e.current)
}

type ErrIncorrectWithdrawalSum struct {
	sum float64
}

func (e *ErrIncorrectWithdrawalSum) Error() string {
	return fmt.Sprintf("withdrawal must be positive float number. Got: %0.1f", e.sum)
}
