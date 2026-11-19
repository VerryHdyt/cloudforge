package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/VerryHdyt/cloudforge/backend/internal/database"
	"github.com/VerryHdyt/cloudforge/backend/internal/handlers"
	"github.com/VerryHdyt/cloudforge/backend/internal/middleware"
	"github.com/VerryHdyt/cloudforge/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := run(); err != nil {
		log.Printf("CloudForge API stopped with an error: %v", err)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("failed to load .env file: %w", err)
		}
		log.Println("Warning: .env file not found, using system environment variables")
	}

	port, err := apiPort(os.Getenv("APP_PORT"))
	if err != nil {
		return err
	}

	db, err := database.Connect()
	if err != nil {
		return fmt.Errorf("database connection failed: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer func() {
		if err := sqlDB.Close(); err != nil {
			log.Printf("Failed to close PostgreSQL connection: %v", err)
		}
	}()

	log.Println("PostgreSQL connected successfully")

	if err := database.Migrate(db); err != nil {
		return fmt.Errorf("database migration failed: %w", err)
	}

	log.Println("Database migrations completed successfully")

	auditService := &services.AuditService{
		DB: db,
	}

	dockerService, err := services.NewDockerService()
	if err != nil {
		return fmt.Errorf("failed to connect to Docker Engine: %w", err)
	}
	defer func() {
		if err := dockerService.Client.Close(); err != nil {
			log.Printf("Failed to close Docker client: %v", err)
		}
	}()

	log.Println("Docker Engine connected successfully")

	router := gin.Default()

	if err := router.SetTrustedProxies(nil); err != nil {
		return fmt.Errorf("failed to configure trusted proxies: %w", err)
	}

	router.Use(corsMiddleware())

	router.GET("/health", handlers.HealthCheck)

	api := router.Group("/api/v1")
	{
		api.GET("/info", handlers.PlatformInfo)

		authHandler := &handlers.AuthHandler{
			DB:    db,
			Audit: auditService,
		}

		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		userHandler := &handlers.UserHandler{
			DB:    db,
			Audit: auditService,
		}

		auditHandler := &handlers.AuditHandler{
			DB: db,
		}

		dockerHandler := &handlers.DockerHandler{
			Docker: dockerService,
			Audit:  auditService,
		}

		protected := api.Group("")
		protected.Use(middleware.AuthMiddleware(db))

		admin := protected.Group("")
		admin.Use(middleware.RequireRoles("admin"))
		{
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

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Printf("CloudForge API starting on :%s", port)

	serverErrors := make(chan error, 1)
	go func() {
		err := server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			serverErrors <- nil
			return
		}
		serverErrors <- err
	}()

	select {
	case err := <-serverErrors:
		if err != nil {
			return fmt.Errorf("API server failed: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Println("Shutdown signal received; stopping API server")
	}

	shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	shutdownErr := server.Shutdown(shutdownContext)
	if shutdownErr != nil {
		if err := server.Close(); err != nil {
			shutdownErr = errors.Join(shutdownErr, fmt.Errorf("failed to force close API server: %w", err))
		}
	}

	if err := <-serverErrors; err != nil {
		return fmt.Errorf("API server failed during shutdown: %w", err)
	}
	if shutdownErr != nil {
		return fmt.Errorf("failed to gracefully shut down API server: %w", shutdownErr)
	}

	log.Println("API server stopped")
	return nil
}

func apiPort(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "8080", nil
	}

	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return "", fmt.Errorf("APP_PORT must be an integer between 1 and 65535")
	}

	return strconv.Itoa(port), nil
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]struct{}{
		"http://localhost:5173": {},
		"http://127.0.0.1:5173": {},
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, allowed := allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Vary", "Origin")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
