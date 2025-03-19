package models

import (
	"time"
)

// User represents a registered user of the application
type User struct {
	ID               int       `json:"id" db:"id"`
	Email            string    `json:"email" db:"email"`
	PasswordHash     string    `json:"-" db:"password_hash"` // Never expose in JSON responses
	SubscriptionType string    `json:"subscription_type" db:"subscription_type"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// Room represents a feedback collection room
type Room struct {
	ID                  string    `json:"id" db:"id"`
	Name                string    `json:"name" db:"name"`
	Password            string    `json:"-" db:"password"` // Never expose in JSON responses
	CreatorID           int       `json:"creator_id" db:"creator_id"`
	IsPasswordProtected bool      `json:"is_password_protected" db:"is_password_protected"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
}

// Feedback represents a piece of feedback submitted in a room
type Feedback struct {
	ID        int       `json:"id" db:"id"`
	RoomID    string    `json:"room_id" db:"room_id"`
	Content   string    `json:"content" db:"content"`
	Sentiment string    `json:"sentiment" db:"sentiment"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Auth Request/Response types
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// Room Request/Response types
type CreateRoomRequest struct {
	Name     string `json:"name" binding:"required"`
	Password string `json:"password"`
}

type JoinRoomRequest struct {
	Password string `json:"password"`
}

type CreateFeedbackRequest struct {
	Content string `json:"content" binding:"required"`
}
