package controller

import (
	"database/sql"
	"encoding/json"
	d "final-project/db"
	h "final-project/helper"
	m "final-project/model"
	"net/http"
	"strconv"
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
func GetGames(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	rows, err := d.Db.Query("SELECT title from games")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var games []m.GameResponse
	for rows.Next() {
		var game m.GameResponse
		if err := rows.Scan(&game.Title); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		games = append(games, game)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(games)
}
func GetGameDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	gameID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	var existingGame m.Game
	err = d.Db.QueryRow("SELECT id, title, developer, release_date, description, created_at, updated_at from games WHERE id = ?", gameID).
		Scan(&existingGame.ID, &existingGame.Title, &existingGame.Developer, &existingGame.ReleaseDate, &existingGame.Description, &existingGame.CreatedAt, &existingGame.UpdatedAt)
	if err != sql.ErrNoRows {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content_Type", "application/json")
	json.NewEncoder(w).Encode(existingGame)
}
func DeleteGame(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	gameID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid game id", http.StatusBadRequest)
		return
	}
	var existingGame m.Game
	err = d.Db.QueryRow("SELECT id, title, developer, release_date, description, created_at, updated_at from games WHERE id = ?", gameID).
		Scan(&existingGame.ID, &existingGame.Title, &existingGame.Developer, &existingGame.ReleaseDate, &existingGame.Description, &existingGame.CreatedAt, &existingGame.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = d.Db.Exec("DELETE FROM games WHERE id = ?", gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"message": "Game successfully deleted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func UpdateGame(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Authenticate and extract UserID from JWT
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
	gameID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid game id", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var existingGame m.Game
	var game m.Game
	if err := json.NewDecoder(r.Body).Decode(&game); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = d.Db.QueryRow("SELECT id, title, developer, release_date, description, updated_at from games WHERE id = ?", gameID).
		Scan(&existingGame.ID, &existingGame.Title, &existingGame.Developer, &existingGame.ReleaseDate, &existingGame.Description, &existingGame.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Game not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdAt := m.NewMySQLTime(time.Now())

	_, err = d.Db.Exec("UPDATE games SET title = ?, developer = ?, release_date = ?, description = ?, updated_at = ? WHERE id = ?",
		game.Title, game.Developer, game.ReleaseDate, game.Description, createdAt, gameID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	game.Title = existingGame.Title
	game.Developer = existingGame.Developer
	game.ReleaseDate = existingGame.ReleaseDate
	game.Description = existingGame.Description
	game.UpdatedAt = existingGame.UpdatedAt

	response := map[string]interface{}{
		"message": "Game data updated",
		"review":  existingGame,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
