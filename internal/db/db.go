package db

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect(databaseURL string) error {
	var err error
	DB, err = sql.Open("postgres", databaseURL)
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(20)
	DB.SetMaxIdleConns(5)
	if err := DB.Ping(); err != nil {
		return err
	}
	log.Println("[db] connected")
	return nil
}

func Migrate(schema string) error {
	_, err := DB.Exec(schema)
	if err != nil {
		return err
	}
	log.Println("[db] migration applied")
	return nil
}
