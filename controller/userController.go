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
// @Param user body m.User true "User object that needs to be registered"
// @Success 200 {object} map[string]interface{} "Registration successful!"
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 409 {object} map[string]string "Email already registered" (when the provided email is already registered)
// @Failure 400 {object} map[string]string "Fill all the blank!" (when name, email, or password is empty)
// @Failure 400 {object} map[string]string "Invalid Email" (when the provided email is not a valid email address)
// @Failure 400 {object} map[string]string "Password should be at least 8 characters" (when the provided password is less than 8 characters)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database or password hashing)
// @Router /register [post]
func Register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var count int
	err := d.Db.QueryRow("SELECT COUNT(*) FROM users WHERE email = ?", user.Email).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Email already registered", http.StatusConflict)
		return
	}
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
// @Param user body m.User true "User object that contains login credentials"
// @Success 200 {object} map[string]interface{} "Login successful"
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided credentials are incorrect)
// @Failure 500 {object} map[string]string "Failed to create JWT token" (when there is an error creating the JWT token)
// @Failure 500 {object} map[string]string "Failed to update access token" (when there is an error updating the access token in the database)
// @Router /login [post]
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

// @Summary Get users
// @Description Get a list of users with limited information based on the user's role
// @Security ApiKeyAuth
// @Success 200 {object} []m.UserResponse "List of users"
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /users [get]
func GetUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	claims, err := h.Authenticate(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	_, ok := claims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	rows, err := d.Db.Query("SELECT name, role_id FROM users")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []m.UserResponse
	for rows.Next() {
		var user m.UserResponse
		if err := rows.Scan(&user.Name, &user.RoleId); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

// @Summary Get user details
// @Description Get detailed information about a specific user (only accessible by admin)
// @Param id path int true "User ID to be retrieved"
// @Security ApiKeyAuth
// @Success 200 {object} m.User "User details"
// @Failure 400 {object} map[string]string "Invalid userID" (when the provided user ID in the URL is not a valid integer)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 403 {object} map[string]string "Access denied" (when the user does not have admin role)
// @Failure 404 {object} map[string]string "User not found" (when the requested user ID does not exist in the database)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /users/{id} [get]
func GetUserDetail(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

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

	//Check if the role claim exists and is equal to 2 (admin role)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existingUser)
}

// @Summary Delete user
// @Description Delete a user by its ID (only accessible by admin)
// @Param id path int true "User ID to be deleted"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "User successfully deleted"
// @Failure 400 {object} map[string]string "Invalid userID" (when the provided user ID in the URL is not a valid integer)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 403 {object} map[string]string "Access denied" (when the user does not have admin role)
// @Failure 404 {object} map[string]string "User not found" (when the requested user ID does not exist in the database)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
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

// @Summary Update user
// @Description Update user data (name and password)
// @Security ApiKeyAuth
// @Param user body m.User true "User object that contains updated user data"
// @Success 200 {object} map[string]interface{} "User data updated"
// @Failure 400 {object} map[string]string "Invalid request body" (when the request body does not contain valid JSON or is missing required fields)
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /users [put]
func UpdateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user m.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
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
	var existingUser m.User

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		return
	}
	userID := int(userIDFloat)

	err = d.Db.QueryRow("SELECT id, email, name, password, role_id, access_token, active, created_at, updated_at FROM users WHERE id = ?", userID).
		Scan(&existingUser.ID, &existingUser.Email, &existingUser.Name, &existingUser.Password, &existingUser.RoleId, &existingUser.AccessToken, &existingUser.Active, &existingUser.CreatedAt, &existingUser.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	createdAt := m.NewMySQLTime(time.Now())
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)

	_, err = d.Db.Exec("UPDATE users SET name = ?, password = ?, updated_at = ? WHERE id = ?",
		user.Name, user.Password, createdAt, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existingUser.Name = user.Name
	existingUser.Password = user.Password
	existingUser.UpdatedAt = createdAt

	response := map[string]interface{}{
		"message": "User data updated",
		"review":  existingUser,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// @Summary Logout user
// @Description Log out user by clearing the access token
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Logout successful"
// @Failure 401 {object} map[string]string "Unauthorized" (when the provided JWT token is invalid or missing)
// @Failure 500 {object} map[string]string "Internal server error" (when there is a problem with the database)
// @Router /logout [post]
func Logout(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	// Clear the access token and set active to false for the user in the database
	_, err = d.Db.Exec("UPDATE users SET access_token = '', active = false WHERE id = ?", userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]string{
		"message": "Logout successful",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
