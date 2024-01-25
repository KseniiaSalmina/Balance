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

	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
)

// @Summary Get user balance
// @Tags info
// @Description get user balance by id
// @Accept json
// @Produce json
// @Param id path int true "user id"
// @Success 200 {string} string
// @Failure 400 {string} string
// @Failure 500	{string} string
// @Router /wallets/{id}/balance [get]
func (s *Server) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parceID(r)
	if err != nil {
		http.Error(w, "incorrect wallet ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	balance, err := s.bill.CheckBalance(id)
	if err != nil {
		if errors.Is(err, database.UserDoesNotExistErr) {
			http.Error(w, database.UserDoesNotExistErr.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error, try again", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(balance)
}

func parceID(r *http.Request) (int, error) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("parceID -> %w", err)
	}
	if id <= 0 {
		return 0, errors.New("invalid ID")
	}
	return id, nil
}

// @Summary Get user balance history
// @Tags info
// @Description get user transaction history by id
// @Accept json
// @Produce json
// @Param id path int true "user id"
// @Param orderBy query string false "string enums, default: date" Enums(date, amount)
// @Param order query string false "string enums, default: DESC" Enums(DESC, ASC)
// @Param limit query int false "default: 100"
// @Success 200 {array} wallet.HistoryChange
// @Failure 400 {string} string
// @Failure 500	{string} string
// @Router /wallets/{id}/history [get]
func (s *Server) getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parceID(r)
	if err != nil {
		http.Error(w, "incorrect wallet ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	orderBy := r.FormValue("orderBy")
	if orderBy != string(database.OrderByAmount) && orderBy != string(database.OrderByDate) {
		orderBy = string(database.OrderByDate)
	}

	order := r.FormValue("order")
	if order != string(database.Desc) && order != string(database.Asc) {
		order = string(database.Desc)
	}

	limitStr := r.FormValue("limit")
	limit, err := strconv.Atoi(limitStr)
	if err != nil && limitStr != "" {
		http.Error(w, "incorrect limit", http.StatusBadRequest)
		return
	}
	if limitStr == "" {
		limit = 100
	}

	history, err := s.bill.CheckHistory(id, database.OrderBy(orderBy), database.Order(order), limit)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, database.UserDoesNotExistErr.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error, try again", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(history)
}

// @Summary Change user balance
// @Tags changing
// @Description produce transaction to change user balance. Support replenishment, withdrawal and transfer between users
// @Accept json
// @Param id path int true "user id"
// @Param input body api.ChangingBalanceRequest true "info about transaction"
// @Success 200
// @Failure 400 {string} string
// @Failure 500	{string} string
// @Router /wallets/{id}/transaction [patch]
func (s *Server) moneyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	id, err := parceID(r)
	if err != nil {
		http.Error(w, "incorrect wallet ID: "+err.Error(), http.StatusBadRequest)
		return
	}

	var changing ChangingBalanceRequest
	err = json.NewDecoder(r.Body).Decode(&changing)
	if err != nil {
		http.Error(w, "incorrect wallet data: "+err.Error(), http.StatusBadRequest)
		return
	}

	if changing.Amount.IsZero() {
		w.WriteHeader(http.StatusOK)
		return
	}

	if changing.To == 0 {
		http.Error(w, "required recipient", http.StatusBadRequest)
		return
	}

	var operation wallet.Operation
	if !changing.IsTransfer {
		if changing.Amount.IsPositive() {
			operation = wallet.Replenishment
		} else {
			operation = wallet.Withdrawal
			changing.Amount = changing.Amount.Mul(decimal.NewFromInt(-1))
		}
	}

	switch changing.IsTransfer {
	case true:
		err = s.bill.Transfer(id, changing.To, changing.Amount)
	case false:
		if changing.Description == "" {
			http.Error(w, "required description", http.StatusBadRequest)
			return
		}
		err = s.bill.MoneyTransaction(id, operation, changing.Amount, changing.Description)
	}

	if err != nil {
		if errors.Is(err, database.UserDoesNotExistErr) || errors.Is(err, wallet.InsufficientFundsErr) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
