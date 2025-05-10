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
		PostgresURI:         getEnvOrDefault("POSTGRES_URI", "postgresql://postgres:yehia@localhost:5432/postgres?sslmode=disable"),
		MongoDBName:         getEnvOrDefault("MONGO_DB_NAME", "circle_connect"),
		Environment:         getEnvOrDefault("ENVIRONMENT", "development"),
		JWTSecret:           getEnvOrDefault("JWT_SECRET", "a5898119500ed0f2fcb2f63a40f03ccbc35ce27ba86b862f427aa9c842ed44cb"),
		UserServiceURL:      getEnvOrDefault("USER_SERVICE_URL", "http://localhost:8081"),
		PostServiceURL:      getEnvOrDefault("POST_SERVICE_URL", "http://localhost:4000/api"),
		CommunityServiceURL: getEnvOrDefault("COMMUNITY_SERVICE_URL", "http://localhost:3000"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
