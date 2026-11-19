package main

import (
	"log"
	"os"

	"github.com/VerryHdyt/cloudforge/backend/internal/database"
	"github.com/VerryHdyt/cloudforge/backend/internal/handlers"
	"github.com/VerryHdyt/cloudforge/backend/internal/middleware"
	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Connect to PostgreSQL
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database connection: %v", err)
	}
	defer sqlDB.Close()

	log.Println("PostgreSQL connected successfully")

	// Run database migrations
	if err := database.Migrate(db); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}

	log.Println("Database migrations completed successfully")

	// Initialize Audit Service
	auditService := &services.AuditService{
		DB: db,
	}

	// Initialize Docker Service
	dockerService, err := services.NewDockerService()
	if err != nil {
		log.Fatalf("Failed to connect to Docker Engine: %v", err)
	}
	defer dockerService.Client.Close()

	log.Println("Docker Engine connected successfully")

	// Initialize Gin router
	router := gin.Default()

	// Disable trusted proxies
	if err := router.SetTrustedProxies(nil); err != nil {
		log.Fatal(err)
	}

	// Health check
	router.GET("/health", handlers.HealthCheck)

	// API version 1
	api := router.Group("/api/v1")
	{
		// Public routes
		api.GET("/info", handlers.PlatformInfo)

		// Authentication handler
		authHandler := &handlers.AuthHandler{
			DB:    db,
			Audit: auditService,
		}

		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		// User management handler
		userHandler := &handlers.UserHandler{
			DB:    db,
			Audit: auditService,
		}

		// Audit handler
		auditHandler := &handlers.AuditHandler{
			DB: db,
		}

		// Docker handler
		dockerHandler := &handlers.DockerHandler{
			Docker: dockerService,
			Audit:  auditService,
		}

		// Protected routes (JWT Authentication required)
		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(db))

		// Admin-only routes (JWT + RBAC required)
		admin := protected.Group("")
		admin.Use(middleware.RequireRoles("admin"))
		{
			// User management
			admin.POST("/users", userHandler.CreateUser)
			admin.GET("/users", userHandler.GetUsers)
			admin.GET("/users/:id", userHandler.GetUser)
			admin.PUT("/users/:id", userHandler.UpdateUser)
			admin.DELETE("/users/:id", userHandler.DeleteUser)
			admin.GET("/audit-logs", auditHandler.GetAuditLogs)
			admin.GET("/containers", dockerHandler.ListContainers)
			admin.POST("/containers/:id/start", dockerHandler.StartContainer)
			admin.POST("/containers/:id/stop", dockerHandler.StopContainer)
			admin.POST("/containers/:id/restart", dockerHandler.RestartContainer)
			admin.GET("/containers/:id/logs", dockerHandler.GetContainerLogs)
			admin.GET("/containers/:id/stats", dockerHandler.GetContainerStats)
			admin.GET("/images", dockerHandler.ListImages)
			admin.GET("/images/:id", dockerHandler.InspectImage)
			admin.POST("/images/pull", dockerHandler.PullImage)
			admin.DELETE("/images/:id", dockerHandler.RemoveImage)
		}
	}

	// Start server
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("CloudForge API starting on :%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
