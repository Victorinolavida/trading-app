package dtos

import "time"

type PriceRequest struct {
	SecurityID   string    `json:"security_id" validate:"required"`
	SettlementDate time.Time `json:"settlement_date" validate:"required"`
	Yield          float64   `json:"yield" validate:"required"`
}

type PriceResult struct {
	CleanPrice      string `json:"clean_price"`
	AccruedInterest string `json:"accrued_interest"`
	DirtyPrice      string `json:"dirty_price"`
	MXNValue        string `json:"mxn_value"`
}

type YieldRequest struct {
	SecurityID   string    `json:"security_id" validate:"required"`
	SettlementDate time.Time `json:"settlement_date" validate:"required"`
	CleanPrice   string    `json:"clean_price" validate:"required"`
}

type CashflowRequest struct {
	SecurityID string `json:"security_id" validate:"required"`
}

type CashFlow struct {
	Date  string `json:"date"`
	Type  string `json:"type"`
	Rate  string `json:"rate"`
}