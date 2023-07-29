package controller

import (
	"database/sql"
	"encoding/json"
	d "final-project/db"
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

// @Summary Delete role by ID
// @Description Delete a role by its ID
// @Param id path int true "Role ID to delete"
// @Security ApiKeyAuth
// @Success 200 {object} map[string]string "Role successfully deleted"
// @Router /role/{id} [delete]
func DeleteRole(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

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
