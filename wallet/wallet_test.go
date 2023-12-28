package wallet

import (
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewChange(t *testing.T) {
	startTime := time.Now().Unix()

	type args struct {
		opt    Operation
		amount string
		descr  string
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "test change 1", args: args{opt: Withdrawal, amount: "300", descr: "test description 1"}},
		{name: "test change 2", args: args{opt: Replenishment, amount: "3", descr: "test description 2"}},
		{name: "test change 3", args: args{opt: Withdrawal, amount: "0.03", descr: "test description 3"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewChange(tt.args.opt, tt.args.amount, tt.args.descr)

			assert.LessOrEqual(t, startTime, got.Date)
			assert.Equal(t, tt.args.opt, got.Operation)
			assert.Equal(t, tt.args.amount, got.Amount)
			assert.Equal(t, tt.args.descr, got.Description)
		})
	}
}

func TestWallet_ChangeBalance(t *testing.T) {
	testBalance1, _ := decimal.NewFromString("300")
	testBalance2, _ := decimal.NewFromString("3")
	testBalance3, _ := decimal.NewFromString("0.03")

	type args struct {
		amount string
		opt    Operation
	}
	tests := []struct {
		name            string
		balance         decimal.Decimal
		args            args
		wantErr         bool
		expectedErr     error
		expectedBalance decimal.Decimal
	}{
		{name: "incorrect amount", balance: testBalance3, args: args{amount: "two", opt: Replenishment}, wantErr: true, expectedBalance: testBalance3},
		{name: "incorrect operation", balance: testBalance3, args: args{amount: "0", opt: "ограбление"}, wantErr: true, expectedBalance: testBalance3},
		{name: "insufficient funds", balance: testBalance2, args: args{amount: "10", opt: Withdrawal}, wantErr: true, expectedErr: InsufficientFundsErr, expectedBalance: testBalance2},
		{name: "successful withdrawal", balance: testBalance1, args: args{amount: "297", opt: Withdrawal}, wantErr: false, expectedBalance: testBalance2},
		{name: "successful replenishment", balance: testBalance2, args: args{amount: "297", opt: Replenishment}, wantErr: false, expectedBalance: testBalance1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Wallet{Balance: tt.balance}
			err := w.ChangeBalance(tt.args.amount, tt.args.opt)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.Equal(t, InsufficientFundsErr, err)
				}
			} else {
				assert.NoError(t, err)
			}

			if !w.Balance.Equal(tt.expectedBalance) {
				t.Errorf("uncorrect balance: want %v, got %v", tt.expectedBalance, w.Balance)
			}
		})
	}
}

func TestWallet_StringBalance(t *testing.T) {
	testBalance1, _ := decimal.NewFromString("300")
	testBalance2, _ := decimal.NewFromString("3")
	testBalance3, _ := decimal.NewFromString("0.03")

	tests := []struct {
		name    string
		balance decimal.Decimal
		want    string
	}{
		{name: "test balance 1", balance: testBalance1, want: "300"},
		{name: "test balance 2", balance: testBalance2, want: "3"},
		{name: "test balance 3", balance: testBalance3, want: "0.03"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &Wallet{Balance: tt.balance}
			got := w.StringBalance()
			assert.Equal(t, tt.want, got)
		})
	}
}
