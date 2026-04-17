package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pdrhlik/deliberix/server/config"
	"github.com/pdrhlik/deliberix/server/notify"
	"github.com/pdrhlik/deliberix/server/store"
)

type AppHandlerFunc func(http.ResponseWriter, *http.Request) error

type Handler struct {
	Store  *store.Store
	Config config.Config
	Notify *notify.Service
}

func parseJSON(r *http.Request, dest interface{}) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dest)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, message string) error {
	return writeJSON(w, status, map[string]string{
		"error": message,
	})
}

func ErrorHandler(f AppHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	}
}
