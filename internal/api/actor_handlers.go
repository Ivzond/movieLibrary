package api

import (
	"database/sql"
	"encoding/json"
	"log"
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
		var actorReq ActorRequest
		if err := json.NewDecoder(r.Body).Decode(&actorReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err := db.Exec("INSERT INTO actors (name, sex, date_of_birth) VALUES ($1, $2, $3)",
			actorReq.Name, actorReq.Sex, actorReq.DateOfBirth)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)

		log.Println("Received request to create actor")
	}
}

func updateActorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var actorReq ActorRequest
		if err := json.NewDecoder(r.Body).Decode(&actorReq); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
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
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Received request to update actor")
	}
}

func deleteActorHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := db.Exec("DELETE FROM actors WHERE actor_id=$1", r.URL.Query().Get("id"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)

		log.Println("Received request to delete actor")
	}
}

func getActorsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT actor_id, name, sex, date_of_birth FROM actors")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var actors []ActorResponse
		for rows.Next() {
			var actor ActorResponse
			if err := rows.Scan(&actor.ID, &actor.Name, &actor.Sex, &actor.DateOfBirth); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			actors = append(actors, actor)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(actors)

		log.Println("Received request to get actors")
	}
}
