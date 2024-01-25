package api

import "github.com/shopspring/decimal"

type changingBalanceRequest struct {
	IsTransfer  bool            //reports whether transaction is a transfer or not, default false
	To          int             //required for a transfer
	Amount      decimal.Decimal //for a transfer must be a positive number, for a not transfer transaction reports whether the operation is a replenishment (positive amount) or withdrawal (negative)
	Description string          //required for a not transfer transactions
}
