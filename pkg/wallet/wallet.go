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
	History []Change
}

type Change struct {
	Date int64
	Operation
	Amount      string
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

func (w *Wallet) ChangeBalance(amount string, opt Operation) error {
	a, err := decimal.NewFromString(amount)
	if err != nil {
		return err
	}

	switch opt {
	case Replenishment:
		w.Balance = w.Balance.Add(a)
		return nil
	case Withdrawal:
		test := w.Balance.Sub(a)
		if test.GreaterThanOrEqual(decimal.Zero) {
			w.Balance = test
			return nil
		}
		return InsufficientFundsErr
	}

	return errors.New("invalid operation")
}

func NewChange(opt Operation, amount string, descr string) Change {
	return Change{Date: time.Now().Unix(), Operation: opt, Amount: amount, Description: descr}
}
