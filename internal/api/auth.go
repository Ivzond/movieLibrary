package api

import (
	"context"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strings"
)

func BasicAuthMiddleware(db *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract username and password from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		auth := strings.SplitN(authHeader, " ", 2)
		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		payload, err := base64.StdEncoding.DecodeString(auth[1])
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		pair := strings.SplitN(string(payload), ":", 2)
		if len(pair) != 2 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		username := pair[0]
		password := pair[1]

		// Authenticate user and get their role
		role, err := authenticateUser(db, username, password)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Set role in request context
		ctx := r.Context()
		ctx = context.WithValue(ctx, "role", role)
		r = r.WithContext(ctx)

		// Call the next handler
		next.ServeHTTP(w, r)
	}
}

// Authenticate user against the database and return their role
func authenticateUser(db *sql.DB, username, password string) (string, error) {
	// Query the database to get the user's role
	var role string
	err := db.QueryRow("SELECT role FROM users WHERE username = $1 AND password = $2", username, password).Scan(&role)
	if err != nil {
		return "", err
	}
	return role, nil
}
