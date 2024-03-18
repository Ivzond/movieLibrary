package validation

import "strconv"

func Name(name string) bool {
	return len(name) > 0 && len(name) <= 150
}

func Description(description string) bool {
	return len(description) <= 1000
}

func Rating(rating string) bool {
	ratingFloat, err := strconv.ParseFloat(rating, 64)
	if err != nil {
		return false
	}
	return ratingFloat >= 0 && ratingFloat <= 10
}
