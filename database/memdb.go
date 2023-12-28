package database

import (
	"errors"
	"github.com/shopspring/decimal"

	"github.com/KseniiaSalmina/Balance/wallet"
)

type MockDb struct{}

func (m *MockDb) GetBalance(id int) (*wallet.Wallet, error) {
	if id <= 0 {
		return nil, UserDoesNotExistErr
	}
	testBalance, _ := decimal.NewFromString("300")
	return &wallet.Wallet{ID: id, Balance: testBalance}, nil
}

func (m *MockDb) GetHistory(id int, orderBy OrderBy, order Order, limit int) (*wallet.Wallet, error) {
	var err = errors.New("test error")
	if id < 0 {
		return nil, err
	}
	return &wallet.Wallet{ID: id, History: make([]wallet.Change, id)}, nil
}

func (m *MockDb) CommitChanges(id int, balance string, ch wallet.Change) error {
	return nil
}

func (m *MockDb) NewUser(id int) error {
	return nil
}
