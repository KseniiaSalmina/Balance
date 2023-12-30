package main

import (
	"github.com/jackc/pgx"
	"log"

	"github.com/KseniiaSalmina/Balance/cmd/api"
)

func main() {
	config := pgx.ConnConfig{User: "user", Password: "password", Database: "testdb"} // TODO
	protocol, address := "protocol", "IP:port"                                       // TODO

	app, err := api.NewApplication(config, protocol, address)
	if err != nil {
		log.Fatal(err)
	}
	app.Run()
}
