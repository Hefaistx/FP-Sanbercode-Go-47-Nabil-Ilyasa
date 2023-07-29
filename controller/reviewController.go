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

// @Summary Add new review
// @Description Add a new review for a game by the authenticated user
// @Param review body Review true "Review object that needs to be added"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Review added"
// @Router /review [post]
func AddReview(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var review m.Review
	if err := json.NewDecoder(r.Body).Decode(&review); err != nil {
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

	if review.Description == "" {
		http.Error(w, "Write something, please", http.StatusBadRequest)
		return
	}
	if review.Rating > 10 {
		http.Error(w, "Rating should be 0-10", http.StatusBadRequest)
		return
	}

	createdAt := m.NewMySQLTime(time.Now())
	review.UserID = userID
	result, err := d.Db.Exec("INSERT INTO reviews (user_id, game_id, description, rating, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		review.UserID, review.GameID, review.Description, review.Rating, createdAt, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reviewID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	review.ID = int(reviewID)
	review.CreatedAt = createdAt
	review.UpdatedAt = createdAt

	response := map[string]interface{}{
		"message": "Review added",
		"review":  review,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//
