package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"

	_ "github.com/KseniiaSalmina/Balance/docs"
	app "github.com/KseniiaSalmina/Balance/internal"
	"github.com/KseniiaSalmina/Balance/internal/config"
)

var (
	cfg config.Application
)

func init() {
	_ = godotenv.Load(".env")
	if err := env.Parse(&cfg); err != nil {
		panic(err)
	}
}

// @title Balance management API
// @version 1.0.0
// @description API to manage users balances
// @host localhost:8088
// @BasePath /
func main() {
	application, err := app.NewApplication(cfg)
	if err != nil {
		log.Fatal(err)
	}
	application.Run()
}
