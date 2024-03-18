package api

import "time"

type Actor struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Sex         string    `json:"sex"`
	DateOfBirth time.Time `json:"date_of_birth"`
}

type Movie struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ReleaseDate time.Time `json:"release_date"`
	Rating      float64   `json:"rating"`
	Actors      []Actor   `json:"actors"`
}
