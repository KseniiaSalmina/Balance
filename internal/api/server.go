package api

import (
	"context"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"

	"github.com/KseniiaSalmina/Balance/internal/config"
	"github.com/KseniiaSalmina/Balance/internal/database"
	"github.com/KseniiaSalmina/Balance/internal/wallet"
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
	router.Name("get_history").Methods(http.MethodGet).Path("/wallets/{id}/history").HandlerFunc(s.getHistoryHandler)
	router.Name("transaction").Methods(http.MethodPatch).Path("/wallets/{id}/transaction").HandlerFunc(s.moneyTransactionHandler)

	swagHandler := httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	)
	router.Methods(http.MethodGet).PathPrefix("/swagger").HandlerFunc(swagHandler)

	s.httpServer = &http.Server{
		Addr:         cfg.Listen,
		Handler:      router,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}
	return s, nil
}

func (s *Server) Run() {
	log.Println("server started")

	go func() {
		err := s.httpServer.ListenAndServe()
		log.Printf("http server stopped: %s", err.Error())
	}()
}

func (s *Server) Shutdown() error {
	return s.httpServer.Shutdown(context.Background())
}
