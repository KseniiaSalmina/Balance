package api

import "github.com/shopspring/decimal"

type ChangingBalanceRequest struct {
	IsTransfer  bool            `json:"is_transfer"` //reports whether transaction is a transfer or not, default false
	To          int             `json:"to"`          //required for a transfer
	Amount      decimal.Decimal `json:"amount"`      //for a transfer must be a positive number, for a not transfer transaction reports whether the operation is a replenishment (positive amount) or withdrawal (negative)
	Description string          `json:"description"` //required for a not transfer transactions
}
