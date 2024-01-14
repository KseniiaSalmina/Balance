package main

import (
	app "github.com/KseniiaSalmina/Balance/internal"
	"github.com/jackc/pgx"
	"log"
)

func main() {
	config := pgx.ConnConfig{User: "user", Password: "password", Database: "testdb"} // TODO
	protocol, address := "protocol", "IP:port"                                       // TODO

	application, err := app.NewApplication(config, protocol, address)
	if err != nil {
		log.Fatal(err)
	}
	application.Run()
}
