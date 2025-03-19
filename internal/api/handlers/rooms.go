package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/panaalexandrucristian/feedback-collector/internal/api/middleware"
	"github.com/panaalexandrucristian/feedback-collector/internal/db"
	"github.com/panaalexandrucristian/feedback-collector/internal/models"
)

// RoomHandler handles room-related routes
type RoomHandler struct {
	DB *db.Database
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(db *db.Database) *RoomHandler {
	return &RoomHandler{DB: db}
}

// CreateRoom handles creating a new feedback room
func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req models.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	// Generate a unique room ID - 6 characters alphanumeric
	roomID := uuid.New().String()[:6]

	// Hash password if provided
	var passwordHash string
	if req.Password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process password"})
			return
		}
		passwordHash = string(hash)
	}

	// Create room
	room, err := h.DB.CreateRoom(roomID, req.Name, userID, passwordHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, room)
}

// GetRooms returns all rooms created by the authenticated user
func (h *RoomHandler) GetRooms(c *gin.Context) {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
		return
	}

	rooms, err := h.DB.GetRoomsByUserID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve rooms"})
		return
	}

	c.JSON(http.StatusOK, rooms)
}

// GetRoomByID retrieves a room by ID
func (h *RoomHandler) GetRoomByID(c *gin.Context) {
	roomID := c.Param("id")

	room, err := h.DB.GetRoomByID(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Remove password from response
	room.Password = ""

	c.JSON(http.StatusOK, room)
}

// JoinRoom handles joining a password-protected room
func (h *RoomHandler) JoinRoom(c *gin.Context) {
	roomID := c.Param("id")

	var req models.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get room
	room, err := h.DB.GetRoomByID(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Check if room requires password
	if room.IsPasswordProtected {
		// Verify password
		err = bcrypt.CompareHashAndPassword([]byte(room.Password), []byte(req.Password))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid room password"})
			return
		}
	}

	// Remove password from response
	room.Password = ""

	c.JSON(http.StatusOK, room)
}
