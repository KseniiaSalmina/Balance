package billing

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"

	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/database/mockdb"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

type Storage interface {
	GetBalance(id int) (*wallet.Wallet, error)
	GetHistory(id int, orderBy database.OrderBy, order database.Order, limit int) (*wallet.Wallet, error)
	CommitChanges(id int, balance decimal.Decimal, ch wallet.HistoryChange) error
	NewUser(id int) error
	Rollback()
	Commit() error
}

type Billing struct {
	db *database.DB
}

func NewBilling(db *database.DB) *Billing {
	return &Billing{
		db: db,
	}
}

func (b *Billing) MoneyTransaction(id int, opt wallet.Operation, amount decimal.Decimal, desc string) error {
	tx, err := b.beginTx()
	if err != nil {
		return fmt.Errorf("MoneyTransaction -> %w", err)
	}
	defer tx.Rollback()

	w, err := tx.GetBalance(id)
	if err != nil {
		if errors.Is(err, database.UserDoesNotExistErr) {
			switch opt {
			case wallet.Withdrawal:
				return fmt.Errorf("problem with getting balance: %w", err)
			case wallet.Replenishment:
				err = tx.NewUser(id)
				fmt.Println(err)
				if err != nil {
					return fmt.Errorf("problem with creating a new user: %w", err)
				}
				w = &wallet.Wallet{ID: id, Balance: decimal.NewFromInt(0)}
			}
		} else {
			return fmt.Errorf("problem with getting balance: %w", err)
		}
	}

	if err = w.ChangeBalance(amount, opt); err != nil {
		return fmt.Errorf("money transaction problem: %w", err)
	}

	ch := wallet.NewChange(opt, amount, desc)

	if err = tx.CommitChanges(id, w.Balance, ch); err != nil {
		return fmt.Errorf("finishing money transaction problem: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("MoneyTransaction -> %w", err)
	}

	return nil
}

func (b *Billing) beginTx() (Storage, error) {
	if b.db == nil {
		return &mockdb.MockDb{}, nil
	}

	tx, err := b.db.NewTransaction()
	if err != nil {
		return nil, fmt.Errorf("beginTx -> %w", err)
	}
	return tx, nil
}

func (b *Billing) moneyTransaction(s Storage, id int, opt wallet.Operation, amount decimal.Decimal, desc string) error {
	w, err := s.GetBalance(id)
	if err != nil {
		if errors.Is(err, database.UserDoesNotExistErr) {
			switch opt {
			case wallet.Withdrawal:
				return fmt.Errorf("problem with getting balance: %w", err)
			case wallet.Replenishment:
				err = s.NewUser(id)
				fmt.Println(err)
				if err != nil {
					return fmt.Errorf("problem with creating a new user: %w", err)
				}
				w = &wallet.Wallet{ID: id, Balance: decimal.NewFromInt(0)}
			}
		} else {
			return fmt.Errorf("problem with getting balance: %w", err)
		}
	}

	if err = w.ChangeBalance(amount, opt); err != nil {
		return fmt.Errorf("money transaction problem: %w", err)
	}

	ch := wallet.NewChange(opt, amount, desc)

	if err = s.CommitChanges(id, w.Balance, ch); err != nil {
		return fmt.Errorf("finishing money transaction problem: %w", err)
	}

	return nil
}

func (b *Billing) Transfer(from, to int, amount decimal.Decimal) error {
	tx, err := b.beginTx()
	if err != nil {
		return fmt.Errorf("Transfer -> %w", err)
	}

	err = b.moneyTransaction(tx, from, wallet.Withdrawal, amount, fmt.Sprintf("transfer to user %v", to))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("transfer error: %w", err)
	}

	err = b.moneyTransaction(tx, to, wallet.Replenishment, amount, fmt.Sprintf("transfer from user %v", from))
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("transfer error: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("Transfer -> %w", err)
	}
	return nil
}

func (b *Billing) CheckBalance(id int) (string, error) {
	tx, err := b.beginTx()
	if err != nil {
		return "", fmt.Errorf("billing.CheckBalance -> %w", err)
	}

	w, err := tx.GetBalance(id)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	tx.Commit()
	return w.StringBalance(), nil
}

func (b *Billing) CheckHistory(id int, orderBy database.OrderBy, order database.Order, limit int) ([]wallet.HistoryChange, error) {
	tx, err := b.beginTx()
	if err != nil {
		return nil, fmt.Errorf("billing.CheckHistory -> %w", err)
	}

	w, err := tx.GetHistory(id, orderBy, order, limit)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return w.History, nil
}
