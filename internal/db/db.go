package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"

	"github.com/panaalexandrucristian/feedback-collector/internal/models"
)

// Database wraps the SQL database connection
type Database struct {
	*sql.DB
}

// New creates a new Database instance
func New(dataSourceName string) (*Database, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("error opening database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Check connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error connecting to database: %w", err)
	}

	log.Println("Database connection established")
	return &Database{db}, nil
}

// RunMigrations applies database migrations from the specified path
func (db *Database) RunMigrations(migrationsPath string) error {
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to initialize migrations: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}

// InitSchema creates the necessary database tables if they don't exist
func (db *Database) InitSchema() error {
	schema := `
    -- Users table
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        password_hash VARCHAR(255) NOT NULL,
        subscription_type VARCHAR(20) NOT NULL DEFAULT 'free',
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

    -- Rooms table
    CREATE TABLE IF NOT EXISTS rooms (
        id VARCHAR(10) PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        password VARCHAR(255),
        creator_id INTEGER REFERENCES users(id),
        is_password_protected BOOLEAN NOT NULL DEFAULT FALSE,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );

    -- Feedback table
    CREATE TABLE IF NOT EXISTS feedback (
        id SERIAL PRIMARY KEY,
        room_id VARCHAR(10) REFERENCES rooms(id),
        content TEXT NOT NULL,
        sentiment VARCHAR(20),
        created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
    );
    `

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("error initializing database schema: %w", err)
	}
	log.Println("Database schema initialized")
	return nil
}

// ===== User-related functions =====

// CreateUser creates a new user in the database
func (db *Database) CreateUser(email, password string) (*models.User, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("error hashing password: %w", err)
	}

	user := &models.User{}
	query := `
        INSERT INTO users (email, password_hash, subscription_type)
        VALUES ($1, $2, 'free')
        RETURNING id, email, subscription_type, created_at
    `

	err = db.QueryRow(query, email, string(hashedPassword)).Scan(
		&user.ID,
		&user.Email,
		&user.SubscriptionType,
		&user.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating user: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email address
func (db *Database) GetUserByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, email, password_hash, subscription_type, created_at 
        FROM users WHERE email = $1
    `

	err := db.QueryRow(query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.SubscriptionType,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (db *Database) GetUserByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
        SELECT id, email, subscription_type, created_at 
        FROM users WHERE id = $1
    `

	err := db.QueryRow(query, id).Scan(
		&user.ID,
		&user.Email,
		&user.SubscriptionType,
		&user.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("error retrieving user: %w", err)
	}

	return user, nil
}

// ===== Room-related functions =====

// CreateRoom creates a new room in the database
func (db *Database) CreateRoom(roomID, name string, creatorID int, password string) (*models.Room, error) {
	room := &models.Room{}
	isPasswordProtected := password != ""

	query := `
        INSERT INTO rooms (id, name, password, creator_id, is_password_protected)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, name, creator_id, is_password_protected, created_at
    `

	err := db.QueryRow(query, roomID, name, password, creatorID, isPasswordProtected).Scan(
		&room.ID,
		&room.Name,
		&room.CreatorID,
		&room.IsPasswordProtected,
		&room.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating room: %w", err)
	}

	return room, nil
}

// GetRoomByID retrieves a room by its ID
func (db *Database) GetRoomByID(id string) (*models.Room, error) {
	room := &models.Room{}
	query := `
        SELECT id, name, password, creator_id, is_password_protected, created_at
        FROM rooms WHERE id = $1
    `

	err := db.QueryRow(query, id).Scan(
		&room.ID,
		&room.Name,
		&room.Password,
		&room.CreatorID,
		&room.IsPasswordProtected,
		&room.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("room not found")
		}
		return nil, fmt.Errorf("error retrieving room: %w", err)
	}

	return room, nil
}

// GetRoomsByUserID retrieves all rooms created by a specific user
func (db *Database) GetRoomsByUserID(userID int) ([]models.Room, error) {
	query := `
        SELECT id, name, creator_id, is_password_protected, created_at
        FROM rooms
        WHERE creator_id = $1
        ORDER BY created_at DESC
    `

	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving rooms: %w", err)
	}
	defer rows.Close()

	var rooms []models.Room
	for rows.Next() {
		var r models.Room
		err := rows.Scan(
			&r.ID,
			&r.Name,
			&r.CreatorID,
			&r.IsPasswordProtected,
			&r.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning room row: %w", err)
		}
		rooms = append(rooms, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rooms: %w", err)
	}

	return rooms, nil
}

// ===== Feedback-related functions =====

// CreateFeedback creates a new feedback entry
func (db *Database) CreateFeedback(roomID, content string) (*models.Feedback, error) {
	feedback := &models.Feedback{}
	query := `
        INSERT INTO feedback (room_id, content)
        VALUES ($1, $2)
        RETURNING id, room_id, content, sentiment, created_at
    `

	err := db.QueryRow(query, roomID, content).Scan(
		&feedback.ID,
		&feedback.RoomID,
		&feedback.Content,
		&feedback.Sentiment,
		&feedback.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("error creating feedback: %w", err)
	}

	return feedback, nil
}

// GetFeedbackByRoomID retrieves all feedback for a specific room
func (db *Database) GetFeedbackByRoomID(roomID string) ([]models.Feedback, error) {
	query := `
        SELECT id, room_id, content, sentiment, created_at
        FROM feedback
        WHERE room_id = $1
        ORDER BY created_at DESC
    `

	rows, err := db.Query(query, roomID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving feedback: %w", err)
	}
	defer rows.Close()

	var feedbackList []models.Feedback
	for rows.Next() {
		var f models.Feedback
		err := rows.Scan(
			&f.ID,
			&f.RoomID,
			&f.Content,
			&f.Sentiment,
			&f.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning feedback row: %w", err)
		}
		feedbackList = append(feedbackList, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feedback: %w", err)
	}

	return feedbackList, nil
}

// UpdateFeedbackSentiment updates the sentiment analysis result for a feedback entry
func (db *Database) UpdateFeedbackSentiment(feedbackID int, sentiment string) error {
	query := `UPDATE feedback SET sentiment = $1 WHERE id = $2`

	_, err := db.Exec(query, sentiment, feedbackID)
	if err != nil {
		return fmt.Errorf("error updating feedback sentiment: %w", err)
	}

	return nil
}
