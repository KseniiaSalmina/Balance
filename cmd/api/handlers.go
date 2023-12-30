package api

import (
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"net/http"
	"strconv"

	"github.com/KseniiaSalmina/Balance/pkg/database"
	"github.com/KseniiaSalmina/Balance/pkg/logic"
	"github.com/KseniiaSalmina/Balance/pkg/wallet"
)

func (a *Application) getBalanceHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		http.Error(w, "incorrect wallet ID", http.StatusBadRequest)
		return
	}

	tx, err := a.db.Begin()
	if err != nil {
		http.Error(w, "unable connect to database", http.StatusServiceUnavailable)
		return
	}

	t := database.NewTransaction(tx)

	balance, err := logic.CheckBalance(t, id)
	if err != nil {
		t.Rollback()
		if errors.Is(err, database.UserDoesNotExistErr) {
			http.Error(w, database.UserDoesNotExistErr.Error(), http.StatusBadRequest) //TODO No content?
			return
		}
		http.Error(w, "internal server error, try again", http.StatusInternalServerError)
		return
	}

	t.Commit()

	json.NewEncoder(w).Encode(balance) //TODO
	w.WriteHeader(http.StatusOK)
}

func (a *Application) getHistoryHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		http.Error(w, "incorrect wallet ID", http.StatusBadRequest)
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

	tx, err := a.db.Begin()
	if err != nil {
		http.Error(w, "unable connect to database", http.StatusServiceUnavailable)
		return
	}

	t := database.NewTransaction(tx)

	history, err := logic.CheckHistory(t, id, database.OrderBy(orderBy), database.Order(order), limit)
	if err != nil {
		t.Rollback()
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, database.UserDoesNotExistErr.Error(), http.StatusBadRequest) //TODO No content?
			return
		}
		http.Error(w, "internal server error, try again", http.StatusInternalServerError)
		return
	}

	t.Commit()

	json.NewEncoder(w).Encode(history) //TODO
	w.WriteHeader(http.StatusOK)
}

type changingBalanceRequest struct {
	isTransfer  bool
	to          int
	amount      int
	description string
}

func (a *Application) moneyTransactionHandler(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"] //TODO middleware?
	id, err := strconv.Atoi(idStr)
	if err != nil || id < 0 {
		http.Error(w, "incorrect wallet ID", http.StatusBadRequest)
		return
	}

	var changing changingBalanceRequest
	err = json.NewDecoder(r.Body).Decode(&changing)
	if err != nil {
		http.Error(w, "incorrect wallet data: "+err.Error(), http.StatusBadRequest) //TODO
		return
	} else if changing.amount == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	var operation wallet.Operation
	if !changing.isTransfer {
		if changing.amount > 0 {
			operation = wallet.Replenishment
		} else {
			operation = wallet.Withdrawal
			changing.amount *= -1
		}
	}

	tx, err := a.db.Begin()
	if err != nil {
		http.Error(w, "unable connect to database", http.StatusServiceUnavailable)
		return
	}

	t := database.NewTransaction(tx)

	switch changing.isTransfer {
	case true:
		err = logic.Transfer(t, id, changing.to, strconv.Itoa(changing.amount))
	case false:
		err = logic.MoneyTransaction(t, id, operation, strconv.Itoa(changing.amount), changing.description)
	}

	if err != nil {
		t.Rollback()
		if errors.Is(err, database.UserDoesNotExistErr) {
			http.Error(w, err.Error(), http.StatusBadRequest) //TODO No content?
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = t.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
