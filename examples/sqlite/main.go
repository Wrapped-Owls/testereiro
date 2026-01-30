package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	_ "modernc.org/sqlite"
)

type IndieGame struct {
	ID          int    `json:"id"           db:"id"`
	Title       string `json:"title"        db:"title"`
	Developer   string `json:"developer"    db:"developer"`
	ReleaseYear int    `json:"release_year" db:"release_year"`
}

type GameStore struct {
	DB *sql.DB
}

func NewHandler(db *sql.DB) http.Handler {
	store := &GameStore{DB: db}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /games", store.handleGetAllGames)
	mux.HandleFunc("GET /games/{id}", store.handleGetGameByID)

	return mux
}

func (s *GameStore) handleGetAllGames(w http.ResponseWriter, r *http.Request) {
	rows, err := s.DB.Query("SELECT id, title, developer, release_year FROM games")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var games []IndieGame
	for rows.Next() {
		var g IndieGame
		if err = rows.Scan(&g.ID, &g.Title, &g.Developer, &g.ReleaseYear); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		games = append(games, g)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(games)
}

func (s *GameStore) handleGetGameByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var g IndieGame
	err = s.DB.QueryRow("SELECT id, title, developer, release_year FROM games WHERE id = ?", id).
		Scan(&g.ID, &g.Title, &g.Developer, &g.ReleaseYear)
	if err != nil {
		statusErr := http.StatusInternalServerError
		if errors.Is(err, sql.ErrNoRows) {
			statusErr = http.StatusNotFound
			err = fmt.Errorf("could not find game with id `%v`: %w", id, err)
		}

		http.Error(w, err.Error(), statusErr)
		return

	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(g)
}
