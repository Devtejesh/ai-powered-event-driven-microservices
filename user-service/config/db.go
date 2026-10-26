package config

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	_ "github.com/lib/pq"
)

func ConnectDB() (*sql.DB, error) {
	host := os.Getenv("PSQL_HOST")
	port := os.Getenv("PSQL_PORT")
	user := os.Getenv("PSQL_USER")
	password := os.Getenv("PSQL_PASS")
	dbname := os.Getenv("PSQL_DBNAME")
	sslmode := os.Getenv("PSQL_SSLMODE")

	if host == "" || port == "" || user == "" ||
		password == "" || dbname == "" || sslmode == "" {
		return nil, fmt.Errorf("missing required database environment variables")
	}

	if _, err := strconv.Atoi(port); err != nil {
		return nil, fmt.Errorf("invalid PSQL_PORT: %w", err)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	fmt.Println("User Service database connected successfully")

	return db, nil
}
