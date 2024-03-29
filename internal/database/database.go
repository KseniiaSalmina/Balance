package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx"
	"log"
	"time"

	"github.com/KseniiaSalmina/Balance/internal/config"
)

type DB struct {
	db *pgx.Conn
}

func NewDB(cfg config.Postgres) (*DB, error) {
	config := pgx.ConnConfig{
		User:     cfg.User,
		Password: cfg.Password,
		Database: cfg.Database,
		Host:     cfg.Host,
		Port:     uint16(cfg.Port),
	}

	db, err := pgx.Connect(config)
	if err != nil {
		return nil, errors.New("cannot connect to database")
	}

	ctx, cansel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cansel()
	if err := db.Ping(ctx); err != nil {
		return nil, errors.New("cannot connect to database: ping fail")
	}

	log.Println("database connected")

	return &DB{
		db: db,
	}, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) NewTransaction() (*Transaction, error) {
	tx, err := db.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("NewTransaction -> %w", err)
	}
	return &Transaction{tx: tx}, nil
}

// OrderBy can be date or amount
type OrderBy string

const (
	OrderByDate   OrderBy = "date"
	OrderByAmount OrderBy = "amount"
)

// Order can be DESC or ASC
type Order string

const (
	Desc Order = "DESC"
	Asc  Order = "ASC"
)

var UserDoesNotExistErr error = errors.New("user does not exist")
