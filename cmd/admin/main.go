package main

import (
	"log"

	"github.com/chess-puzzle-app/backend/internal/admin"
	"github.com/chess-puzzle-app/backend/internal/categories"
	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/chess-puzzle-app/backend/internal/puzzles"
	"github.com/chess-puzzle-app/backend/internal/settings"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Could not initialize database: %v", err)
	}

	r := gin.Default()

	// CORS Middleware
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

	adminSvc := admin.NewAdminService(db)
	puzzleSvc := puzzles.NewPuzzleService(db)
	categorySvc := categories.NewCategoryService(db)
	settingSvc := settings.NewSettingService(db)

	// In microservices, the admin service might not have direct access to the engine workers 
	// but we can initialize a dummy engine service for status reporting if needed, 
	// or ideally, it should query the Engine Service API.
	// For now, we'll keep it simple to restore functionality.
	adminHandler := admin.NewAdminHandler(adminSvc, nil) 
	puzzleHandler := puzzles.NewPuzzleHandler(puzzleSvc)
	categoryHandler := categories.NewCategoryHandler(categorySvc)
	settingHandler := settings.NewSettingHandler(settingSvc)

	api := r.Group("/api/v1/admin")
	{
		api.GET("/stats", adminHandler.GetStats)
		api.GET("/users", adminHandler.ListUsers)
		api.GET("/engine/status", adminHandler.GetEngineStatus)

		// Puzzles management
		api.POST("/puzzles", puzzleHandler.CreatePuzzle)
		api.PUT("/puzzles/:id", puzzleHandler.UpdatePuzzle)
		api.DELETE("/puzzles/:id", puzzleHandler.DeletePuzzle)

		// Categories management
		api.GET("/categories", categoryHandler.List)
		api.POST("/categories", categoryHandler.Create)
		api.PUT("/categories/:id", categoryHandler.Update)
		api.DELETE("/categories/:id", categoryHandler.Delete)

		// Settings management
		api.GET("/settings", settingHandler.List)
		api.POST("/settings", settingHandler.Update)
	}

	log.Printf("Admin Service starting on port 8084")
	if err := r.Run(":8084"); err != nil {
		log.Fatalf("Could not start admin service: %v", err)
	}
}
