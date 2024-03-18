package api

import (
	"database/sql"
	"net/http"
)

func StartApi(db *sql.DB) http.Handler {
	router := http.NewServeMux()

	router.HandleFunc("/actors/create", BasicAuthMiddleware(db, createActorHandler(db)))
	router.HandleFunc("/actors/update", BasicAuthMiddleware(db, updateActorHandler(db)))
	router.HandleFunc("/actors/delete", BasicAuthMiddleware(db, deleteActorHandler(db)))
	router.HandleFunc("/actors", BasicAuthMiddleware(db, getActorsHandler(db)))

	router.HandleFunc("/movies/create", BasicAuthMiddleware(db, createMovieHandler(db)))
	router.HandleFunc("/movies/update", BasicAuthMiddleware(db, updateMovieHandler(db)))
	router.HandleFunc("/movies/delete", BasicAuthMiddleware(db, deleteMovieHandler(db)))
	router.HandleFunc("/movies", BasicAuthMiddleware(db, getMoviesHandler(db)))
	router.HandleFunc("/movies/search", BasicAuthMiddleware(db, searchMoviesHandler(db)))

	return router
}
