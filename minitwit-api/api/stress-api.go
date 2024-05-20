package api

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func Stress(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		lg.Infof("Stress handler execution time: %s", elapsed)
	}()
	lg.Info("Stress handler invoked")
	password := []byte("stresstest")
	bcrypt.GenerateFromPassword(password, bcrypt.MaxCost)
	w.WriteHeader(http.StatusNoContent)
}
