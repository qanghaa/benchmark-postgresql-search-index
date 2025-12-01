package main

import (
	"context"
	"log"
	"net/http"

	"log-project/config"
	"log-project/database"
	"log-project/handlers"
	"log-project/middleware"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Log Project API
// @version 1.0
// @description API for managing logs with full-text search and performance testing
// @host localhost:8080
// @BasePath /api
func main() {
	ctx := context.Background()

	// Load configuration
	cfg := config.Load()

	// Initialize database connection for migrations
	log.Println("Initializing database connection for migrations...")
	sqlDB, err := database.Initialize(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize database for migrations:", err)
	}
	log.Println("Database connection for migrations initialized.")
	defer sqlDB.Close() // Close the migration database connection when main exits

	// Initialize pgx pool for application use
	log.Println("Debug: Before pgx pool initialization.")
	log.Println("Initializing pgx pool...")
	pool, err := database.InitializePool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to initialize connection pool:", err)
	}
	log.Println("pgx pool initialized.")
	log.Println("Debug: After pgx pool initialization.")
	defer pool.Close()

	log.Println("Database initialized successfully with pgx pool")

	// Initialize handlers
	h := handlers.New(pool)

	// Setup Gin router
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/*")

	// API routes
	api := r.Group("/api")
	{
		api.POST("/initialize", h.InitializeData)
		api.GET("/logs", h.GetLogs)
		api.GET("/search/partial", h.SearchLogsPartial)
		api.DELETE("/truncate", h.TruncateDatabase)

		// Example API with logging middleware
		api.POST("/example/:id", middleware.RequestLogger(pool), h.ExampleAPI)
	}

	// Web interface
	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"title": "Log Performance Testing",
		})
	})

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Start server
	log.Printf("Server starting on port %s", cfg.Port)
	log.Printf("Swagger documentation available at http://localhost:%s/swagger/index.html", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, r))
}
