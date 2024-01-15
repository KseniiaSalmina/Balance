package api

import (
	"context"
	"github.com/KseniiaSalmina/Balance/internal/config"
	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	"log"
	"net/http"
)

type BillingManager interface {
	MoneyTransaction(id int, opt wallet.Operation, amount decimal.Decimal, desc string) error
	Transfer(from, to int, amount decimal.Decimal) error
	CheckBalance(id int) (string, error)
	CheckHistory(id int, orderBy database.OrderBy, order database.Order, limit int) ([]wallet.HistoryChange, error)
}

type Server struct {
	bill       BillingManager
	httpServer *http.Server
}

func NewServer(cfg config.Server, bill BillingManager) (*Server, error) {
	s := &Server{
		bill: bill,
	}

	router := mux.NewRouter()
	router.Name("get_balance").Methods(http.MethodGet).Path("/wallets/{id}/balance").HandlerFunc(s.getBalanceHandler)
	router.Name("get_history").Methods(http.MethodGet).Path("/wallets/{id}/history?orderBy={orderBy}&order={order}&limit={limit}").HandlerFunc(s.getHistoryHandler)
	router.Name("transaction").Methods(http.MethodPatch).Path("wallets/{id}/transaction").HandlerFunc(s.moneyTransactionHandler)

	s.httpServer = &http.Server{
		Addr:         cfg.Listen,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return s, nil
}

func (s *Server) Run() {
	go func() {
		err := s.httpServer.ListenAndServe()
		log.Printf("http server stopped: %s", err.Error())
	}()
}

func (s *Server) Shutdown() error {
	return s.httpServer.Shutdown(context.Background())
}
