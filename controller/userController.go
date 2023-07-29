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
	"golang.org/x/crypto/bcrypt"
)

var user m.User

// @Summary Register new user
// @Description Register a new user with the provided information
// @Param user body User true "User object that needs to be registered"
// @Success 200 {object} map[string]interface{} "Registration successful!"
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if user.Name == "" || user.Email == "" || user.Password == "" {
		http.Error(w, "Fill all the blank!", http.StatusBadRequest)
		return
	}

	if !h.IsValidEmail(user.Email) {
		http.Error(w, "Invalid Email", http.StatusBadRequest)
		return
	}

	if len(user.Password) < 8 {
		http.Error(w, "Password should be at least 8 characters", http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	user.Password = string(hashedPassword)
	user.RoleId = 1
	user.AccessToken = ""
	user.Active = false
	createdAt := m.NewMySQLTime(time.Now())
	result, err := d.Db.Exec("INSERT INTO users (email, name, password, role_id, access_token,active,  created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.Email, user.Name, user.Password, user.RoleId, user.AccessToken, user.Active, createdAt, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	userID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user.ID = int(userID)
	user.CreatedAt = createdAt
	user.UpdatedAt = createdAt

	response := map[string]interface{}{
		"message": "Registration successful!",
		"user":    user,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Login user
// @Description Log in user with the provided credentials
// @Param user body User true "User object that contains login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Router /login [post]
// Fungsi untuk handle login
func Login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	registeredUser, err := h.GetUserByEmail(user.Email)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(registeredUser.Password), []byte(user.Password))
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	token, err := h.CreateToken(registeredUser)
	if err != nil {
		http.Error(w, "Failed to create JWT token", http.StatusInternalServerError)
		return
	}
	registeredUser.AccessToken = token
	registeredUser.Active = true
	updateTokenQuery := "UPDATE users SET access_token = ?, active = ? WHERE id = ?"
	_, err = d.Db.Exec(updateTokenQuery, token, registeredUser.Active, registeredUser.ID)
	if err != nil {
		http.Error(w, "Failed to update access token", http.StatusInternalServerError)
		return
	}
	response := map[string]interface{}{
		"message": "Login successful",
		"user":    registeredUser,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Delete user
// @Description Delete a user by its ID (only accessible by admin)
// @Param id path int true "User ID to be deleted"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "User successfully deleted"
// @Router /users/{id} [delete]
func DeleteUser(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	userID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}
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
	var existingUser m.User
	err = d.Db.QueryRow("SELECT id, email, name, password, role_id, access_token, active, created_at, updated_at FROM users WHERE id = ?", userID).
		Scan(&existingUser.ID, &existingUser.Email, &existingUser.Name, &existingUser.Password, &existingUser.RoleId, &existingUser.AccessToken, &existingUser.Active, &existingUser.CreatedAt, &existingUser.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = d.Db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"message": "User successfully deleted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

//
