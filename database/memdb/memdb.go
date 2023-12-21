package memdb

import (
	"errors"
	"github.com/shopspring/decimal"
)

type MockDb struct{}

func (m *MockDb) GetBalance(id int) (*Wallet, error) {
	if id <= 0 {
		return nil, UserDoesNotExistErr
	}
	testBalance, _ := decimal.NewFromString("300")
	return &Wallet{ID: id, Balance: testBalance}, nil
}

func (m *MockDb) GetHistory(id int, orderBy OrderBy, order Order) (*Wallet, error) {
	var err = errors.New("test error")
	if id < 0 {
		return nil, err
	}
	return &Wallet{ID: id, History: make([]Change, id)}, nil
}

func (m *MockDb) CommitChanges(id int, balance string, ch Change) error {
	return nil
}

func (m *MockDb) NewUser(id int) error {
	return nil
}
