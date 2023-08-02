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

// @Summary Create new role
// @Description Create a new role with the specified role name
// @Param role body Role true "Role object that needs to be created"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]interface{} "Role created"
// @Router /role [post]
func CreateRole(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var role m.Role
	if err := json.NewDecoder(r.Body).Decode(&role); err != nil {
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
	var count int
	err = d.Db.QueryRow("SELECT COUNT(*) FROM roles WHERE role_name = ?", role.RoleName).Scan(&count)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if count > 0 {
		http.Error(w, "Role already exist", http.StatusConflict)
		return
	}
	if role.RoleName == "" {
		http.Error(w, "Role Name should be filled!", http.StatusBadRequest)
		return
	}
	createdAt := m.NewMySQLTime(time.Now())
	result, err := d.Db.Exec("INSERT INTO roles (role_name, created_at, updated_at) VALUES (?, ?, ?)",
		role.RoleName, createdAt, createdAt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	roleID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	role.ID = int(roleID)
	role.CreatedAt = createdAt
	role.UpdatedAt = createdAt
	response := map[string]interface{}{
		"message": "Role creared",
		"role":    role,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetRole(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	// Implement logic to filter data based on user ID or any other criteria, if needed.
	// For example:
	// userID := int(claims["id"].(float64))
	// rows, err := d.Db.Query("SELECT name, role_id FROM users WHERE id = ?", userID)
	rows, err := d.Db.Query("SELECT id, role_name FROM roles")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var roles []m.Role
	for rows.Next() {
		var role m.Role
		if err := rows.Scan(&role.ID, &role.RoleName); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		roles = append(roles, role)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roles)
}

// @Summary Delete role by ID
// @Description Delete a role by its ID
// @Param id path int true "Role ID to delete"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Role successfully deleted"
// @Router /role/{id} [delete]
func DeleteRole(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

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
	roleID, err := strconv.Atoi(ps.ByName("id"))
	if err != nil {
		http.Error(w, "Invalid userID", http.StatusBadRequest)
		return
	}

	var existingRole m.Role
	err = d.Db.QueryRow("SELECT id, role_name, created_at, updated_at FROM roles WHERE id = ?", roleID).
		Scan(&existingRole.ID, &existingRole.RoleName, &existingRole.CreatedAt, &existingRole.UpdatedAt)
	if err == sql.ErrNoRows {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = d.Db.Exec("DELETE FROM roles WHERE id = ?", roleID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	response := map[string]string{
		"message": "Role successfully deleted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
