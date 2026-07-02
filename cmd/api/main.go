package main

import (
	"log"

	"github.com/chess-puzzle-app/backend/internal/admin"
	"github.com/chess-puzzle-app/backend/internal/analysis"
	"github.com/chess-puzzle-app/backend/internal/auth"
	"github.com/chess-puzzle-app/backend/internal/categories"
	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/chess-puzzle-app/backend/internal/difficulties"
	"github.com/chess-puzzle-app/backend/internal/engine"
	"github.com/chess-puzzle-app/backend/internal/learning"
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

	// Status Endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "Online",
			"system": "Tactics Master Monolith",
			"version": "1.0.0",
			"endpoints": gin.H{
				"api": "/api/v1",
			},
		})
	})

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
	difficultySvc := difficulties.NewDifficultyService(db)
	learningSvc := learning.NewLearningService(db)

	engineSvc, err := engine.NewEngineService(cfg.StockfishPath, cfg.EnginePoolSize)
	if err != nil {
		log.Printf("Warning: Could not start chess engine: %v", err)
	} else {
		defer engineSvc.Stop()
	}

	// Handlers
	authHandler := auth.NewAuthHandler(authSvc)
	puzzleHandler := puzzles.NewPuzzleHandler(puzzleSvc, userSvc)
	userHandler := users.NewUserHandler(userSvc)
	categoryHandler := categories.NewCategoryHandler(categorySvc)
	settingHandler := settings.NewSettingHandler(settingSvc)
	difficultyHandler := difficulties.NewDifficultyHandler(difficultySvc)
	learningHandler := learning.NewLearningHandler(learningSvc)
	analysisSvc := analysis.NewAnalysisService(db)
	analysisHandler := analysis.NewAnalysisHandler(analysisSvc)

	var engineHandler *engine.EngineHandler
	if engineSvc != nil {
		engineHandler = engine.NewEngineHandler(engineSvc, userSvc)
	}

	adminHandler := admin.NewAdminHandler(adminSvc, engineSvc, engineHandler)

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
			userGroup.POST("/hint", userHandler.UseHint)
			userGroup.POST("/add-xp", userHandler.AddXP)
			userGroup.POST("/progress", userHandler.SaveProgress)
			userGroup.GET("/progress", userHandler.GetProgressList)
		}

		// Engine
		if engineHandler != nil {
			engineGroup := api.Group("/engine")
			engineGroup.Use(middleware.AuthMiddleware(cfg))
			engineGroup.Use(middleware.UsageMiddleware(userSvc))
			{
				engineGroup.POST("/analyze", engineHandler.Analyze)
				engineGroup.POST("/play", engineHandler.Play)
			}
		}

		// Puzzles
		puzzleGroup := api.Group("/puzzles")
		puzzleGroup.Use(middleware.AuthMiddleware(cfg))
		puzzleGroup.Use(middleware.UsageMiddleware(userSvc))
		{
			puzzleGroup.GET("", puzzleHandler.ListPuzzles)
			puzzleGroup.GET("/daily", puzzleHandler.GetDailyPuzzle)
			puzzleGroup.GET("/:id", puzzleHandler.GetPuzzle)
			puzzleGroup.POST("/:id/solve", puzzleHandler.SolvePuzzle)
		}

		// Difficulties (Public for selection)
		api.GET("/difficulties", difficultyHandler.List)
		api.GET("/categories", categoryHandler.List)

		// Learning LMS (Public / Auth Protected for Mobile App)
		learningGroup := api.Group("/learning")
		learningGroup.Use(middleware.AuthMiddleware(cfg))
		{
			learningGroup.GET("/categories", learningHandler.ListCategories)
			learningGroup.GET("/categories/:id", learningHandler.GetCategory)
			learningGroup.GET("/lessons", learningHandler.ListLessons)
			learningGroup.GET("/lessons/:id", learningHandler.GetLesson)
			learningGroup.GET("/lessons/:id/steps", learningHandler.ListSteps)
			learningGroup.GET("/steps/:id", learningHandler.GetStep)
		}

		// Analysis Sessions (Protected) — save/resume analysis move trees
		analysisGroup := api.Group("/analysis")
		analysisGroup.Use(middleware.AuthMiddleware(cfg))
		{
			analysisGroup.GET("", analysisHandler.List)
			analysisGroup.POST("", analysisHandler.Create)
			analysisGroup.GET("/:id", analysisHandler.Get)
			analysisGroup.PUT("/:id", analysisHandler.Update)
			analysisGroup.DELETE("/:id", analysisHandler.Delete)
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

			// Difficulties
			adminGroup.POST("/difficulties", difficultyHandler.Create)
			adminGroup.PUT("/difficulties/:id", difficultyHandler.Update)
			adminGroup.DELETE("/difficulties/:id", difficultyHandler.Delete)

			// Settings
			adminGroup.GET("/settings", settingHandler.List)
			adminGroup.POST("/settings", settingHandler.Update)

			// Learning LMS Admin
			adminGroup.GET("/learning/categories", learningHandler.ListCategories)
			adminGroup.POST("/learning/categories", learningHandler.CreateCategory)
			adminGroup.PUT("/learning/categories/:id", learningHandler.UpdateCategory)
			adminGroup.DELETE("/learning/categories/:id", learningHandler.DeleteCategory)

			adminGroup.GET("/learning/lessons", learningHandler.ListLessons)
			adminGroup.POST("/learning/lessons", learningHandler.CreateLesson)
			adminGroup.PUT("/learning/lessons/:id", learningHandler.UpdateLesson)
			adminGroup.DELETE("/learning/lessons/:id", learningHandler.DeleteLesson)

			adminGroup.GET("/learning/lessons/:id/steps", learningHandler.ListSteps)
			adminGroup.POST("/learning/lessons/:id/steps", learningHandler.CreateStep)
			adminGroup.PUT("/learning/steps/:id", learningHandler.UpdateStep)
			adminGroup.DELETE("/learning/steps/:id", learningHandler.DeleteStep)
			adminGroup.PUT("/learning/lessons/:id/steps/reorder", learningHandler.ReorderSteps)
		}
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Could not start server: %v", err)
	}
}
