package routes

import (
	"smartcv-backend/internal/config"
	"smartcv-backend/internal/handlers"
	"smartcv-backend/internal/middleware"

	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter(cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// CORS middleware
	allowedOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
	for i := range allowedOrigins {
		allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
	}

	router.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// Health check
		v1.GET("/health", handlers.HealthCheck)

		// Public routes
		v1.POST("/midtrans/callback", handlers.MidtransCallback)

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User
			protected.GET("/user", handlers.GetUser)
			protected.PUT("/user", handlers.UpdateUser)

			// Profile
			protected.GET("/profile", handlers.GetProfile)
			protected.PUT("/profile", handlers.UpdateProfile)

			// Experience
			protected.GET("/experiences", handlers.GetExperiences)
			protected.POST("/experiences", handlers.CreateExperience)
			protected.PUT("/experiences/:id", handlers.UpdateExperience)
			protected.DELETE("/experiences/:id", handlers.DeleteExperience)

			// Education
			protected.GET("/educations", handlers.GetEducations)
			protected.POST("/educations", handlers.CreateEducation)
			protected.PUT("/educations/:id", handlers.UpdateEducation)
			protected.DELETE("/educations/:id", handlers.DeleteEducation)

			// Skills
			protected.GET("/skills", handlers.GetSkills)
			protected.POST("/skills", handlers.CreateSkill)
			protected.PUT("/skills/:id", handlers.UpdateSkill)
			protected.DELETE("/skills/:id", handlers.DeleteSkill)

			// Certifications
			protected.GET("/certifications", handlers.GetCertifications)
			protected.POST("/certifications", handlers.CreateCertification)
			protected.DELETE("/certifications/:id", handlers.DeleteCertification)

			// Projects
			protected.GET("/projects", handlers.GetProjects)
			protected.POST("/projects", handlers.CreateProject)
			protected.PUT("/projects/:id", handlers.UpdateProject)
			protected.DELETE("/projects/:id", handlers.DeleteProject)

			// Job Applications
			protected.GET("/jobs", handlers.GetJobApplications)
			protected.POST("/jobs", handlers.CreateJobApplication)
			protected.GET("/jobs/:id", handlers.GetJobApplication)
			protected.PUT("/jobs/:id", handlers.UpdateJobApplication)
			protected.DELETE("/jobs/:id", handlers.DeleteJobApplication)

			// CV Generation
			protected.POST("/cv/generate", handlers.GenerateCV)
			protected.GET("/cv", handlers.GetCVs)
			protected.GET("/cv/:id", handlers.GetCV)
			protected.PUT("/cv/:id", handlers.UpdateCV)

			// Comments
			protected.GET("/cv/:id/comments", handlers.GetComments)
			protected.POST("/cv/:id/comments", handlers.CreateComment)
			protected.PATCH("/cv/comments/:id", handlers.ResolveComment)

			// AI Chat
			protected.POST("/chat", handlers.ChatWithAI)
			protected.GET("/chat/:job_id", handlers.GetChatHistory)

			// Gap Analysis
			protected.POST("/analyze", handlers.AnalyzeGap)

			// CV Upload & Parse
			protected.POST("/cv/upload", handlers.UploadCV)

			// Credits
			protected.GET("/credits", handlers.GetCredits)
			protected.POST("/credits/purchase", handlers.PurchaseCredits)

			// PDF Download
			protected.GET("/cv/:id/download", handlers.DownloadCV)
		}
	}

	return router
}
