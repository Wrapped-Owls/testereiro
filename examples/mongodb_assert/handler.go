package mongodb_assert

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/wrapped-owls/testereiro/examples/mongodb_assert/dungeonstore"
)

// NewHandler creates an HTTP handler backed by a MongoDB database.
func NewHandler(db *mongo.Database) http.Handler {
	store := dungeonstore.NewDungeonStore(db)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /dungeonformers", func(w http.ResponseWriter, r *http.Request) {
		results, err := store.FindAll(r.Context())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err = writeJSON(w, http.StatusOK, results); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	mux.HandleFunc(
		"GET /dungeonformers/class/{class}",
		func(w http.ResponseWriter, r *http.Request) {
			class := r.PathValue("class")

			results, err := store.FindByClass(r.Context(), class)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if len(results) == 0 {
				http.Error(w, "no dungeonformers found for class", http.StatusNotFound)
				return
			}

			if err = writeJSON(w, http.StatusOK, results); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		},
	)

	return mux
}

func writeJSON(w http.ResponseWriter, statusCode int, payload any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		return fmt.Errorf("encode json response: %w", err)
	}
	return nil
}
