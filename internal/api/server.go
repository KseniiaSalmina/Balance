package api

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/shopspring/decimal"
	httpSwagger "github.com/swaggo/http-swagger"
	"log"
	"net/http"

	_ "github.com/KseniiaSalmina/Balance/docs"
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
	router.Name("get_history").Methods(http.MethodGet).Path("/wallets/{id}/history?orderBy={orderBy:[A-z]*(?=&)}&order={order:[A-z]*(?=&)}&limit={limit}").HandlerFunc(s.getHistoryHandler)
	router.Name("transaction").Methods(http.MethodPatch).Path("/wallets/{id}/transaction").HandlerFunc(s.moneyTransactionHandler)

	swagConnString := fmt.Sprintf("http://localhost%s/swagger/doc.json", cfg.Listen)
	router.PathPrefix("/swagger").Handler(httpSwagger.Handler(
		httpSwagger.URL(swagConnString),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("#swagger-ui"),
	))

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
