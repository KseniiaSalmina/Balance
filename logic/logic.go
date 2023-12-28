package logic

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"

	"github.com/KseniiaSalmina/Balance/database"
	"github.com/KseniiaSalmina/Balance/wallet"
)

type Storage interface {
	GetBalance(id int) (*wallet.Wallet, error)
	GetHistory(id int, orderBy database.OrderBy, order database.Order, limit int) (*wallet.Wallet, error)
	CommitChanges(id int, balance string, ch wallet.Change) error
	NewUser(id int) error
}

func MoneyTransaction(s Storage, id int, opt wallet.Operation, amount, desc string) error {
	w, err := s.GetBalance(id)
	if err != nil && errors.Is(err, database.UserDoesNotExistErr) {
		switch opt {
		case wallet.Withdrawal:
			return fmt.Errorf("problem with getting balance: %w", err)
		case wallet.Replenishment:
			err = s.NewUser(id)
			w = &wallet.Wallet{ID: id, Balance: decimal.NewFromInt(0)}
			if err != nil {
				return fmt.Errorf("problem with creating a new user: %w", err)
			}
		}
	}

	if err = w.ChangeBalance(amount, opt); err != nil {
		return fmt.Errorf("money transaction problem: %w", err)
	}

	ch := wallet.NewChange(opt, amount, desc)

	if err = s.CommitChanges(id, w.StringBalance(), ch); err != nil {
		return fmt.Errorf("finishing money transaction problem: %w", err)
	}

	return nil
}

func Transfer(s Storage, from, to int, amount string) error {
	err := MoneyTransaction(s, from, wallet.Withdrawal, amount, fmt.Sprintf("transfer to user %v", to))
	if err != nil {
		return fmt.Errorf("transfer error: %w", err)
	}

	err = MoneyTransaction(s, to, wallet.Replenishment, amount, fmt.Sprintf("transfer from user %v", from))
	if err != nil {
		return fmt.Errorf("transfer error: %w", err)
	}

	return nil
}

func CheckBalance(s Storage, id int) (string, error) {
	w, err := s.GetBalance(id)
	if err != nil {
		return "", err
	}

	return w.StringBalance(), nil
}

func CheckHistory(s Storage, id int, orderBy database.OrderBy, order database.Order, limit int) ([]wallet.Change, error) {
	w, err := s.GetHistory(id, orderBy, order, limit)
	if err != nil {
		return nil, err
	}

	return w.History, nil
}
