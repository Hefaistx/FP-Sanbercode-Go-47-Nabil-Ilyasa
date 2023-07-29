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

// @Summary Add new wishlist
// @Description Add a new wishlist item for the authenticated user
// @Param wishlist body Wishlist true "Wishlist object that needs to be added"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Wishlist item added"
// @Router /wishlist [post]
func AddWish(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var wishlist m.Wishlist
	if err := json.NewDecoder(r.Body).Decode(&wishlist); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Authenticate and extract UserID from JWT
	claims, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}

	userID := int(userIDFloat)

	if wishlist.GameID <= 0 {
		http.Error(w, "Invalid game ID", http.StatusBadRequest)
		return
	}

	createdAt := m.NewMySQLTime(time.Now())
	wishlist.UserID = userID
	result, err := d.Db.Exec("INSERT INTO wishlists (user_id, game_id, created_at, updated_at) VALUES (?, ?, ?, ?)",
		wishlist.UserID, wishlist.GameID, createdAt, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	wishID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	wishlist.ID = int(wishID)
	wishlist.CreatedAt = createdAt
	wishlist.UpdatedAt = createdAt

	response := map[string]interface{}{
		"message": "wish added",
		"wish":    wishlist,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//
