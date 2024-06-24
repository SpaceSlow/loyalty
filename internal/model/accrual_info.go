package model

import "time"

type AccrualInfo struct {
	OrderNumber string    `json:"order"`
	Status      string    `json:"status"`
	Sum         float64   `json:"accrual,omitempty"`
	CreatedAt   time.Time `json:"uploaded_at,omitempty"`
}
