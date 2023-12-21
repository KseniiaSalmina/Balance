package main

import (
	"fmt"
	"github.com/jackc/pgx"
	"log"
)

func main() {
	config := pgx.ConnConfig{User: "kseniia", Password: "Efbcnwww1", Database: "testdb"}
	db, err := pgx.Connect(config)
	if err != nil {
		fmt.Println(err)
		log.Fatal("pisos db")
	}
	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		log.Fatal("pisos tx")
	}
	t := Transaction{tx: tx}
	fmt.Println(t.NewUser(1))
	fmt.Println(t.NewUser(3))
	fmt.Println(t.NewUser(2))
	t.tx.Commit()

	db.Close()
	fmt.Println("success")
}
