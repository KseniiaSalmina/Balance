package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"github.com/shopspring/decimal"
	"net/http"
	"strconv"

	"github.com/KseniiaSalmina/Balance/internal/billing"
	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

func (s *Server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parceID(r)
	if err != nil {
		http.Error(w, "incorrect wallet ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	balance, err := billing.CheckBalance(s.db, id)
	if err != nil {
		if errors.Is(err, database.UserDoesNotExistErr) {
			http.Error(w, database.UserDoesNotExistErr.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error, try again", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(balance)
	w.WriteHeader(http.StatusOK)
}

func parceID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("parceID -> %w", err)
	}
	if id < 0 {
		return 0, errors.New("invalid ID")
	}
	return id, nil
}

func (s *Server) getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parceID(r)
	if err != nil {
		http.Error(w, "incorrect wallet ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderBy := mux.Vars(r)["orderBy"]
	if orderBy != string(database.OrderByAmount) && orderBy != string(database.OrderByDate) {
		orderBy = string(database.OrderByDate)
	}

	order := mux.Vars(r)["order"]
	if order != string(database.Desc) && order != string(database.Asc) {
		order = string(database.Desc)
	}

	limitStr := mux.Vars(r)["limit"]
	limit, err := strconv.Atoi(limitStr)
	if err != nil && limitStr != "" {
		http.Error(w, "incorrect limit", http.StatusBadRequest)
		return
	}
	if limitStr == "" {
		limit = 100
	}

	history, err := billing.CheckHistory(s.db, id, database.OrderBy(orderBy), database.Order(order), limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, database.UserDoesNotExistErr.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error, try again", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(history)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) moneyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parceID(r)
	if err != nil {
		http.Error(w, "incorrect wallet ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	var changing changingBalanceRequest
	err = json.NewDecoder(r.Body).Decode(&changing)
	if err != nil {
		http.Error(w, "incorrect wallet data: "+err.Error(), http.StatusBadRequest)
		return
	}
	if changing.amount.IsZero() {
		w.WriteHeader(http.StatusOK)
		return
	}

	var operation wallet.Operation
	if !changing.isTransfer {
		if changing.amount.IsPositive() {
			operation = wallet.Replenishment
		} else {
			operation = wallet.Withdrawal
			changing.amount = changing.amount.Mul(decimal.NewFromInt(-1))
		}
	}

	switch changing.isTransfer {
	case true:
		err = billing.Transfer(s.db, id, changing.to, changing.amount)
	case false:
		err = billing.MoneyTransaction(s.db, id, operation, changing.amount, changing.description)
	}

	if err != nil {
		if errors.Is(err, database.UserDoesNotExistErr) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
