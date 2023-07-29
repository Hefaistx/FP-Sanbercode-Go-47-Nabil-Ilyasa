package controller

import (
	"encoding/json"
	d "final-project/db"
	h "final-project/helper"
	m "final-project/model"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
)

// @Summary Add new game
// @Description Add a new game with the provided information (admin only)
// @Param game body Game true "Game object that needs to be added"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Game added"
// @Router /games [post]
func AddGame(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var game m.Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	claims, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Check if the role claim exists and is equal to 2 (admin role)
	if role, ok := claims["role"].(float64); !ok || role != 2 {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	if game.Title == "" || game.Developer == "" || game.Description == "" {
		http.Error(w, "Fill all the blank!", http.StatusBadRequest)
		return
	}
	createdAt := m.NewMySQLTime(time.Now())
	releaseDate, err := time.Parse("2006-01-02", game.ReleaseDate)
	if err != nil {
		http.Error(w, "Invalid Release Date format. Use 'YYYY-MM-DD'.", http.StatusBadRequest)
		return
	}
	result, err := d.Db.Exec("INSERT INTO games (title, developer, release_date, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		game.Title, game.Developer, releaseDate, game.Description, createdAt, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	gameID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	game.ID = int(gameID)
	game.CreatedAt = createdAt
	game.UpdatedAt = createdAt

	response := map[string]interface{}{
		"message": "Game added",
		"game":    game,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
