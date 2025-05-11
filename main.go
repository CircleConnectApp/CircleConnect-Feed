package main

import (
	"context"
	"log"
	"os"

	"github.com/CircleConnectApp/feed-service/config"
	"github.com/CircleConnectApp/feed-service/database"
	"github.com/CircleConnectApp/feed-service/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	cfg := config.LoadConfig()

	mongoClient, err := database.ConnectMongoDB(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(context.Background())
	log.Println("Connected to MongoDB")

	pgDB, err := database.ConnectPostgres(cfg.PostgresURI)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer pgDB.Close()
	log.Println("Connected to PostgreSQL")

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	routes.SetupRoutes(router, mongoClient.Database(cfg.MongoDBName), pgDB)

	port := os.Getenv("PORT")
	if port == "" {
		port = "4004"
	}
	log.Printf("Feed service running on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
