package config

import (
	"database/sql"
	"fmt"
)

func ConnectDB() (*sql.DB, error) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=postgres password=postgres123 dbname=order_db sslmode=disable")
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}
	fmt.Println("Database connected successfully")
	return db, err
}
