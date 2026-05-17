package main

import (
	"log"
	"smartcv-backend/internal/config"
	"smartcv-backend/internal/database"
	"smartcv-backend/internal/handlers"
	"smartcv-backend/internal/routes"
	"smartcv-backend/internal/services"
)

func main() {
	// Load config
	cfg := config.Load()

	// Initialize database
	database.InitDB(cfg)

	// Initialize AI service
	aiService := services.NewAIService(cfg.AIBaseURL, cfg.AIModel, cfg.AIAPIKey)
	handlers.InitAI(aiService)

	// Setup router
	router := routes.SetupRouter(cfg)

	// Start server
	log.Printf("SmartCV API Server starting on port %s", cfg.Port)
	log.Printf("AI Model: %s @ %s", cfg.AIModel, cfg.AIBaseURL)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
