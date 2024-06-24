package store

import "fmt"

type ErrNoUser struct {
	username string
}

func (e *ErrNoUser) Error() string {
	return fmt.Sprintf("%s is not existed", e.username)
}

type ErrOrderAlreadyExist struct {
	UserID int
}

func (e *ErrOrderAlreadyExist) Error() string {
	return fmt.Sprintf("order has been already added user (user_id = %d)", e.UserID)
}

type ErrWithdrawalAlreadyExist struct {
	Order string
}

func (e *ErrWithdrawalAlreadyExist) Error() string {
	return fmt.Sprintf("withdrawal with order (%s) has been already exist", e.Order)
}
