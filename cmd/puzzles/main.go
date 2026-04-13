package main

import (
	"log"

	"github.com/chess-puzzle-app/backend/internal/categories"
	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/database"
	"github.com/chess-puzzle-app/backend/internal/puzzles"
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

	puzzleSvc := puzzles.NewPuzzleService(db)
	categorySvc := categories.NewCategoryService(db)

	puzzleHandler := puzzles.NewPuzzleHandler(puzzleSvc)
	categoryHandler := categories.NewCategoryHandler(categorySvc)

	api := r.Group("/api/v1")
	{
		puzzlesGroup := api.Group("/puzzles")
		{
			puzzlesGroup.GET("", puzzleHandler.ListPuzzles)
			puzzlesGroup.GET("/:id", puzzleHandler.GetPuzzle)
		}

		categoriesGroup := api.Group("/categories")
		{
			categoriesGroup.GET("", categoryHandler.List)
		}
	}

	log.Printf("Puzzle Service starting on port 8082")
	if err := r.Run(":8082"); err != nil {
		log.Fatalf("Could not start service: %v", err)
	}
}
