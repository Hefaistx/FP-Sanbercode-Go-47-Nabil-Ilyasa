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

	var count int
	err = d.Db.QueryRow("SELECT COUNT(*) FROM wishlists WHERE game_id = ? AND user_id = ?", wishlist.GameID, userID).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "The Game is already exists in your list", http.StatusBadRequest)
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

func GetWish(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	claims, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	userID, ok := claims["id"].(float64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	rows, err := d.Db.Query("SELECT w.id, w.user_id, g.title AS game_title, w.created_at, w.updated_at FROM wishlists w JOIN games g ON w.game_id = g.id WHERE w.user_id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var wishes []m.WishlistWithGameTitle
	for rows.Next() {
		var wish m.WishlistWithGameTitle
		if err := rows.Scan(&wish.ID, &wish.UserID, &wish.GameTitle, &wish.CreatedAt, &wish.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		wishes = append(wishes, wish)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(wishes)
}
func DeleteWish(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	wishID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	claims, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	userID, ok := claims["id"].(float64)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var existingWish m.Wishlist
	err = d.Db.QueryRow("SELECT id, user_id, game_id, created_at, updated_at from wishlists where id = ? AND user_id = ?", wishID, userID).
		Scan(&existingWish.ID, &existingWish.UserID, &existingWish.GameID, &existingWish.CreatedAt, &existingWish.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Wish not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = d.Db.Exec("DELETE FROM wishlists WHERE id = ? AND user_id = ?", existingWish.ID, existingWish.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"message": "Wish has been deleted",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
