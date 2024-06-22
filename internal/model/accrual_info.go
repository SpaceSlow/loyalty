package model

type AccrualInfo struct {
	OrderNumber string `json:"order"`
	Status      string `json:"status"`
	Sum         int    `json:"accrual,omitempty"`
}
