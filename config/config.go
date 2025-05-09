package config

import (
	"os"
)

type Config struct {
	MongoURI            string
	PostgresURI         string
	MongoDBName         string
	Environment         string
	JWTSecret           string
	UserServiceURL      string
	PostServiceURL      string
	CommunityServiceURL string
}

func LoadConfig() Config {
	return Config{
		MongoURI:            getEnvOrDefault("MONGO_URI", "mongodb://localhost:27017"),
		PostgresURI:         getEnvOrDefault("POSTGRES_URI", "postgres://postgres:postgres@localhost:5432/circle_connect?sslmode=disable"),
		MongoDBName:         getEnvOrDefault("MONGO_DB_NAME", "circle_connect_feeds"),
		Environment:         getEnvOrDefault("ENVIRONMENT", "development"),
		JWTSecret:           getEnvOrDefault("JWT_SECRET", "your-secret-key"),
		UserServiceURL:      getEnvOrDefault("USER_SERVICE_URL", "http://localhost:4001/api"),
		PostServiceURL:      getEnvOrDefault("POST_SERVICE_URL", "http://localhost:4000/api"),
		CommunityServiceURL: getEnvOrDefault("COMMUNITY_SERVICE_URL", "http://localhost:4002/api"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
