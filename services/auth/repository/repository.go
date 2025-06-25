// repository/user_repository.go

package repository

import (
	"database/sql"
	"errors"
	"log"
	"strconv"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
)

// DBUser represents a user in the database
type DBUser struct {
	ID       string
	Username string
	PasswordHash string
	Email    string
	FullName string
}

// UserRepositoryInterface defines methods for the user repository
type UserRepositoryInterface interface {
	CreateUser(email, username, password, fullName string) (string, error)
	CheckUserExists(username, email string) (bool, error)
	GetUserByUsername(username string) (*DBUser, error)
	GetUserByID(userID string) (*DBUser, error)
	UpdateUser(userID, fullName, email string) error
}

// UserRepository implements UserRepositoryInterface
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// InitDB initializes the database connection
func InitDB() *sql.DB {
	// Replace with your actual database connection details
	db, err := sql.Open("mysql", "username:password@tcp(localhost:3306)/auth_service")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	log.Println("Connected to database")
	return db
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(email, username, password, fullName string) (string, error) {
	// Prepare statement to prevent SQL injection
	stmt, err := r.db.Prepare("INSERT INTO users (email, username, password_hash, full_name) VALUES (?, ?, ?, ?)")
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	// Execute statement
	result, err := stmt.Exec(email, username, password, fullName)
	if err != nil {
		return "", err
	}

	// Get inserted ID
	id, err := result.LastInsertId()
	if err != nil {
		return "", err
	}

	return strconv.FormatInt(id, 10), nil
}

// CheckUserExists checks if a user exists with the given username or email
func (r *UserRepository) CheckUserExists(username, email string) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(username string) (*DBUser, error) {
	user := &DBUser{}
	err := r.db.QueryRow("SELECT id, username, password_hash, email, full_name FROM users WHERE username = ?", username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(userID string) (*DBUser, error) {
	user := &DBUser{}
	err := r.db.QueryRow("SELECT id, username, password_hash, email, full_name FROM users WHERE id = ?", userID).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Email, &user.FullName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user, nil
}

// UpdateUser updates a user's profile
func (r *UserRepository) UpdateUser(userID, fullName, email string) error {
	stmt, err := r.db.Prepare("UPDATE users SET full_name = ?, email = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(fullName, email, userID)
	return err
}
