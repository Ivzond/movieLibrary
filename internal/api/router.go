package api

import (
	"database/sql"
	"net/http"
)

func StartApi(db *sql.DB) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/actors/create", createActorHandler(db))
	router.HandleFunc("/actors/update", updateActorHandler(db))
	router.HandleFunc("/actors/delete", deleteActorHandler(db))
	router.HandleFunc("/actors", getActorsHandler(db))

	router.HandleFunc("/movies/create", createMovieHandler(db))
	router.HandleFunc("/movies/update", updateMovieHandler(db))
	router.HandleFunc("/movies/delete", deleteMovieHandler(db))
	router.HandleFunc("/movies", getMoviesHandler(db))
	router.HandleFunc("/movies/search", searchMoviesHandler(db))

	return router
}
