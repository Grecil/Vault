package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"filevault-backend/internal/config"
	"filevault-backend/internal/database"
	"filevault-backend/internal/handlers"
	"filevault-backend/internal/middleware"
	"filevault-backend/internal/services"
	"filevault-backend/internal/storage"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	gin.SetMode(cfg.GinMode)

	// Initialize Clerk SDK
	middleware.InitializeClerk(cfg)

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize storage
	minioStorage, err := storage.NewMinIOStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO storage: %v", err)
	}

	// Initialize rate limiting service
	rateLimitService := services.NewRateLimitService(cfg)
	defer rateLimitService.Close()

	// Initialize services
	userService := services.NewUserService(db.DB, cfg)
	fileService := services.NewFileService(db.DB, minioStorage)

	// Initialize handlers
	userHandler := handlers.NewUserHandler(userService)
	fileHandler := handlers.NewFileHandler(fileService, userService)
	adminHandler := handlers.NewAdminHandler(userService, fileService)

	// Setup router
	router := gin.New()
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		healthStatus := gin.H{
			"status":     "healthy",
			"timestamp":  time.Now().UTC(),
			"database":   "connected",
			"storage":    "connected",
			"rate_limit": "enabled",
		}

		if !cfg.RateLimitEnabled {
			healthStatus["rate_limit"] = "disabled"
		}

		c.JSON(http.StatusOK, healthStatus)
	})

	// Share routes (clean URLs for sharing - at root level)
	router.GET("/share/:id", fileHandler.ShareFileDownload)

	// API routes
	api := router.Group("/api/v1")
	{
		api.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "FileVault API v1.0.0",
				"status":  "running",
			})
		})

		// Public routes (no auth required, but rate limited)
		public := api.Group("/public")
		public.Use(middleware.RateLimit(rateLimitService))
		{
			public.GET("/files/:id", fileHandler.GetPublicFile)
			public.GET("/files/:id/download", fileHandler.DownloadPublicFile)
		}

		// Protected routes (auth required)
		protected := api.Group("/")
		protected.Use(middleware.RequireAuth(cfg))
		protected.Use(middleware.RateLimit(rateLimitService))
		{
			// User routes
			user := protected.Group("/user")
			{
				user.GET("/profile", userHandler.GetProfile)
				user.GET("/storage", userHandler.GetStorageInfo)
				user.GET("/storage/statistics", userHandler.GetStorageStatistics)
			}

			// File routes
			files := protected.Group("/files")
			{
				files.POST("/upload-url", fileHandler.GenerateUploadURL)
				files.POST("/complete", fileHandler.CompleteUpload)
				files.POST("/batch/prepare", fileHandler.BatchPrepareUpload)
				files.POST("/batch/complete", fileHandler.BatchCompleteUpload)
				files.GET("", fileHandler.ListFiles)
				files.GET("/:id/download", fileHandler.DownloadFile)
				files.GET("/:id/share-link", fileHandler.GetShareLink)
				files.DELETE("/:id", fileHandler.DeleteFile)
				files.PATCH("/:id/public", fileHandler.TogglePublic)
			}
		}

		// Admin routes (admin auth required)
		admin := api.Group("/admin")
		admin.Use(middleware.RequireAuth(cfg))
		admin.Use(middleware.RequireAdmin())
		admin.Use(middleware.RateLimit(rateLimitService))
		{
			admin.GET("/users", adminHandler.ListUsers)
			admin.DELETE("/users/:id", adminHandler.DeleteUser)
			admin.PATCH("/users/:id/role", adminHandler.UpdateUserRole)
			admin.PATCH("/users/:id/quota", adminHandler.UpdateUserQuota)
			admin.GET("/stats", adminHandler.GetStats)
		}
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ServerPort),
		Handler: router,
	}

	go func() {
		log.Printf("🚀 Server starting on http://localhost:%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
