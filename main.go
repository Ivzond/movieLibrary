package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"movieLibrary/internal/api"
	"movieLibrary/internal/database"
	"movieLibrary/internal/pkg/helpers"
	"net/http"
)

func main() {
	helpers.InitLogger()

	db, err := database.InitDB()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Error closing database: %v", err)
		}
	}(db)

	router := api.StartApi(db)
	log.Println("App is working on port :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
