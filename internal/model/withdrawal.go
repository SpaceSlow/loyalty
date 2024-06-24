package model

import "time"

type Withdrawal struct {
	OrderNumber string    `json:"order"`
	Sum         float64   `json:"sum"`
	CreatedAt   time.Time `json:"processed_at,omitempty"`
}
