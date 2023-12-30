package api

import (
	"errors"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

type Application struct {
	db             *pgx.Conn
	wg             *sync.WaitGroup
	listener       net.Listener
	handler        *mux.Router
	close          <-chan struct{}
	waitForClosing *sync.WaitGroup
}

func NewApplication(config pgx.ConnConfig, protocol, address string) (*Application, error) {
	db, err := pgx.Connect(config)
	if err != nil {
		return nil, errors.New("cannot connect to database")
	}

	listener, err := net.Listen(protocol, address)
	if err != nil {
		return nil, errors.New("cannot connect to the network")
	}

	app := Application{db: db, listener: listener}

	router := mux.NewRouter()
	router.Name("get_balance").Methods(http.MethodGet).Path("/wallets/{id}/balance").HandlerFunc(app.getBalanceHandler)
	router.Name("get_history").Methods(http.MethodGet).Path("/wallets/{id}/history?orderBy={orderBy}&order={order}&limit={limit}").HandlerFunc(app.getHistoryHandler)
	router.Name("transaction").Methods(http.MethodPatch).Path("wallets/{id}/transaction").HandlerFunc(app.moneyTransactionHandler)

	app.handler = router
	app.readyToShutdown()

	return &app, nil
}

func (a *Application) Run() {

	for {
		select {
		case <-a.close:
			a.waitForClosing.Wait()
			return
		default:
		}
		a.wg.Add(1)
		http.Serve(a.listener, a.handler) //TODO
	}
}

func (a *Application) readyToShutdown() {
	closeCh := make(chan struct{})
	a.close = closeCh

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-ch
		signal.Stop(ch)
		a.waitForClosing.Add(1)
		close(closeCh)
		a.wg.Wait()
		defer a.waitForClosing.Done()
		defer close(ch)
		defer a.db.Close()
	}()
}
