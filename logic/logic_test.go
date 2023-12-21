package logic

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckBalance(t *testing.T) {
	mock := MockDb{}
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
			got, err := CheckBalance(&mock, tt.id)
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
	mock := MockDb{}
	tests := []struct {
		name           string
		id             int
		wantErr        bool
		expectedLength int
	}{
		{name: "expected data: database returns 100 notes", id: 100, wantErr: false, expectedLength: 10},
		{name: "expected data: database returns less than 100 notes", id: 34, wantErr: false, expectedLength: 4},
		{name: "unexpected data: user does not exist or have ero balance", id: -5, wantErr: true, expectedLength: 0},
		{name: "unexpected data: database returns more than 100 notes", id: 120, wantErr: false, expectedLength: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckHistory(&mock, tt.id, OrderByDate, Desc)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedLength, len(got))
		})
	}
}

func TestMoneyTransaction(t *testing.T) {
	mock := MockDb{}
	type args struct {
		id     int
		opt    Operation
		amount string
		desc   string
	}
	tests := []struct {
		name        string
		args        args
		wantErr     bool
		expectedErr error
	}{
		{name: "withdrawal: user does not exist", args: args{id: 0, opt: Withdrawal, amount: "200", desc: "advertising purchase"}, wantErr: true, expectedErr: UserDoesNotExistErr},
		{name: "replenishment: user does not exist", args: args{id: 0, opt: Replenishment, amount: "1000", desc: "bribe"}, wantErr: false},
		{name: "replenishment: user exist", args: args{id: 100, opt: Replenishment, amount: "100", desc: "donation"}, wantErr: false},
		{name: "withdrawal: user exist, insufficient funds", args: args{id: 456, opt: Withdrawal, amount: "4600", desc: "buying phone"}, wantErr: true, expectedErr: InsufficientFundsErr},
		{name: "withdrawal: user exist", args: args{id: 5000, opt: Withdrawal, amount: "4600", desc: "buying phone"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := MoneyTransaction(&mock, tt.args.id, tt.args.opt, tt.args.amount, tt.args.desc)
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
	mock := MockDb{}
	type args struct {
		from   int
		to     int
		amount string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "successful transfer", args: args{from: 1000, to: 100, amount: "400"}, wantErr: false},
		{name: "unsuccessful transfer: insufficient funds", args: args{from: 1000, to: 100, amount: "1200"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Transfer(&mock, tt.args.from, tt.args.to, tt.args.amount)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, InsufficientFundsErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
