package main

import (
	"log"

	"github.com/chess-puzzle-app/backend/internal/auth"
	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/chess-puzzle-app/backend/internal/database"
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

	authSvc := auth.NewAuthService(db, cfg)
	userSvc := users.NewUserService(db)
	settingSvc := settings.NewSettingService(db)

	authHandler := auth.NewAuthHandler(authSvc)
	userHandler := users.NewUserHandler(userSvc)
	settingHandler := settings.NewSettingHandler(settingSvc)

	api := r.Group("/api/v1")
	{
		authGroup := api.Group("/auth")
		{
			authGroup.POST("/google", authHandler.GoogleLogin)
			authGroup.POST("/guest", authHandler.GuestLogin)
		}

		api.GET("/user/me", userHandler.GetMe) // Auth middleware might be needed here later
		api.GET("/settings", settingHandler.List)
	}

	log.Printf("Identity Service starting on port 8081")
	if err := r.Run(":8081"); err != nil {
		log.Fatalf("Could not start service: %v", err)
	}
}
