package database

import (
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/shopspring/decimal"
	"strconv"

	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

type Transaction struct {
	tx *pgx.Tx
}

func NewTransaction(db *pgx.Conn) (*Transaction, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, fmt.Errorf("NewTransaction -> %w", err)
	}
	return &Transaction{tx: tx}, nil
}

func (t *Transaction) Rollback() {
	t.tx.Rollback()
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) GetBalance(id int) (*wallet.Wallet, error) {
	var balance decimal.Decimal
	if err := t.tx.QueryRow(`SELECT balance FROM balances WHERE id = $1`, id).Scan(&balance); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, UserDoesNotExistErr
		}
		return nil, fmt.Errorf("GetBalance -> %w", err)
	}
	return &wallet.Wallet{ID: id, Balance: balance}, nil
}

func (t *Transaction) GetHistory(walletID int, orderBy OrderBy, order Order, limit int) (*wallet.Wallet, error) {
	query := `SELECT date, option, amount, description FROM history WHERE wallet_id = $1` + ` ORDER BY ` + string(orderBy) + ` ` + string(order) + ` LIMIT ` + strconv.Itoa(limit)
	rows, err := t.tx.Query(query, walletID)
	if err != nil {
		return nil, fmt.Errorf("getHistory -> %w", err)
	}

	w := &wallet.Wallet{ID: walletID, History: make([]wallet.HistoryChange, 0, limit+1)}
	for rows.Next() {
		var c wallet.HistoryChange
		var date pgtype.Int8
		var amount decimal.Decimal
		var operation, description pgtype.Text
		if err = rows.Scan(&date, &operation, &amount, &description); err != nil {
			return nil, fmt.Errorf("GetHistory -> %w", err)
		}

		c.Date, c.Operation, c.Amount, c.Description = date.Int, wallet.Operation(operation.String), amount, description.String
		w.History = append(w.History, c)
	}

	if len(w.History) == 0 {
		return nil, UserDoesNotExistErr
	}

	return w, nil
}

func (t *Transaction) CommitChanges(id int, balance decimal.Decimal, ch wallet.HistoryChange) error {
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
