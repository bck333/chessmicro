package middleware

import (
	"net/http"

	"github.com/chess-puzzle-app/backend/internal/users"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func UsageMiddleware(userSvc *users.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("user_id")
		if !exists {
			// If no auth, assume guest for now or abort if auth is required
			c.Set("subscription_type", "guest")
			c.Next()
			return
		}

		userID, ok := userIDStr.(uuid.UUID)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user id in context"})
			c.Abort()
			return
		}

		user, err := userSvc.GetUserByID(userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			c.Abort()
			return
		}

		// Inject subscription info for other handlers (like engine)
		c.Set("subscription_type", user.SubscriptionType)
		c.Set("user_object", user)
		
		c.Next()
	}
}
