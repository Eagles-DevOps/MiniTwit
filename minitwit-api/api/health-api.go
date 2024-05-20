package api

import (
	"net/http"

	"golang.org/x/crypto/bcrypt"
)

func Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func Stress(w http.ResponseWriter, r *http.Request) {
	password := []byte("stresstest")
	bcrypt.GenerateFromPassword(password, bcrypt.MaxCost)
	w.WriteHeader(http.StatusNoContent)
}
