package main

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/shopspring/decimal"
)

type Database struct {
	db *pgx.Conn
}

type Transaction struct {
	tx *pgx.Tx
}

//OrderBy can be date or amount
type OrderBy string

const (
	OrderByDate   OrderBy = "date"
	OrderByAmount OrderBy = "amount"
)

//Order can be DESC or ASC
type Order string

const (
	Desc Order = "DESC"
	Asc  Order = "ASC"
)

var UserDoesNotExistErr error = errors.New("user does not exist")

func (t *Transaction) GetBalance(id int) (*Wallet, error) {
	row := t.tx.QueryRow(`SELECT balance FROM balances WHERE id = $1`, id)

	var balance pgtype.Text
	if err := row.Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, UserDoesNotExistErr
		}
		return nil, fmt.Errorf("GetBalance -> %w", err)
	}

	formatBalance, err := decimal.NewFromString(balance.String)
	if err != nil {
		return nil, fmt.Errorf("GetBalance -> %w", err)
	}

	return &Wallet{ID: id, Balance: formatBalance}, nil
}

func (t *Transaction) GetHistory(wallet_id int, orderBy OrderBy, order Order) (*Wallet, error) {
	query := `SELECT date, option, amount, description FROM history WHERE wallet_id = $1` + ` ORDER BY ` + string(orderBy) + ` ` + string(order) + ` LIMIT 100`
	rows, err := t.tx.Query(query, wallet_id)
	if err != nil {
		return nil, fmt.Errorf("getHistory -> %w", err)
	}

	w := &Wallet{ID: wallet_id, History: make([]Change, 0, 101)}
	for rows.Next() {
		var c Change
		var date pgtype.Int8
		var operation, amount, description pgtype.Text
		if err = rows.Scan(&date, &operation, &amount, &description); err != nil {
			return nil, fmt.Errorf("GetHistory -> %w", err)
		}

		c.Date, c.Operation, c.Amount, c.Description = date.Int, Operation(operation.String), amount.String, description.String
		w.History = append(w.History, c)
	}

	if len(w.History) == 0 {
		return nil, UserDoesNotExistErr
	}

	return w, nil
}

func (t *Transaction) CommitChanges(id int, balance string, ch Change) error {
	_, err := t.tx.Exec(`UPDATE balances SET balance = $1 WHERE id = $2`, balance, id)
	if err != nil {
		return fmt.Errorf("ChangeBalance -> %w", err)
	}

	_, err = t.tx.Exec(`INSERT INTO history (wallet_id, date, option, amount, description) VALUES ($1, $2, $3, $4, $5)`, id, ch.Date, ch.Operation, ch.Amount, ch.Description)
	if err != nil {
		return fmt.Errorf("ChangeBalance -> %w", err)
	}

	return nil
}

func (t *Transaction) NewUser(id int) error {
	_, err := t.tx.Exec(`INSERT INTO balances (id) VALUES ($1)`, id)
	if err != nil {
		return fmt.Errorf("NewUser -> %w", err)
	}
	return nil
}