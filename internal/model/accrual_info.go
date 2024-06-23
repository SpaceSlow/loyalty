package model

import "time"

type AccrualInfo struct {
	OrderNumber string    `json:"order"`
	Status      string    `json:"status"`
	Sum         int       `json:"accrual,omitempty"`
	CreatedAt   time.Time `json:"uploaded_at,omitempty"`
}
