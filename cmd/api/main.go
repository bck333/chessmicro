package main

import (
	"log"

	"github.com/chess-puzzle-app/backend/internal/admin"
	"github.com/chess-puzzle-app/backend/internal/auth"
	"github.com/chess-puzzle-app/backend/internal/categories"
	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/chess-puzzle-app/backend/internal/engine"
	"github.com/chess-puzzle-app/backend/internal/middleware"
	"github.com/chess-puzzle-app/backend/internal/puzzles"
	"github.com/chess-puzzle-app/backend/internal/settings"
	"github.com/chess-puzzle-app/backend/internal/users"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	r := gin.Default()

	// CORS Middleware (Simple for Phase 1)
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Services
	authSvc := auth.NewAuthService(db, cfg)
	puzzleSvc := puzzles.NewPuzzleService(db)
	userSvc := users.NewUserService(db)
	adminSvc := admin.NewAdminService(db)
	categorySvc := categories.NewCategoryService(db)
	settingSvc := settings.NewSettingService(db)

	engineSvc, err := engine.NewEngineService(cfg.StockfishPath, 2) // Pool of 2 for dev
	if err != nil {
		log.Printf("Warning: Could not start chess engine: %v", err)
	} else {
		defer engineSvc.Stop()
	}

	// Handlers
	authHandler := auth.NewAuthHandler(authSvc)
	puzzleHandler := puzzles.NewPuzzleHandler(puzzleSvc)
	userHandler := users.NewUserHandler(userSvc)
	adminHandler := admin.NewAdminHandler(adminSvc, engineSvc)
	categoryHandler := categories.NewCategoryHandler(categorySvc)
	settingHandler := settings.NewSettingHandler(settingSvc)

	var engineHandler *engine.EngineHandler
	if engineSvc != nil {
		engineHandler = engine.NewEngineHandler(engineSvc)
	}

	api := r.Group("/api/v1")
	{
		// Auth
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/google", authHandler.GoogleLogin)
			authGroup.POST("/guest", authHandler.GuestLogin)
		}

		// User (Protected)
		userGroup := api.Group("/user")
		userGroup.Use(middleware.AuthMiddleware(cfg))
		{
			userGroup.GET("/me", userHandler.GetMe)
		}

		// Engine
		if engineHandler != nil {
			engineGroup := api.Group("/engine")
			{
				engineGroup.POST("/analyze", engineHandler.Analyze)
				engineGroup.POST("/play", engineHandler.Play)
			}
		}

		// Puzzles
		puzzleGroup := api.Group("/puzzles")
		{
			puzzleGroup.GET("", puzzleHandler.ListPuzzles)
			puzzleGroup.GET("/:id", puzzleHandler.GetPuzzle)
			
			// Protected Puzzle Routes
			puzzleProtected := puzzleGroup.Group("")
			puzzleProtected.Use(middleware.AuthMiddleware(cfg))
			{
				puzzleProtected.POST("/:id/solve", puzzleHandler.SolvePuzzle)
			}
		}

		// Admin (Protected)
		adminGroup := api.Group("/admin")
		adminGroup.Use(middleware.AuthMiddleware(cfg))
		{
			// Stats & Users
			adminGroup.GET("/stats", adminHandler.GetStats)
			adminGroup.GET("/users", adminHandler.ListUsers)
			adminGroup.GET("/engine/status", adminHandler.GetEngineStatus)

			// Puzzles
			adminGroup.POST("/puzzles", puzzleHandler.CreatePuzzle)
			adminGroup.PUT("/puzzles/:id", puzzleHandler.UpdatePuzzle)
			adminGroup.DELETE("/puzzles/:id", puzzleHandler.DeletePuzzle)

			// Categories
			adminGroup.GET("/categories", categoryHandler.List)
			adminGroup.POST("/categories", categoryHandler.Create)
			adminGroup.PUT("/categories/:id", categoryHandler.Update)
			adminGroup.DELETE("/categories/:id", categoryHandler.Delete)

			// Settings
			adminGroup.GET("/settings", settingHandler.List)
			adminGroup.POST("/settings", settingHandler.Update)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
