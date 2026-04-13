package main

import (
	"log"

	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/engine"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

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

	engineSvc, err := engine.NewEngineService(cfg.StockfishPath, 2)
	if err != nil {
		log.Fatalf("Could not start engine service: %v", err)
	}
	defer engineSvc.Stop()

	engineHandler := engine.NewEngineHandler(engineSvc)

	api := r.Group("/api/v1/engine")
	{
		api.POST("/analyze", engineHandler.Analyze)
		api.POST("/play", engineHandler.Play)
	}

	log.Printf("Engine Service starting on port 8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatalf("Could not start service: %v", err)
	}
}
