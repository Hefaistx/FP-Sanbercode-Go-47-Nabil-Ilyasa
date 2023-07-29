package helper

import (
	"database/sql"
	"errors"
	d "final-project/db"
	m "final-project/model"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
)

//Email validation

func IsValidEmail(email string) bool {
	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	match, err := regexp.MatchString(emailRegex, email)
	if err != nil {
		fmt.Println("Error:", err)
		return false
	}

	return match
}

// JWT
func CreateToken(user m.User) (string, error) {
	claims := jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.RoleId,
		"exp":   time.Now().Add(time.Hour * 24).Unix(), // Token kadaluwarsa setelah 24 jam
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte("secret-key"))
}

func ExtractToken(r *http.Request) (string, error) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return "", errors.New("Authorization token not provided")
	}

	// The token is usually sent in the format "Bearer <token>"
	// We will remove the "Bearer " prefix to get the actual token
	tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

	return tokenString, nil
}

func Authenticate(r *http.Request) (jwt.MapClaims, error) {
	tokenString, err := ExtractToken(r)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte("secret-key"), nil // Replace "secret-key" with your actual secret key
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}

func AuthenticateRole(roleId int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Authorization token not provided", http.StatusUnauthorized)
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte("secret-key"), nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		roleIdFromToken, ok := claims["role"].(int64)
		if !ok || int(roleIdFromToken) != roleId {
			http.Error(w, "Access denied", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// Get email
func GetUserByEmail(email string) (m.User, error) {
	query := "SELECT id, email, name, password, role_id, access_token, active, created_at, updated_at FROM users WHERE email = ?"
	var user m.User
	err := d.Db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Name, &user.Password, &user.RoleId, &user.AccessToken, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return m.User{}, fmt.Errorf("Unauthorized access")
		}

		return m.User{}, err
	}

	return user, nil
}

func IsUserRole(userID int, roleID int) bool {
	query := "SELECT role_id FROM users WHERE id = ?"
	var userRoleID int
	err := d.Db.QueryRow(query, userID).Scan(&userRoleID)
	if err != nil {
		if err == sql.ErrNoRows {
			// User not found, return false
			return false
		}
		fmt.Println("Error:", err)
		return false
	}

	// Check if the user's role matches the specified roleID
	return userRoleID == roleID
}

func GetUserIDFromToken(r *http.Request) (int, bool) {
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		return 0, false
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte("secret-key"), nil
	})

	if err != nil || !token.Valid {
		return 0, false
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, false
	}

	userIDFloat, ok := claims["id"].(float64)
	if !ok {
		return 0, false
	}

	userID := int(userIDFloat)
	return userID, true
}
