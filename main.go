package main

import (
	app "github.com/KseniiaSalmina/Balance/internal"
	"github.com/KseniiaSalmina/Balance/internal/config"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"log"
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

func main() {
	application, err := app.NewApplication(cfg)
	if err != nil {
		log.Fatal(err)
	}
	application.Run()
}
