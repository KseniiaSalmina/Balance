package logic

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
)

type Storage interface {
	GetBalance(id int) (*Wallet, error)
	GetHistory(id int, orderBy OrderBy, order Order) (*Wallet, error)
	CommitChanges(id int, balance string, ch Change) error
	NewUser(id int) error
}

func MoneyTransaction(s Storage, id int, opt Operation, amount, desc string) error {
	wallet, err := s.GetBalance(id)
	if err != nil && errors.Is(err, UserDoesNotExistErr) {
		switch opt {
		case Withdrawal:
			return fmt.Errorf("problem with getting balance: %w", err)
		case Replenishment:
			err = s.NewUser(id)
			wallet = &Wallet{ID: id, Balance: decimal.NewFromInt(0)}
			if err != nil {
				return fmt.Errorf("problem with creating a new user: %w", err)
			}
		}
	}

	if err = wallet.ChangeBalance(amount, opt); err != nil {
		return fmt.Errorf("money transaction problem: %w", err)
	}

	ch := NewChange(opt, amount, desc)

	if err = s.CommitChanges(id, wallet.StringBalance(), ch); err != nil {
		return fmt.Errorf("finishing money transaction problem: %w", err)
	}

	return nil
}

func Transfer(s Storage, from, to int, amount string) error {
	err := MoneyTransaction(s, from, Withdrawal, amount, fmt.Sprintf("transfer to user %v", to))
	if err != nil {
		return fmt.Errorf("transfer error: %w", err)
	}

	err = MoneyTransaction(s, to, Replenishment, amount, fmt.Sprintf("transfer from user %v", from))
	if err != nil {
		return fmt.Errorf("transfer error: %w", err)
	}

	return nil
}

func CheckBalance(s Storage, id int) (string, error) {
	wallet, err := s.GetBalance(id)
	if err != nil {
		return "", err
	}

	return wallet.StringBalance(), nil
}

type Page []Change

func CheckHistory(s Storage, id int, orderBy OrderBy, order Order) ([]Page, error) {
	wallet, err := s.GetHistory(id, orderBy, order)
	if err != nil {
		return nil, err
	}

	pages := make([]Page, 0, 10)
	for i := 0; i < len(wallet.History) && i < 100; i += 10 {
		page := make(Page, 0, 10)
		for j := 0; i+j < len(wallet.History) && j < 10; j++ {
			page = append(page, wallet.History[i+j])
		}
		pages = append(pages, page)
	}

	return pages, nil
}
