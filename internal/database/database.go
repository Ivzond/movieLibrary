package database

import (
	"database/sql"
	"fmt"
	"movieLibrary/internal/pkg/helpers"
)

const (
	DBHost     = "localhost"
	DBPort     = "5432"
	DBUser     = "postgres"
	DBPassword = "12345678"
	DBName     = "movieLibrary"
)

func InitDB() (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		DBHost,
		DBPort,
		DBUser,
		DBName,
		DBPassword,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		helpers.ErrorLogger.Println("Error on connection to database:", err)
		return nil, err
	}
	return db, err
}
