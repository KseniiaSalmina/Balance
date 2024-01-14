package api

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"net"
	"net/http"
)

type Server struct {
	db       *pgx.Conn
	listener net.Listener
	handler  *mux.Router
}

func NewServer(cfg pgx.ConnConfig, protocol, address string) (*Server, error) {
	db, err := pgx.Connect(cfg)
	if err != nil {
		return nil, errors.New("cannot connect to database")
	}

	listener, err := net.Listen(protocol, address)
	if err != nil {
		return nil, errors.New("cannot connect to network")
	}

	s := &Server{db: db, listener: listener}

	router := mux.NewRouter()
	router.Name("get_balance").Methods(http.MethodGet).Path("/wallets/{id}/balance").HandlerFunc(s.getBalanceHandler)
	router.Name("get_history").Methods(http.MethodGet).Path("/wallets/{id}/history?orderBy={orderBy}&order={order}&limit={limit}").HandlerFunc(s.getHistoryHandler)
	router.Name("transaction").Methods(http.MethodPatch).Path("wallets/{id}/transaction").HandlerFunc(s.moneyTransactionHandler)

	s.handler = router
	return s, nil
}

func (s *Server) Run() {
	http.Serve(s.listener, s.handler)
}

func (s *Server) Shutdown() {
	s.db.Close()
	s.listener.Close()
}
