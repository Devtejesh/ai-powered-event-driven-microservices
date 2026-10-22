package config

import (
	"database/sql"
	"fmt"
	"os"
	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	password := os.Getenv("DB_PASSWORD")
dsn := fmt.Sprintf("host=localhost port=5432 user=postgres password=%s", password)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	fmt.Println("Database connected successfully")

	return db, nil
}
