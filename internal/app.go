package app

import (
	"fmt"
	"github.com/jackc/pgx"
	"os"
	"os/signal"
	"syscall"

	"github.com/KseniiaSalmina/Balance/internal/api"
)

type Application struct {
	close  chan os.Signal
	server *api.Server
}

func NewApplication(cfg pgx.ConnConfig, protocol, address string) (*Application, error) {
	app := Application{}

	server, err := api.NewServer(cfg, protocol, address)
	if err != nil {
		return nil, fmt.Errorf("NewApplication -> %w", err)
	}

	app.server = server
	app.readyToShutdown()

	return &app, nil
}

func (a *Application) Run() {
	a.server.Run()

	<-a.close
	a.server.Shutdown()
}

func (a *Application) readyToShutdown() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	a.close = ch
}
