package middleware

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Decode JWT payload to extract Clerk user ID (sub claim)
		clerkUserID, err := extractClerkUserID(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// Store the Clerk user ID (not the raw token)
		c.Set("clerk_user_id", clerkUserID)
		c.Set("token", token)
		c.Next()
	}
}

// extractClerkUserID decodes the JWT payload (without verification) to extract the "sub" claim.
// In production, you should verify the JWT signature using Clerk's public keys.
func extractClerkUserID(token string) (string, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid JWT format")
	}

	// Decode the payload (second part)
	payload := parts[1]
	// Add padding if necessary
	switch len(payload) % 4 {
	case 2:
		payload += "=="
	case 3:
		payload += "="
	}

	decoded, err := base64.URLEncoding.DecodeString(payload)
	if err != nil {
		// Try without padding
		decoded, err = base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			return "", err
		}
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		return "", err
	}

	sub, ok := claims["sub"].(string)
	if !ok || sub == "" {
		return "", fmt.Errorf("missing sub claim in JWT")
	}

	return sub, nil
}

func GetUserFromContext(c *gin.Context) string {
	clerkUserID, exists := c.Get("clerk_user_id")
	if !exists {
		return ""
	}
	return clerkUserID.(string)
}
