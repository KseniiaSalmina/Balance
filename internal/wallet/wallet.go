package wallet

import (
	"errors"
	"github.com/shopspring/decimal"
	"time"
)

var InsufficientFundsErr = errors.New("insufficient funds")

type Wallet struct {
	ID      int
	Balance decimal.Decimal
	History []HistoryChange
}

type HistoryChange struct {
	Date int64
	Operation
	Amount      decimal.Decimal
	Description string
}

//Operation can be replenishment or withdrawal
type Operation string

const (
	Replenishment Operation = "replenishment"
	Withdrawal    Operation = "withdrawal"
)

func (w *Wallet) StringBalance() string {
	return w.Balance.String()
}

func (w *Wallet) ChangeBalance(amount decimal.Decimal, opt Operation) error {
	switch opt {
	case Replenishment:
		w.Balance = w.Balance.Add(amount)
		return nil
	case Withdrawal:
		test := w.Balance.Sub(amount)
		if test.GreaterThanOrEqual(decimal.Zero) {
			w.Balance = test
			return nil
		}
		return InsufficientFundsErr
	}

	return errors.New("invalid operation")
}

func NewChange(opt Operation, amount decimal.Decimal, descr string) HistoryChange {
	return HistoryChange{Date: time.Now().Unix(), Operation: opt, Amount: amount, Description: descr}
}
