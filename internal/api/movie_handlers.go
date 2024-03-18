package api

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
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
		var movieReq MovieRequest
		if err := json.NewDecoder(r.Body).Decode(&movieReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tx.Commit()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}()

		var movieID int
		insertQuery := "INSERT INTO movies (name, description, release_date, rating) VALUES ($1,$2, $3, $4) RETURNING movie_id"
		err = tx.QueryRow(insertQuery, movieReq.Name, movieReq.Description, movieReq.ReleaseDate, movieReq.Rating).Scan(&movieID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		for _, actorName := range movieReq.Actors {
			var actorID int
			err := db.QueryRow("SELECT actor_id FROM actors WHERE name = $1", actorName).Scan(&actorID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			_, err = tx.Exec("INSERT INTO movies_actors (movie_id, actor_id) VALUES ($1,$2)", movieID, actorID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
		w.WriteHeader(http.StatusCreated)

		log.Println("Received request to create movie")
	}
}

func updateMovieHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var movieReq MovieRequest
		if err := json.NewDecoder(r.Body).Decode(&movieReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		tx, err := db.Begin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			if err != nil {
				tx.Rollback()
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			err = tx.Commit()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}()

		var queryArgs []interface{}
		updateQuery := "UPDATE movies SET"
		if movieReq.Name != "" {
			updateQuery += " name=$1,"
			queryArgs = append(queryArgs, movieReq.Name)
		}
		if movieReq.Description != "" {
			updateQuery += " description=$2,"
			queryArgs = append(queryArgs, movieReq.Description)
		}
		if movieReq.ReleaseDate != "" {
			updateQuery += " release_date=$3,"
			queryArgs = append(queryArgs, movieReq.ReleaseDate)
		}
		if movieReq.Rating != "" {
			updateQuery += " rating=$4,"
			queryArgs = append(queryArgs, movieReq.Rating)
		}

		updateQuery = updateQuery[:len(updateQuery)-1] + " WHERE movie_id=$5"
		queryArgs = append(queryArgs, r.URL.Query().Get("id"))

		_, err = tx.Exec(updateQuery, queryArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(movieReq.Actors) != 0 {
			_, err := tx.Exec("DELETE FROM movies_actors WHERE movie_id = $1", r.URL.Query().Get("id"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			for _, actorName := range movieReq.Actors {
				var actorID int
				err := db.QueryRow("SELECT actor_id FROM actors WHERE name=$1", actorName).Scan(&actorID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				_, err = tx.Exec("INSERT INTO movies_actors (movie_id, actor_id) VALUES ($1, $2)", r.URL.Query().Get("id"), actorID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
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
		_, err := db.Exec("DELETE FROM movies_actors WHERE movie_id=$1", r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, err = db.Exec("DELETE FROM movies WHERE movie_id=$1", r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
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
			return
		}
		defer rows.Close()

		var movies []MovieResponse
		for rows.Next() {
			var movie MovieResponse
			var actorsJSON []byte
			if err := rows.Scan(&movie.ID, &movie.Name, &movie.Description, &movie.ReleaseDate, &movie.Rating, &actorsJSON); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			var actors []string
			if err := json.Unmarshal(actorsJSON, &actors); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
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

	}
}
