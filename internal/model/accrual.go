package model

import "time"

type Accrual struct {
	OrderNumber string    `json:"number"`
	Status      string    `json:"status"`
	Sum         float64   `json:"accrual,omitempty"`
	CreatedAt   time.Time `json:"uploaded_at,omitempty"`
}

type ExternalAccrual struct {
	Accrual
	OrderNumber string `json:"order"`
}
