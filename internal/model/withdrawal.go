package model

import "time"

type WithdrawalInfo struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	CreatedAt   time.Time `json:"processed_at,omitempty"`
}
