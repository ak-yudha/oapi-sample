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

	"github.com/auliayudha/oapi-sample/internal/api"
	"github.com/auliayudha/oapi-sample/internal/db"
	"github.com/auliayudha/oapi-sample/internal/gen"
	"github.com/auliayudha/oapi-sample/internal/repository"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	// Database configuration
	dbConfig := db.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "auliayudha"),
		Password: getEnv("DB_PASSWORD", ""),
		DBName:   getEnv("DB_NAME", "openapi"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Initialize database connection
	database, err := db.NewDB(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Run migrations
	err = db.RunMigrations(database, "migrations")
	if err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Create repository
	userRepo := repository.NewUserRepository(database)

	// Create handler
	userHandler := api.NewUserHandler(userRepo)

	// Set up Echo
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register API routes using the generated wrapper
	gen.RegisterHandlers(e, userHandler)

	// Get server port
	port := getEnv("PORT", "8080")
	serverAddr := fmt.Sprintf(":%s", port)

	// Start server
	server := &http.Server{
		Addr:    serverAddr,
		Handler: e,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server is running on http://localhost%s", serverAddr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

// getEnv returns environment variable or fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
