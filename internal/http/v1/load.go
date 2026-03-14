package v1

import (
	"net/http"
)

func Load() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	}
}
