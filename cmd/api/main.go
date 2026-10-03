package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/khoirirrosikin/task-management/internal/auth"
	"github.com/khoirirrosikin/task-management/internal/database"
	"github.com/khoirirrosikin/task-management/internal/database/db"
	"github.com/khoirirrosikin/task-management/internal/middleware"
	"github.com/khoirirrosikin/task-management/internal/project"
	"github.com/khoirirrosikin/task-management/internal/task"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	dbURL := getEnv("DATABASE_URL")
	jwtSecret := getEnv("JWT_SECRET")
	port := getEnv("PORT")

	ctx := context.Background()

	pool, err := database.NewPostgresPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()
	log.Println("PostgresSQL successfully connected")

	queries := db.New(pool)

	authRepo := auth.NewRepository(queries)
	authService := auth.NewService(authRepo, jwtSecret)
	authHandler := auth.NewHandler(authService)

	projectRepo := project.NewRepository(queries)
	projectService := project.NewService(projectRepo)
	projectHandler := project.NewHandler(projectService)

	taskRepo := task.NewRepository(queries)
	taskService := task.NewService(taskRepo, projectRepo)
	taskHandler := task.NewHandler(taskService)

	authMiddleware := middleware.AuthMiddleware(jwtSecret)

	router := gin.Default()

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "OK", "timestamp": time.Now()})
	})

	v1 := router.Group("/api/v1")
	authHandler.RegisterRoutes(v1)
	projectHandler.RegisterRoutes(v1, authMiddleware)
	taskHandler.RegisterRoutes(v1, authMiddleware)

	srv := &http.Server{
		Addr: ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("Server is running on post :%s", port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v\n", err)
		}
	}()
	
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited properly")
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists || value == "" {
		log.Fatalf("Fatal: environtment variable %s is required but not set", key)
	}
	return value
}