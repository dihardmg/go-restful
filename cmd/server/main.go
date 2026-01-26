package main

import (
	"log"

	"go-rest/internal/config"
	"go-rest/internal/database"
	"go-rest/internal/handlers"
	"go-rest/internal/repository"

	"github.com/gin-gonic/gin"

	_ "go-rest/docs" // Import swagger docs
)

// @title Go REST API
// @version 1.0
// @description A Product CRUD REST API built with Go, Gin, GORM, and PostgreSQL
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize repository
	productRepo := repository.NewProductRepository(db)

	// Initialize handlers
	productHandler := handlers.NewProductHandler(productRepo)

	// Setup Gin router
	router := gin.Default()

	// Serve custom Swagger UI with displayRequestDuration
	router.GET("/swagger", func(c *gin.Context) {
		c.File("./docs/index.html")
	})
	router.GET("/swagger/", func(c *gin.Context) {
		c.File("./docs/index.html")
	})
	router.GET("/swagger/index.html", func(c *gin.Context) {
		c.File("./docs/index.html")
	})

	// Swagger API docs endpoint
	router.GET("/swagger/doc.json", func(c *gin.Context) {
		c.File("./docs/swagger.json")
	})

	// Health check endpoint
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
			"status":  "healthy",
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Product routes
		products := v1.Group("/products")
		{
			products.POST("", productHandler.CreateProduct)
			products.GET("", productHandler.GetProducts)
			products.GET("/multiple", productHandler.GetMultipleProducts)
			products.POST("/bulk", productHandler.BulkCreate)
			products.GET("/:id", productHandler.GetProduct)
			products.PUT("/:id", productHandler.UpdateProduct)
			products.DELETE("/:id", productHandler.DeleteProduct)
		}
	}

	// Start server
	addr := ":" + cfg.ServerPort
	log.Printf("Server starting on %s", addr)
	log.Printf("Swagger UI available at http://localhost:%s/swagger/", cfg.ServerPort)
	log.Printf("API Documentation at http://localhost:%s/swagger/doc.json", cfg.ServerPort)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
