package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"time"

	"github.com/chess-puzzle-app/backend/internal/config"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	r := gin.Default()

	// CORS Middleware
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "Online",
			"system": "Tactics Master Microservices Mesh",
			"endpoints": gin.H{
				"admin": "/admin",
				"api": "/api/v1",
			},
		})
	})

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

	identityTarget, _ := url.Parse(cfg.IdentityURL)
	puzzlesTarget, _ := url.Parse(cfg.PuzzlesURL)
	engineTarget, _ := url.Parse(cfg.EngineURL)
	adminTarget, _ := url.Parse(cfg.AdminURL)
	adminFrontendTarget, _ := url.Parse(cfg.AdminFrontendURL)

	// Create customized proxies with timeouts
	newProxy := func(target *url.URL, serviceName string) *httputil.ReverseProxy {
		proxy := httputil.NewSingleHostReverseProxy(target)
		
		// Professional Error Handler
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("[%s ERROR]: %v", serviceName, err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error": "Service temporarily unavailable", "code": 503, "service": "` + serviceName + `"}`))
		}

		// Standard Transport with Timeouts
		proxy.Transport = &http.Transport{
			ResponseHeaderTimeout: 10 * time.Second,
			IdleConnTimeout:        30 * time.Second,
		}

		return proxy
	}

	identityProxy := newProxy(identityTarget, "IDENTITY")
	puzzlesProxy := newProxy(puzzlesTarget, "PUZZLES")
	engineProxy := newProxy(engineTarget, "ENGINE")
	adminProxy := newProxy(adminTarget, "ADMIN_API")
	adminFrontendProxy := newProxy(adminFrontendTarget, "ADMIN_UI")

	// Admin UI Routing
	r.Any("/admin/*path", func(c *gin.Context) {
		adminFrontendProxy.ServeHTTP(c.Writer, c.Request)
	})

	r.Any("/api/v1/*path", func(c *gin.Context) {
		path := c.Param("path")
		
		var proxy *httputil.ReverseProxy
		
		if strings.HasPrefix(path, "/auth") || strings.HasPrefix(path, "/user") || strings.HasPrefix(path, "/settings") {
			proxy = identityProxy
		} else if strings.HasPrefix(path, "/puzzles") || strings.HasPrefix(path, "/categories") {
			proxy = puzzlesProxy
		} else if strings.HasPrefix(path, "/engine") {
			proxy = engineProxy
		} else if strings.HasPrefix(path, "/admin") {
			proxy = adminProxy
		}

		if proxy != nil {
			proxy.ServeHTTP(c.Writer, c.Request)
		} else {
			c.JSON(404, gin.H{"error": "Service not found in gateway"})
		}
	})

	log.Printf("Gateway starting on port 8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Could not start gateway: %v", err)
	}
}
