package database

import (
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/pgtype"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
	"time"

	"github.com/KseniiaSalmina/Balance/pkg/wallet"
)

var config = pgx.ConnConfig{User: "user", Password: "password", Database: "testdb"} //TODO
var testTime = time.Now().Unix()
var testTime2 = testTime + 1

func prepareDB() *pgx.Conn {
	db, err := pgx.Connect(config)
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`INSERT INTO balances VALUES (4, $1), (5, $2), (6, $3);`, "1", "0", "44000")
	db.Exec(`INSERT INTO history(wallet_id, date, option, amount, description) VALUES(4, $1, $2, $3, $4), (4, $5, $6, $7, $8)`, testTime, wallet.Replenishment, "1001", "деньги за продажу почки", testTime2, wallet.Withdrawal, "1000", "почка не подошла")

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
		balance string
		ch      wallet.Change
	}

	tests := []struct {
		name string
		args args
	}{

		{name: "transfer money to user 5", args: args{id: 5, balance: "45",
			ch: wallet.Change{Date: testTime, Operation: wallet.Replenishment, Amount: "45", Description: "на проезд"}}},

		{name: "transfer money to user 6", args: args{id: 6, balance: "45000",
			ch: wallet.Change{Date: testTime, Operation: wallet.Replenishment, Amount: "1000", Description: "зачисление через банкомат"}}},

		{name: "transfer money from user 5", args: args{id: 5, balance: "0",
			ch: wallet.Change{Date: testTime2, Operation: wallet.Withdrawal, Amount: "45", Description: "на трамвай"}}},

		{name: "transfer money from user 6", args: args{id: 6, balance: "44000",
			ch: wallet.Change{Date: testTime2, Operation: wallet.Withdrawal, Amount: "1000", Description: "покупка почки"}}},
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

			var resBalance string
			err = db.QueryRow(`SELECT balance FROM balances WHERE id = $1`, tt.args.id).Scan(&resBalance)
			assert.NoError(t1, err)
			assert.Equal(t1, tt.args.balance, resBalance)

			var date pgtype.Int8
			var option, amount, description pgtype.Text
			err = db.QueryRow(`SELECT date, option, amount, description FROM history WHERE wallet_id = $1 ORDER BY date DESC LIMIT 1`, tt.args.id).Scan(&date, &option, &amount, &description)
			resHistory := &wallet.Change{Date: date.Int, Operation: wallet.Operation(option.String), Amount: amount.String, Description: description.String}

			assert.NoError(t1, err)
			assert.Equal(t1, tt.args.ch, *resHistory)
		})
	}
}

func TestTransaction_GetBalance(t1 *testing.T) {
	db := prepareDB()
	defer cleanup(db)

	balance1, err := decimal.NewFromString("1")
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
}

func TestTransaction_GetHistory(t1 *testing.T) {
	db := prepareDB()
	defer cleanup(db)

	type args struct {
		id      int
		orderBy OrderBy
		order   Order
	}
	tests := []struct {
		name    string
		args    args
		want    wallet.Wallet
		wantErr bool
	}{
		{name: "get history of not existed user", args: args{id: 10, orderBy: OrderByAmount, order: Asc}, wantErr: true},

		{name: "get history of existed user order by amount asc", args: args{id: 4, orderBy: OrderByAmount, order: Asc},
			want:    wallet.Wallet{ID: 4, History: []wallet.Change{{Date: testTime2, Operation: wallet.Withdrawal, Amount: "1000", Description: "почка не подошла"}, {Date: testTime, Operation: wallet.Replenishment, Amount: "1001", Description: "деньги за продажу почки"}}},
			wantErr: false},

		{name: "get history of existed user order by amount desc", args: args{id: 4, orderBy: OrderByAmount, order: Desc},
			want:    wallet.Wallet{ID: 4, History: []wallet.Change{{Date: testTime, Operation: wallet.Replenishment, Amount: "1001", Description: "деньги за продажу почки"}, {Date: testTime2, Operation: wallet.Withdrawal, Amount: "1000", Description: "почка не подошла"}}},
			wantErr: false},

		{name: "get history of of existed user order by date asc", args: args{id: 4, orderBy: OrderByDate, order: Asc},
			want:    wallet.Wallet{ID: 4, History: []wallet.Change{{Date: testTime, Operation: wallet.Replenishment, Amount: "1001", Description: "деньги за продажу почки"}, {Date: testTime2, Operation: wallet.Withdrawal, Amount: "1000", Description: "почка не подошла"}}},
			wantErr: false},

		{name: "get history of existed user order by date desc", args: args{id: 4, orderBy: OrderByDate, order: Desc},
			want:    wallet.Wallet{ID: 4, History: []wallet.Change{{Date: testTime2, Operation: wallet.Withdrawal, Amount: "1000", Description: "почка не подошла"}, {Date: testTime, Operation: wallet.Replenishment, Amount: "1001", Description: "деньги за продажу почки"}}},
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
			assert.Equal(t1, tt.want, *got)
			t.tx.Commit()
		})
	}
}
