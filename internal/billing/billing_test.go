package billing

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

func TestCheckBalance(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		want    string
		wantErr bool
	}{
		{name: "check balance of existing user", id: 10, want: "300", wantErr: false},
		{name: "check balance of not existing user", id: -10, want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckBalance(nil, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCheckHistory(t *testing.T) {
	tests := []struct {
		name           string
		id             int
		limit          int
		wantErr        bool
		expectedLength int
	}{
		{name: "expected data: database got 100 returns 100 notes", id: 100, limit: 100, wantErr: false, expectedLength: 100},
		{name: "unexpected data: user does not exist or have ero balance", id: -5, limit: 100, wantErr: true, expectedLength: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckHistory(nil, tt.id, database.OrderByDate, database.Desc, tt.limit)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLength, len(got))
			}
		})
	}
}

func TestMoneyTransaction(t *testing.T) {
	type args struct {
		id     int
		opt    wallet.Operation
		amount decimal.Decimal
		desc   string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		expectedErr error
	}{
		{name: "withdrawal: user does not exist", args: args{id: 0, opt: wallet.Withdrawal, amount: decimal.NewFromInt(200), desc: "advertising purchase"}, wantErr: true, expectedErr: database.UserDoesNotExistErr},
		{name: "replenishment: user does not exist", args: args{id: 0, opt: wallet.Replenishment, amount: decimal.NewFromInt(1000), desc: "bribe"}, wantErr: false},
		{name: "replenishment: user exist", args: args{id: 100, opt: wallet.Replenishment, amount: decimal.NewFromInt(100), desc: "donation"}, wantErr: false},
		{name: "withdrawal: user exist, insufficient funds", args: args{id: 456, opt: wallet.Withdrawal, amount: decimal.NewFromInt(4600), desc: "buying phone"}, wantErr: true, expectedErr: wallet.InsufficientFundsErr},
		{name: "withdrawal: user exist", args: args{id: 5000, opt: wallet.Withdrawal, amount: decimal.NewFromInt(60), desc: "buying cake"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MoneyTransaction(nil, tt.args.id, tt.args.opt, tt.args.amount, tt.args.desc)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestTransfer(t *testing.T) {
	type args struct {
		from   int
		to     int
		amount decimal.Decimal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "successful transfer", args: args{from: 456, to: 123, amount: decimal.NewFromInt(120)}, wantErr: false},
		{name: "unsuccessful transfer: insufficient funds", args: args{from: 123, to: 456, amount: decimal.NewFromInt(400)}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Transfer(nil, tt.args.from, tt.args.to, tt.args.amount)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, wallet.InsufficientFundsErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
