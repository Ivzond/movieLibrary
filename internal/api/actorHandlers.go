package api

import (
	"database/sql"
	"encoding/json"
	"log"
	helpers2 "movieLibrary/internal/pkg/helpers"
	"net/http"
	"strconv"
	"strings"
)

type ActorRequest struct {
	Name        string `json:"name,omitempty"`
	Sex         string `json:"sex,omitempty"`
	DateOfBirth string `json:"date_of_birth,omitempty"`
}

func createActorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if helpers2.GetRoleFromContext(r.Context()) != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var actorReq ActorRequest
		if err := json.NewDecoder(r.Body).Decode(&actorReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			helpers2.ErrorLogger.Println("Error decoding actor request on creating:", err)
			return
		}

		_, err := db.Exec("INSERT INTO actors (name, sex, date_of_birth) VALUES ($1, $2, $3)",
			actorReq.Name, actorReq.Sex, actorReq.DateOfBirth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error executing SQL query on creating actor:", err)
			return
		}

		w.WriteHeader(http.StatusCreated)

		log.Println("Received request to create actor")
	}
}

func updateActorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if helpers2.GetRoleFromContext(r.Context()) != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var actorReq ActorRequest
		if err := json.NewDecoder(r.Body).Decode(&actorReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			helpers2.ErrorLogger.Println("Error decoding actor request on updating:", err)
			return
		}

		var queryArgs []interface{}
		updateQuery := "UPDATE actors SET"
		pIndex := 1
		if actorReq.Name != "" {
			updateQuery += " name=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, actorReq.Name)
			pIndex++
		}
		if actorReq.Sex != "" {
			updateQuery += " sex=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, actorReq.Sex)
			pIndex++
		}
		if actorReq.DateOfBirth != "" {
			updateQuery += " date_of_birth=$" + strconv.Itoa(pIndex) + ","
			queryArgs = append(queryArgs, actorReq.DateOfBirth)
			pIndex++
		}

		updateQuery = strings.TrimSuffix(updateQuery, ",") + " WHERE actor_id=$" + strconv.Itoa(pIndex)
		queryArgs = append(queryArgs, r.URL.Query().Get("id"))

		_, err := db.Exec(updateQuery, queryArgs...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error executing SQL query on updating actor:", err)
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Received request to update actor")
	}
}

func deleteActorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if helpers2.GetRoleFromContext(r.Context()) != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		_, err := db.Exec("DELETE FROM actors WHERE actor_id=$1", r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error executing SQL query on deleting actor:", err)
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Received request to delete actor")
	}
}

func getActorsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query(`SELECT a.actor_id, a.name, a.sex, a.date_of_birth,
       			array_to_json(array_agg(m.name))
			FROM actors a 
			JOIN movies_actors ma ON a.actor_id = ma.actor_id
    		JOIN movies m ON ma.movie_id = m.movie_id
			GROUP BY a.actor_id`)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error executing SQL query on reading actors:", err)
			return
		}
		defer rows.Close()

		var actors []ActorResponse
		for rows.Next() {
			var actor ActorResponse
			var moviesJSON []byte
			if err := rows.Scan(&actor.ID, &actor.Name, &actor.Sex, &actor.DateOfBirth, &moviesJSON); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error scanning rows:", err)
				return
			}
			var movies []string
			if err := json.Unmarshal(moviesJSON, &movies); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				helpers2.ErrorLogger.Println("Error unmarshalling movies JSON:", err)
				return
			}

			actor.Movies = movies
			actors = append(actors, actor)
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(actors); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			helpers2.ErrorLogger.Println("Error encoding actors response:", err)
			return
		}

		log.Println("Received request to get actors")
	}
}
