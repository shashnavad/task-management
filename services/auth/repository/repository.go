package repository

import (
	"database/sql"

	"github.com/task-management-system/services/auth/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(email, username, password, fullName string) (int, error) {
	// In a real scenario, you'd use a prepared statement and insert user data
	// For this example, we'll use a simplified approach
	// IMPORTANT: Hash the password before storing it!
	var userID int
	stmt := `INSERT INTO users (email, username, password_hash, full_name) VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(stmt, email, username, password, fullName).Scan(&userID)
	if err != nil {
		return 0, err
	}
	return userID, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*models.User, error) {
	// For demo, just return a mock user
	return &models.User{
		Username:     username,
		PasswordHash: "mock-hash", // In real code, retrieve from DB
	}, nil
}
