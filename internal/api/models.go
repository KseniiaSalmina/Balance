package api

import "github.com/shopspring/decimal"

type changingBalanceRequest struct {
	isTransfer  bool
	to          int
	amount      decimal.Decimal
	description string
}
