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

// AddReview handles the HTTP request to add a new review for a game by the authenticated user.
// @Summary Add new review
// @Description Add a new review for a game by the authenticated user
// @Param review body m.Review true "Review object that needs to be added"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Review added"
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 401 {object} map[string]string "Unauthorized" (when the JWT token is missing or invalid)
// @Failure 400 {object} map[string]string "Nothing found with given id" (when the specified game ID in the review does not exist)
// @Failure 409 {object} map[string]string "You can only make one review per game" (when the user has already reviewed the specified game)
// @Failure 400 {object} map[string]string "Write something, please" (when the review description is empty)
// @Failure 400 {object} map[string]string "Rating should be 0-10" (when the review rating is greater than 10)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
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
	gameExists, err := h.IsGameExists(review.GameID)
	if err != nil {
		http.Error(w, "Failed to check game existence", http.StatusInternalServerError)
		return
	}

	if !gameExists {
		http.Error(w, "Nothing found with given id", http.StatusBadRequest)
		return
	}
	var count int
	err = d.Db.QueryRow("SELECT COUNT(*) FROM reviews WHERE user_id = ? AND game_id = ?", userID, review.GameID).Scan(&count)
	if err != nil {
		http.Error(w, "Not found", http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "You can only make one review per game", http.StatusConflict)
		return
	}

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

// UpdateReview handles the HTTP request to update an existing review for a game by the authenticated user.
// @Summary Update review
// @Description Update an existing review for a game by the authenticated user
// @Param review body m.Review true "Review object that needs to be updated"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Review updated"
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 401 {object} map[string]string "Unauthorized" (when the JWT token is missing or invalid)
// @Failure 404 {object} map[string]string "Review not found" (when the specified review ID does not exist in the database)
// @Failure 400 {object} map[string]string "Write something, please" (when the updated review description is empty)
// @Failure 400 {object} map[string]string "Rating should be 0-10" (when the updated review rating is greater than 10)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /review [post]
func UpdateReview(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	var existingReview m.Review
	err = d.Db.QueryRow("SELECT id, user_id, game_id, created_at, updated_at FROM reviews WHERE user_id = ?", review.ID).
		Scan(&existingReview.ID, &existingReview.UserID, &existingReview.GameID, &existingReview.CreatedAt, &existingReview.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
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

	_, err = d.Db.Exec("UPDATE reviews SET rating = ?, description = ?, updated_at = ? WHERE id = ?",
		review.Rating, review.Description, createdAt, review.UserID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update review fields
	review.UpdatedAt = createdAt

	response := map[string]interface{}{
		"message": "Review updated",
		"review":  review,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetReview handles the HTTP request to fetch all reviews from the database.
// Only authenticated users can access this endpoint.
// @Summary Get all reviews
// @Description Get a list of all reviews
// @Security ApiKeyAuth
// @Success 200 {object} []m.Review "List of reviews"
// @Failure 401 {object} map[string]string "Unauthorized" (when the JWT token is missing or invalid)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /review [get]
func GetReview(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	claims, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	_, ok := claims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	rows, err := d.Db.Query("SELECT id, user_id, game_id, rating, description, created_at, updated_at from reviews")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var reviews []m.Review
	for rows.Next() {
		var review m.Review
		if err := rows.Scan(&review.ID, &review.UserID, &review.GameID, &review.Rating, &review.Description, &review.CreatedAt, &review.UpdatedAt); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		reviews = append(reviews, review)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reviews)
}

// DeleteReview handles the HTTP request to delete a review by its ID.
// Only authenticated users can delete their own reviews.
// @Summary Delete review by ID
// @Description Delete a review by its ID
// @Param id path int true "Review ID to delete"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Review successfully deleted"
// @Failure 400 {object} map[string]string "Invalid review ID" (when the review ID in the URL path is not a valid integer)
// @Failure 401 {object} map[string]string "Unauthorized" (when the JWT token is missing or invalid)
// @Failure 404 {object} map[string]string "Review not found" (when the specified review ID does not exist in the database)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /review/{id} [delete]
func DeleteReview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userID, err := strconv.Atoi(ps.ByName("uid"))
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}
	_, err = h.Authenticate(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var existingReview m.Review
	err = d.Db.QueryRow("SELECT id, user_id, game_id, rating, description, created_at, updated_at from reviews WHERE user_id = ?", userID).
		Scan(&existingReview.ID, &existingReview.UserID, &existingReview.GameID, &existingReview.Rating, &existingReview.Description, &existingReview.CreatedAt, &existingReview.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = d.Db.Exec("DELETE from reviews where id = ?", existingReview.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"message": "Review has been deleted",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
