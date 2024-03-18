package api

type ActorResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Sex         string   `json:"sex"`
	DateOfBirth string   `json:"date_of_birth"`
	Movies      []string `json:"movies"`
}

type MovieResponse struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ReleaseDate string   `json:"release_date"`
	Rating      float64  `json:"rating"`
	Actors      []string `json:"actors"`
}
