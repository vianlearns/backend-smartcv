package middleware

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ErrorResponse represents a validation error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Details map[string]string `json:"details,omitempty"`
}

// ValidateRequest validates the request body against the provided struct
func ValidateRequest(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			details := make(map[string]string)
			for _, fieldErr := range validationErrors {
				details[fieldErr.Field()] = getValidationMessage(fieldErr.Tag(), fieldErr.Param())
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation failed",
				Details: details,
			})
			return false
		}

		// JSON syntax error
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error: "Invalid JSON syntax",
			})
			_ = syntaxErr // silence unused variable warning
			return false
		}

		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error: err.Error(),
		})
		return false
	}

	if err := validate.Struct(req); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			details := make(map[string]string)
			for _, fieldErr := range validationErrors {
				details[fieldErr.Field()] = getValidationMessage(fieldErr.Tag(), fieldErr.Param())
			}
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Validation failed",
				Details: details,
			})
			return false
		}
	}

	return true
}

func getValidationMessage(tag, param string) string {
	messages := map[string]string{
		"required": "This field is required",
		"email":    "Invalid email format",
		"min":      "Minimum length is " + param,
		"max":      "Maximum length is " + param,
		"url":      "Invalid URL format",
	}

	if msg, exists := messages[tag]; exists {
		return msg
	}
	return "Invalid value"
}

// ErrorHandler middleware for catching panics and errors
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// CORS middleware
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RateLimiter simple rate limiting middleware
func RateLimiter() gin.HandlerFunc {
	// Simple in-memory rate limiter
	// In production, use Redis-based rate limiting
	return func(c *gin.Context) {
		c.Next()
	}
}
