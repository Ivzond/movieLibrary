package api

import (
	"database/sql"
	"encoding/json"
	"log"
	helpers2 "movieLibrary/internal/pkg/helpers"
	"movieLibrary/internal/pkg/validation"
	"net/http"
	"strconv"
	"strings"
)

type MovieRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	ReleaseDate string   `json:"release_date,omitempty"`
	Rating      string   `json:"rating,omitempty"`
	Actors      []string `json:"actors,omitempty"`
}

func createMovieHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if helpers2.GetRoleFromContext(r.Context()) != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var movieReq MovieRequest
		if err := json.NewDecoder(r.Body).Decode(&movieReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			helpers2.ErrorLogger.Println("Error decoding request body on creating movie:", err)
			return
		}
		if !validation.Name(movieReq.Name) || !validation.Description(movieReq.Description) || !validation.Rating(movieReq.Rating) {
			http.Error(w, "Bad request body", http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error starting transaction:", err)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				helpers2.ErrorLogger.Println("Transaction rolled back:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tx.Commit()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error committing transaction:", err)
				return
			}
		}()

		var movieID int
		insertQuery := "INSERT INTO movies (name, description, release_date, rating) VALUES ($1,$2, $3, $4) RETURNING movie_id"
		err = tx.QueryRow(insertQuery, movieReq.Name, movieReq.Description, movieReq.ReleaseDate, movieReq.Rating).Scan(&movieID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error inserting movie into database:", err)
			return
		}

		for _, actorName := range movieReq.Actors {
			var actorID int
			err := db.QueryRow("SELECT actor_id FROM actors WHERE name = $1", actorName).Scan(&actorID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error getting actor ID from database:", err)
				return
			}

			_, err = tx.Exec("INSERT INTO movies_actors (movie_id, actor_id) VALUES ($1,$2)", movieID, actorID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error inserting movie actor into database:", err)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)

		log.Println("Received request to create movie")
	}
}

func updateMovieHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if helpers2.GetRoleFromContext(r.Context()) != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var movieReq MovieRequest
		if err := json.NewDecoder(r.Body).Decode(&movieReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			helpers2.ErrorLogger.Println("Error decoding request body:", err)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error starting transaction:", err)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				helpers2.ErrorLogger.Println("Transaction rolled back:", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tx.Commit()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error committing transaction:", err)
				return
			}
		}()

		var queryArgs []interface{}
		updateQuery := "UPDATE movies SET"
		pIndex := 1
		if movieReq.Name != "" {
			if !validation.Name(movieReq.Name) {
				http.Error(w, "Invalid movie name", http.StatusBadRequest)
				return
			}
			updateQuery += " name=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, movieReq.Name)
			pIndex++
		}
		if movieReq.Description != "" {
			if !validation.Description(movieReq.Description) {
				http.Error(w, "Invalid movie description", http.StatusBadRequest)
				return
			}
			updateQuery += " description=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, movieReq.Description)
			pIndex++
		}
		if movieReq.ReleaseDate != "" {
			updateQuery += " release_date=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, movieReq.ReleaseDate)
			pIndex++
		}
		if movieReq.Rating != "" {
			if !validation.Rating(movieReq.Rating) {
				http.Error(w, "Invalid movie rating", http.StatusBadRequest)
				return
			}
			updateQuery += " rating=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, movieReq.Rating)
			pIndex++
		}

		updateQuery = strings.TrimSuffix(updateQuery, ",") + " WHERE movie_id=$" + strconv.Itoa(pIndex)
		queryArgs = append(queryArgs, r.URL.Query().Get("id"))

		_, err = tx.Exec(updateQuery, queryArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error updating movie in database:", err)
			return
		}

		if len(movieReq.Actors) != 0 {
			_, err := tx.Exec("DELETE FROM movies_actors WHERE movie_id = $1", r.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error deleting movie actors from database:", err)
				return
			}

			for _, actorName := range movieReq.Actors {
				var actorID int
				err := db.QueryRow("SELECT actor_id FROM actors WHERE name=$1", actorName).Scan(&actorID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Println("Error getting actor ID from database:", err)
					return
				}

				_, err = tx.Exec("INSERT INTO movies_actors (movie_id, actor_id) VALUES ($1, $2)", r.URL.Query().Get("id"), actorID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					log.Println("Error inserting movie actor into database:", err)
					return
				}
			}
		}
		w.WriteHeader(http.StatusOK)

		log.Println("Received request to update movie")
	}
}

func deleteMovieHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if helpers2.GetRoleFromContext(r.Context()) != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		_, err := db.Exec("DELETE FROM movies_actors WHERE movie_id=$1", r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error deleting movie actors from database:", err)
			return
		}
		_, err = db.Exec("DELETE FROM movies WHERE movie_id=$1", r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error deleting movie from database:", err)
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Received request to delete movie")
	}
}

func getMoviesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sortBy := r.URL.Query().Get("sort")
		var orderBy string
		switch sortBy {
		case "title":
			orderBy = "name"
		case "release_date":
			orderBy = "release_date"
		default:
			orderBy = "rating"
		}
		rows, err := db.Query(`SELECT m.movie_id, m.name, m.description, 
       			m.release_date, m.rating, array_to_json(array_agg(a.name))
			FROM movies m 
			JOIN movies_actors ma ON m.movie_id = ma.movie_id
    		JOIN actors a ON ma.actor_id = a.actor_id
			GROUP BY m.movie_id
			ORDER BY ` + orderBy + ` DESC`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error getting movies from database:", err)
			return
		}
		defer rows.Close()

		var movies []MovieResponse
		for rows.Next() {
			var movie MovieResponse
			var actorsJSON []byte
			if err := rows.Scan(&movie.ID, &movie.Name, &movie.Description, &movie.ReleaseDate, &movie.Rating, &actorsJSON); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error scanning movies from database:", err)
				return
			}
			var actors []string
			if err := json.Unmarshal(actorsJSON, &actors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error unmarshalling actors JSON:", err)
				return
			}
			movie.Actors = actors
			movies = append(movies, movie)
		}
		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(movies)

		log.Println("Received request to get movies")
	}
}

func searchMoviesHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("query")
		rows, err := db.Query(
			`SELECT m.movie_id, m.name, m.description, m.release_date, m.rating, 
       				(
						SELECT array_to_json(array_agg(a.name))
						FROM actors a
						JOIN movies_actors ma ON a.actor_id = ma.actor_id
						WHERE ma.movie_id=m.movie_id
       				) AS actors
					FROM movies m
					WHERE m.name ILIKE '%' || $1 || '%' OR
					    EXISTS(
					        SELECT 1
					        FROM actors a
					        JOIN movies_actors ma ON a.actor_id = ma.actor_id
					        WHERE ma.movie_id = m.movie_id AND a.name ILIKE '%' || $1 || '%'
					    )
				`, query)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error searching movies in database:", err)
			return
		}
		defer rows.Close()

		var movies []MovieResponse
		for rows.Next() {
			var movie MovieResponse
			var actorsJSON []byte
			if err := rows.Scan(&movie.ID, &movie.Name, &movie.Description, &movie.ReleaseDate, &movie.Rating, &actorsJSON); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error scanning searched movies from database:", err)
				return
			}
			var actors []string
			if err := json.Unmarshal(actorsJSON, &actors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error unmarshalling actors JSON in searched movies:", err)
				return
			}
			movie.Actors = actors
			movies = append(movies, movie)
		}
		if len(movies) == 0 {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(movies)

		log.Printf("Received request to search movies with query: %s\n", query)
	}
}
