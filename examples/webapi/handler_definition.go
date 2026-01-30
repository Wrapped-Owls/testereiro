package webapi

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type IndieGame struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Developer   string `json:"developer"`
	ReleaseYear int    `json:"release_year"`
}

var Games = []IndieGame{
	{ID: 1, Title: "Hollow Knight", Developer: "Team Cherry", ReleaseYear: 2017},
	{ID: 2, Title: "Outer Wilds", Developer: "Mobius Digital", ReleaseYear: 2019},
	{ID: 3, Title: "Celeste", Developer: "Extremely OK Games", ReleaseYear: 2018},
	{ID: 4, Title: "Hades", Developer: "Supergiant Games", ReleaseYear: 2020},
}

func NewHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /games", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Games)
	})

	mux.HandleFunc("GET /games/{id}", func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		for _, game := range Games {
			if game.ID == id {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(game)
				return
			}
		}

		http.Error(w, "Game not found", http.StatusNotFound)
	})

	return mux
}
