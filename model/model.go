package model

import (
	"database/sql/driver"
	"fmt"
	"time"
)

// Current and Updated time
type MySQLTime struct {
	time.Time
}

func (t MySQLTime) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("\"%s\"", t.Time.Format("2006-01-02 15:04:05"))
	return []byte(stamp), nil
}

func (t *MySQLTime) UnmarshalJSON(data []byte) error {
	parseTime, err := time.Parse("\"2006-01-02 15:04:05\"", string(data))
	if err != nil {
		return err
	}

	t.Time = parseTime
	return nil
}

func (t MySQLTime) Value() (driver.Value, error) {
	return t.Time.Format("2006-01-02 15:04:05"), nil
}

func (t *MySQLTime) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		parseTime, err := time.Parse("2006-01-02 15:04:05", string(v))
		if err != nil {
			return err
		}
		t.Time = parseTime
		return nil
	case time.Time:
		t.Time = v
		return nil
	default:
		return fmt.Errorf("unsupported Scan: %T", value)
	}
}

func NewMySQLTime(t time.Time) MySQLTime {
	return MySQLTime{t}
}

type Role struct {
	ID        int       `json:"id"`
	RoleName  string    `json:"role_name"`
	CreatedAt MySQLTime `json:"created_at"`
	UpdatedAt MySQLTime `json:"updated_at"`
}

type User struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Password    string    `json:"password"`
	RoleId      int       `json:"role_id"`
	AccessToken string    `json:"access_token"`
	Active      bool      `json:"active"`
	CreatedAt   MySQLTime `json:"created_at"`
	UpdatedAt   MySQLTime `json:"updated_at"`
}

type Game struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Developer   string    `json:"developer"`
	ReleaseDate string    `json:"release_date"`
	Description string    `json:"description"`
	CreatedAt   MySQLTime `json:"created_at"`
	UpdatedAt   MySQLTime `json:"updated_at"`
}

type Review struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	GameID      int       `json:"game_id"`
	Rating      int       `json:"rating"`
	Description string    `json:"description"`
	CreatedAt   MySQLTime `json:"created_at"`
	UpdatedAt   MySQLTime `json:"updated_at"`
}

type Wishlist struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	GameID    int       `json:"game_id"`
	CreatedAt MySQLTime `json:"created_at"`
	UpdatedAt MySQLTime `json:"updated_at"`
}
