package routes

import (
	"database/sql"
	"log"

	"github.com/CircleConnectApp/feed-service/config"
	"github.com/CircleConnectApp/feed-service/controllers"
	"github.com/CircleConnectApp/feed-service/middleware"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRoutes(r *gin.Engine, db *mongo.Database, pgDB *sql.DB) {
	log.Println("Setting up routes...")

	cfg := config.LoadConfig()
	feedController := controllers.NewFeedController(
		db,
		pgDB,
		cfg.UserServiceURL,
		cfg.PostServiceURL,
		cfg.CommunityServiceURL,
	)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "feed-service",
		})
	})

	api := r.Group("/api")
	log.Println("Registering /api routes...")

	auth := api.Group("/")
	auth.Use(middleware.AuthMiddleware())
	{
		auth.GET("/feed", func(c *gin.Context) {
			log.Println("GET /api/feed route hit")
			feedController.GetFeed(c)
		})

		auth.GET("/feed/recommended", func(c *gin.Context) {
			log.Println("GET /api/feed/recommended route hit")
			feedController.GetRecommendedPosts(c)
		})

		auth.GET("/feed/preferences", func(c *gin.Context) {
			log.Println("GET /api/feed/preferences route hit")
			feedController.GetUserPreferences(c)
		})

		auth.PUT("/feed/preferences", func(c *gin.Context) {
			log.Println("PUT /api/feed/preferences route hit")
			feedController.UpdateUserPreferences(c)
		})
	}

	log.Println("Routes registered successfully.")
}
