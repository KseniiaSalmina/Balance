package database

import (
	"github.com/jackc/pgx"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"
)

var config = pgx.ConnConfig{User: "kseniia", Password: "Efbcnwww1", Database: "testdb"}
var testTime = time.Now().Unix()

func TestTransaction_NewUser(t1 *testing.T) {
	db, err := pgx.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	tests := []struct {
		name    string
		argID   int
		wantErr bool
	}{
		{name: "new user with unique id 1", argID: 123, wantErr: false},
		{name: "new user with unique id 2", argID: 1, wantErr: false},
		{name: "new user with unique id 3", argID: 3, wantErr: false},
		{name: "new user with not unique id 1", argID: 123, wantErr: true},
		{name: "new user with not unique id 2", argID: 1, wantErr: true},
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
			var res int
			err = db.QueryRow(`SELECT id FROM balances WHERE id = $1`, tt.argID).Scan(&res)
			assert.NoError(t1, err)
			assert.Equal(t1, tt.argID, res)

		})
	}

	db.Close()
}

func TestTransaction_CommitChanges(t1 *testing.T) {
	db, err := pgx.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	type args struct {
		id      int
		balance string
		ch      Change
	}

	tests := []struct {
		name string
		args args
	}{
		{name: "transfer money to user 1", args: args{id: 1, balance: "1000.456",
			ch: Change{Date: testTime, Operation: Replenishment, Amount: "1000.456", Description: "деньги за продажу почки"}}},

		{name: "transfer money to user 123", args: args{id: 123, balance: "45",
			ch: Change{Date: testTime, Operation: Replenishment, Amount: "45", Description: "на проезд"}}},

		{name: "transfer money to user 3", args: args{id: 123, balance: "45000",
			ch: Change{Date: testTime, Operation: Replenishment, Amount: "45000", Description: "зачисление через банкомат"}}},

		{name: "transfer money from user 1", args: args{id: 1, balance: "0.456",
			ch: Change{Date: testTime, Operation: Withdrawal, Amount: "1000", Description: "почка не подошла"}}},

		{name: "transfer money from user 123", args: args{id: 123, balance: "0",
			ch: Change{Date: testTime, Operation: Withdrawal, Amount: "45", Description: "на трамвай"}}},

		{name: "transfer money from user 3", args: args{id: 3, balance: "44000",
			ch: Change{Date: testTime, Operation: Withdrawal, Amount: "1000", Description: "покупка почки"}}},
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
			if err != nil {
				t.tx.Commit()
			} else {
				t.tx.Rollback()
				return
			}

			var resBalance string
			err = db.QueryRow(`SELECT balance FROM balances WHERE id = $1`, tt.args.id).Scan(&resBalance)
			assert.NoError(t1, err)
			assert.Equal(t1, tt.args.balance, resBalance)

			resHistory := &Change{}
			err = db.QueryRow(`SELECT date, option, amount, description FROM history WHERE wallet_id = $1 ORDER BY date DESC LIMIT 1`, tt.args.id).Scan(resHistory.Date, resHistory.Operation, resHistory.Amount, resHistory.Description)
			assert.NoError(t1, err)
			assert.Equal(t1, tt.args.ch, *resHistory)
		})
	}

	db.Close()
}

func TestTransaction_GetBalance(t1 *testing.T) {
	db, err := pgx.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	balance1, err := decimal.NewFromString("0.456")
	if err != nil {
		log.Fatal(err)
	}
	balance2, err := decimal.NewFromString("0")
	if err != nil {
		log.Fatal(err)
	}
	balance3, err := decimal.NewFromString("44000")
	if err != nil {
		log.Fatal(err)
	}

	tests := []struct {
		name    string
		argID   int
		want    Wallet
		wantErr bool
	}{
		{name: "get balance from existing user 1", argID: 1, want: Wallet{ID: 1, Balance: balance1}, wantErr: false},
		{name: "get balance from existing user 2", argID: 123, want: Wallet{ID: 123, Balance: balance2}, wantErr: false},
		{name: "get balance from existing user 3", argID: 3, want: Wallet{ID: 3, Balance: balance3}, wantErr: false},
		{name: "get balance from not existing user 1", argID: 4, wantErr: true},
		{name: "get balance from not existing user 2", wantErr: true},
		{name: "get balance from not existing user 3", wantErr: true},
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
				assert.Equal(t1, err, UserDoesNotExistErr)
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

	db.Close()
}

func TestTransaction_GetHistory(t1 *testing.T) {
	db, err := pgx.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	type args struct {
		id      int
		orderBy OrderBy
		order   Order
	}
	tests := []struct {
		name    string
		args    args
		want    Wallet
		wantErr bool
	}{
		{name: "get history of not existed user", args: args{id: 10, orderBy: OrderByAmount, order: Asc}, wantErr: true},

		{name: "get history of existed user order by amount asc", args: args{id: 1, orderBy: OrderByAmount, order: Asc},
			want:    Wallet{ID: 1, History: []Change{{Date: testTime, Operation: Withdrawal, Amount: "1000", Description: "почка не подошла"}, {Date: testTime, Operation: Replenishment, Amount: "1000.456", Description: "деньги за продажу почки"}}},
			wantErr: false},

		{name: "get history of existed user order by amount desc", args: args{id: 1, orderBy: OrderByAmount, order: Desc},
			want:    Wallet{ID: 1, History: []Change{{Date: testTime, Operation: Replenishment, Amount: "1000.456", Description: "деньги за продажу почки"}, {Date: testTime, Operation: Withdrawal, Amount: "1000", Description: "почка не подошла"}}},
			wantErr: false},

		{name: "get history of of existed user order by date asc", args: args{id: 1, orderBy: OrderByDate, order: Asc},
			want:    Wallet{ID: 1, History: []Change{{Date: testTime, Operation: Replenishment, Amount: "1000.456", Description: "деньги за продажу почки"}, {Date: testTime, Operation: Withdrawal, Amount: "1000", Description: "почка не подошла"}}},
			wantErr: false},

		{name: "get history of existed user order by date desc", args: args{id: 1, orderBy: OrderByDate, order: Desc},
			want:    Wallet{ID: 1, History: []Change{{Date: testTime, Operation: Withdrawal, Amount: "1000", Description: "почка не подошла"}, {Date: testTime, Operation: Replenishment, Amount: "1000.456", Description: "деньги за продажу почки"}}},
			wantErr: false},
	}

	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			tx, err := db.Begin()
			if err != nil {
				log.Fatal(err)
			}
			t := &Transaction{tx: tx}

			got, err := t.GetHistory(tt.args.id, tt.args.orderBy, tt.args.order)
			if tt.wantErr {
				assert.Error(t1, err)
				assert.Nil(t1, got)
				t.tx.Rollback()
				return
			}
			assert.NoError(t1, err)
			assert.Equal(t1, tt.want, *got)
			t.tx.Commit()
		})
	}

	db.Close()
	//	defer db.Exec(`TRUNCATE TABLE history`)
	//	defer db.Exec(`TRUNCATE TABLE balances`)
}
