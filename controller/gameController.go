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

// AddGame handles the HTTP request to add a new game with the provided information (admin only).
// @Summary Add new game
// @Description Add a new game with the provided information (admin only)
// @Param game body m.Game true "Game object that needs to be added"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Game added"
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 403 {object} map[string]string "Access denied" (when the user does not have admin role)
// @Failure 400 {object} map[string]string "Fill all the blank!" (when title, developer, or description is empty)
// @Failure 400 {object} map[string]string "Invalid Release Date format. Use 'YYYY-MM-DD'." (when the provided Release Date has an invalid format)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
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

// GetGames retrieves a list of game titles.
// @Summary Get games
// @Description Get a list of game titles
// @Success 200 {object} []m.GameResponse "List of game titles" // Sesuaikan dengan tipe m.GameResponse
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /games [get]
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

// GetGameDetail retrieves detailed information about a specific game.
// @Summary Get game details
// @Description Get detailed information about a specific game
// @Param id path int true "Game ID to be retrieved"
// @Success 200 {object} m.Game "Game details"
// @Failure 400 {object} map[string]string "Invalid game ID" (when the provided game ID in the URL is not a valid integer)
// @Failure 404 {object} map[string]string "Game not found" (when the requested game ID does not exist in the database)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /games/{id} [get]
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

// DeleteGame handles the HTTP request to delete a game by its ID (admin only).
// @Summary Delete game
// @Description Delete a game by its ID (admin only)
// @Param id path int true "Game ID to be deleted"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Game successfully deleted"
// @Failure 400 {object} map[string]string "Invalid game ID" (when the provided game ID in the URL is not a valid integer)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 403 {object} map[string]string "Access denied" (when the user does not have admin role)
// @Failure 404 {object} map[string]string "Game not found" (when the requested game ID does not exist in the database)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /games/{id} [delete]
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

// UpdateGame handles the HTTP request to update game data (admin only).
// @Summary Update game
// @Description Update game data (title, developer, release date, and description) (admin only)
// @Security ApiKeyAuth
// @Param id path int true "Game ID to be updated"
// @Param game body m.Game true "Game object that contains updated game data" // Sesuaikan dengan tipe m.Game
// @Success 200 {object} map[string]interface{} "Game data updated"
// @Failure 400 {object} map[string]string "Invalid game ID" (when the provided game ID in the URL is not a valid integer)
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 403 {object} map[string]string "Access denied" (when the user does not have admin role)
// @Failure 400 {object} map[string]string "Fill all the blank!" (when title, developer, or description is empty)
// @Failure 400 {object} map[string]string "Invalid Release Date format. Use 'YYYY-MM-DD'." (when the provided Release Date has an invalid format)
// @Failure 404 {object} map[string]string "Game not found" (when the requested game ID does not exist in the database)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /games/{id} [post]
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
		"review":  game,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
