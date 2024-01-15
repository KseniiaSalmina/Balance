package app

import (
	"github.com/KseniiaSalmina/Balance/internal/billing"
	"github.com/KseniiaSalmina/Balance/internal/config"
	"github.com/KseniiaSalmina/Balance/internal/database"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/KseniiaSalmina/Balance/internal/api"
)

type Application struct {
	cfg    config.Application
	close  chan os.Signal
	server *api.Server
	db     *database.DB
	bill   *billing.Billing
}

func NewApplication(cfg config.Application) (*Application, error) {
	app := Application{
		cfg: cfg,
	}

	if err := app.bootstrap(); err != nil {
		return nil, err
	}

	app.readyToShutdown()

	return &app, nil
}

func (a *Application) bootstrap() error {
	//init dependencies
	if err := a.initDatabase(); err != nil {
		return err
	}

	//init services
	a.initBilling()

	//init controllers
	if err := a.initServer(); err != nil {
		return err
	}

	return nil
}

func (a *Application) initDatabase() error {
	db, err := database.NewDB(a.cfg.Postgres)
	if err != nil {
		return err
	}

	a.db = db
	return nil
}

func (a *Application) initBilling() {
	a.bill = billing.NewBilling(a.db)
}

func (a *Application) initServer() error {
	s, err := api.NewServer(a.cfg.Server, a.bill)
	if err != nil {
		return err
	}

	a.server = s
	return nil
}

func (a *Application) Run() {
	defer a.stop()

	a.server.Run()

	<-a.close
}

func (a *Application) stop() {
	if err := a.db.Close(); err != nil {
		log.Printf("incorrect closing of database: %s", err.Error())
	} else {
		log.Print("database closed")
	}

	if err := a.server.Shutdown(); err != nil {
		log.Printf("incorrect closing of server: %s", err.Error())
	} else {
		log.Print("server closed")
	}
}

func (a *Application) readyToShutdown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	a.close = ch
}
