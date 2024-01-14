package mockdb

import (
	"errors"
	"github.com/shopspring/decimal"

	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

type MockDb struct{}

func (m *MockDb) GetBalance(id int) (*wallet.Wallet, error) {
	if id <= 0 {
		return nil, database.UserDoesNotExistErr
	}
	testBalance, _ := decimal.NewFromString("300")
	return &wallet.Wallet{ID: id, Balance: testBalance}, nil
}

func (m *MockDb) GetHistory(id int, orderBy database.OrderBy, order database.Order, limit int) (*wallet.Wallet, error) {
	var err = errors.New("test error")
	if id < 0 {
		return nil, err
	}
	return &wallet.Wallet{ID: id, History: make([]wallet.HistoryChange, id)}, nil
}

func (m *MockDb) CommitChanges(id int, balance decimal.Decimal, ch wallet.HistoryChange) error {
	return nil
}

func (m *MockDb) NewUser(id int) error {
	return nil
}

func (m *MockDb) Rollback() {}

func (m *MockDb) Commit() error {
	return nil
}
