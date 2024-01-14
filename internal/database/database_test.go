package database

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"

	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

var config = pgx.ConnConfig{User: "kseniia", Password: "Efbcnwww1", Database: "testdb"} //TODO
var testTime = time.Now().Unix()
var testTime2 = testTime + 1

func prepareDB() *pgx.Conn {
	db, err := pgx.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	balance1, balance2, balance3 := decimal.NewFromInt(1), decimal.NewFromInt(0), decimal.NewFromInt(44000)
	balanceToHistory1, balanceToHistory2 := decimal.NewFromInt(1001), decimal.NewFromInt(1000)

	db.Exec(`INSERT INTO balances VALUES (4, $1), (5, $2), (6, $3);`, balance1, balance2, balance3)
	db.Exec(`INSERT INTO history(wallet_id, date, option, amount, description) VALUES(4, $1, $2, $3, $4), (4, $5, $6, $7, $8)`, testTime, wallet.Replenishment, balanceToHistory1, "деньги за продажу почки", testTime2, wallet.Withdrawal, balanceToHistory2, "почка не подошла")

	return db
}

func cleanup(db *pgx.Conn) {
	db.Exec(`TRUNCATE TABLE history;`)
	db.Exec(`DELETE FROM balances WHERE id = 1;`)
	db.Exec(`DELETE FROM balances WHERE id = 3;`)
	db.Exec(`DELETE FROM balances WHERE id = 123;`)
	db.Exec(`DELETE FROM balances WHERE id = 4;`)
	db.Exec(`DELETE FROM balances WHERE id = 5;`)
	db.Exec(`DELETE FROM balances WHERE id = 6;`)
	db.Exec(`ALTER SEQUENCE history_id_seq RESTART WITH 1;`)
	db.Close()
}

func TestTransaction_NewUser(t1 *testing.T) {
	db := prepareDB()
	defer cleanup(db)

	tests := []struct {
		name    string
		argID   int
		wantErr bool
	}{
		{name: "new user with unique id 123", argID: 123, wantErr: false},
		{name: "new user with unique id 1", argID: 1, wantErr: false},
		{name: "new user with unique id 3", argID: 3, wantErr: false},
		{name: "new user with not unique id 4", argID: 4, wantErr: true},
		{name: "new user with not unique id 5", argID: 5, wantErr: true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			t := &Transaction{tx: tx}

			err = t.NewUser(tt.argID)

			if tt.wantErr {
				assert.Error(t1, err)
				tx.Rollback()
				return

			}

			assert.NoError(t1, err)
			tx.Commit()
			var res pgtype.Int4
			err = db.QueryRow(`SELECT id FROM balances WHERE id = $1`, tt.argID).Scan(&res)
			assert.NoError(t1, err)
			assert.Equal(t1, tt.argID, int(res.Int))

		})
	}
}

func TestTransaction_CommitChanges(t1 *testing.T) {
	db := prepareDB()
	defer cleanup(db)

	type args struct {
		id      int
		balance decimal.Decimal
		ch      wallet.HistoryChange
	}

	amount1, amount2 := decimal.NewFromInt(45), decimal.NewFromInt(1000)

	tests := []struct {
		name string
		args args
	}{

		{name: "transfer money to user 5", args: args{id: 5, balance: decimal.NewFromInt(45),
			ch: wallet.HistoryChange{Date: testTime, Operation: wallet.Replenishment, Amount: amount1, Description: "на проезд"}}},

		{name: "transfer money to user 6", args: args{id: 6, balance: decimal.NewFromInt(45000),
			ch: wallet.HistoryChange{Date: testTime, Operation: wallet.Replenishment, Amount: amount2, Description: "зачисление через банкомат"}}},

		{name: "transfer money from user 5", args: args{id: 5, balance: decimal.NewFromInt(0),
			ch: wallet.HistoryChange{Date: testTime2, Operation: wallet.Withdrawal, Amount: amount1, Description: "на трамвай"}}},

		{name: "transfer money from user 6", args: args{id: 6, balance: decimal.NewFromInt(44000),
			ch: wallet.HistoryChange{Date: testTime2, Operation: wallet.Withdrawal, Amount: amount2, Description: "покупка почки"}}},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			t := &Transaction{tx: tx}

			err = t.CommitChanges(tt.args.id, tt.args.balance, tt.args.ch)
			assert.NoError(t1, err)
			if err == nil {
				t.tx.Commit()
			} else {
				t.tx.Rollback()
				return
			}

			var resBalance decimal.Decimal
			err = db.QueryRow(`SELECT balance FROM balances WHERE id = $1`, tt.args.id).Scan(&resBalance)
			assert.NoError(t1, err)
			assert.Equal(t1, tt.args.balance.String(), resBalance.String())

			var date pgtype.Int8
			var amount decimal.Decimal
			var option, description pgtype.Text
			err = db.QueryRow(`SELECT date, option, amount, description FROM history WHERE wallet_id = $1 ORDER BY date DESC LIMIT 1`, tt.args.id).Scan(&date, &option, &amount, &description)
			resHistory := &wallet.HistoryChange{Date: date.Int, Operation: wallet.Operation(option.String), Amount: amount, Description: description.String}

			assert.NoError(t1, err)
			assert.Equal(t1, tt.args.ch.Date, resHistory.Date)
			assert.Equal(t1, tt.args.ch.Amount.String(), resHistory.Amount.String())
			assert.Equal(t1, tt.args.ch.Description, resHistory.Description)
			assert.Equal(t1, tt.args.ch.Operation, resHistory.Operation)
		})
	}
}

func TestTransaction_GetBalance(t1 *testing.T) {
	db := prepareDB()
	defer cleanup(db)

	balance1 := decimal.NewFromInt(1)
	balance2 := decimal.NewFromInt(0)
	balance3 := decimal.NewFromInt(44000)

	tests := []struct {
		name    string
		argID   int
		want    wallet.Wallet
		wantErr bool
	}{
		{name: "get balance from existing user 4", argID: 4, want: wallet.Wallet{ID: 4, Balance: balance1}, wantErr: false},
		{name: "get balance from existing user 5", argID: 5, want: wallet.Wallet{ID: 5, Balance: balance2}, wantErr: false},
		{name: "get balance from existing user 6", argID: 6, want: wallet.Wallet{ID: 6, Balance: balance3}, wantErr: false},
		{name: "get balance from not existing user 1", argID: 1, wantErr: true},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			t := &Transaction{tx: tx}

			got, err := t.GetBalance(tt.argID)

			if tt.wantErr {
				assert.ErrorIs(t1, err, UserDoesNotExistErr)
				assert.Nil(t1, got)
				t.tx.Rollback()
				return
			}

			assert.NoError(t1, err)
			assert.Equal(t1, tt.want.ID, got.ID)
			assert.Equal(t1, tt.want.History, got.History)
			assert.Equal(t1, tt.want.Balance.String(), got.Balance.String())
			t.tx.Commit()
		})
	}
}

func TestTransaction_GetHistory(t1 *testing.T) {
	db := prepareDB()
	defer cleanup(db)

	type args struct {
		id      int
		orderBy OrderBy
		order   Order
	}

	amount1, amount2 := decimal.NewFromInt(1000), decimal.NewFromInt(1001)

	tests := []struct {
		name    string
		args    args
		want    wallet.Wallet
		wantErr bool
	}{
		{name: "get history of not existed user", args: args{id: 10, orderBy: OrderByAmount, order: Asc}, wantErr: true},

		{name: "get history of existed user order by amount asc", args: args{id: 4, orderBy: OrderByAmount, order: Asc},
			want:    wallet.Wallet{ID: 4, History: []wallet.HistoryChange{{Date: testTime2, Operation: wallet.Withdrawal, Amount: amount1, Description: "почка не подошла"}, {Date: testTime, Operation: wallet.Replenishment, Amount: amount2, Description: "деньги за продажу почки"}}},
			wantErr: false},

		{name: "get history of existed user order by amount desc", args: args{id: 4, orderBy: OrderByAmount, order: Desc},
			want:    wallet.Wallet{ID: 4, History: []wallet.HistoryChange{{Date: testTime, Operation: wallet.Replenishment, Amount: amount2, Description: "деньги за продажу почки"}, {Date: testTime2, Operation: wallet.Withdrawal, Amount: amount1, Description: "почка не подошла"}}},
			wantErr: false},

		{name: "get history of of existed user order by date asc", args: args{id: 4, orderBy: OrderByDate, order: Asc},
			want:    wallet.Wallet{ID: 4, History: []wallet.HistoryChange{{Date: testTime, Operation: wallet.Replenishment, Amount: amount2, Description: "деньги за продажу почки"}, {Date: testTime2, Operation: wallet.Withdrawal, Amount: amount1, Description: "почка не подошла"}}},
			wantErr: false},

		{name: "get history of existed user order by date desc", args: args{id: 4, orderBy: OrderByDate, order: Desc},
			want:    wallet.Wallet{ID: 4, History: []wallet.HistoryChange{{Date: testTime2, Operation: wallet.Withdrawal, Amount: amount1, Description: "почка не подошла"}, {Date: testTime, Operation: wallet.Replenishment, Amount: amount2, Description: "деньги за продажу почки"}}},
			wantErr: false},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			t := &Transaction{tx: tx}

			got, err := t.GetHistory(tt.args.id, tt.args.orderBy, tt.args.order, 100)
			if tt.wantErr {
				assert.Error(t1, err)
				assert.Nil(t1, got)
				t.tx.Rollback()
				return
			}
			assert.NoError(t1, err)
			assert.Equal(t1, tt.want.ID, got.ID)
			assert.Equal(t1, tt.want.Balance.String(), got.Balance.String())

			for i, historyChange := range tt.want.History {
				assert.Equal(t1, historyChange.Date, got.History[i].Date)
				assert.Equal(t1, historyChange.Amount.String(), got.History[i].Amount.String())
				assert.Equal(t1, historyChange.Description, got.History[i].Description)
				assert.Equal(t1, historyChange.Operation, got.History[i].Operation)
			}
		})
	}
}
