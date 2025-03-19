package api

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/panaalexandrucristian/feedback-collector/internal/api/handlers"
	"github.com/panaalexandrucristian/feedback-collector/internal/api/middleware"
	"github.com/panaalexandrucristian/feedback-collector/internal/config"
	"github.com/panaalexandrucristian/feedback-collector/internal/db"
)

// SetupRouter configures the HTTP router
func SetupRouter(cfg *config.Config, db *db.Database) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Configure CORS
	router.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.AllowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Create handlers
	authHandler := handlers.NewAuthHandler(db, cfg)
	roomHandler := handlers.NewRoomHandler(db)
	feedbackHandler := handlers.NewFeedbackHandler(db)

	// Auth routes
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.GET("/me", middleware.AuthMiddleware(cfg), authHandler.GetCurrentUser)
	}

	// Room routes
	rooms := router.Group("/api/rooms")
	{
		rooms.Use(middleware.AuthMiddleware(cfg))
		rooms.POST("", roomHandler.CreateRoom)
		rooms.GET("", roomHandler.GetRooms)
		rooms.GET("/:id", roomHandler.GetRoomByID)
	}

	// Public room access
	publicRooms := router.Group("/api/public/rooms")
	{
		publicRooms.GET("/:id", roomHandler.GetRoomByID)
		publicRooms.POST("/:id/join", roomHandler.JoinRoom)
	}

	// Feedback routes
	feedback := router.Group("/api/public/rooms/:id/feedback")
	{
		feedback.POST("", feedbackHandler.CreateFeedback)
	}

	// Protected feedback retrieval (only for room creators)
	protectedFeedback := router.Group("/api/rooms/:id/feedback")
	{
		protectedFeedback.Use(middleware.AuthMiddleware(cfg))
		protectedFeedback.GET("", feedbackHandler.GetFeedback)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	return router
}
