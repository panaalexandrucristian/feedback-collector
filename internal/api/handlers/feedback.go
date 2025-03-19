package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/panaalexandrucristian/feedback-collector/internal/db"
	"github.com/panaalexandrucristian/feedback-collector/internal/models"
)

// FeedbackHandler handles feedback-related routes
type FeedbackHandler struct {
	DB *db.Database
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(db *db.Database) *FeedbackHandler {
	return &FeedbackHandler{DB: db}
}

// CreateFeedback handles creating a new feedback entry
func (h *FeedbackHandler) CreateFeedback(c *gin.Context) {
	roomID := c.Param("id")

	var req models.CreateFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify room exists
	room, err := h.DB.GetRoomByID(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Create feedback
	feedback, err := h.DB.CreateFeedback(room.ID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save feedback"})
		return
	}

	c.JSON(http.StatusCreated, feedback)
}

// GetFeedback retrieves all feedback for a room
func (h *FeedbackHandler) GetFeedback(c *gin.Context) {
	roomID := c.Param("id")

	// Check if room exists
	_, err := h.DB.GetRoomByID(roomID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Get all feedback for the room
	feedback, err := h.DB.GetFeedbackByRoomID(roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve feedback"})
		return
	}

	c.JSON(http.StatusOK, feedback)
}
