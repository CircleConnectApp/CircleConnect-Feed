package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/CircleConnectApp/feed-service/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("AuthMiddleware: Checking Authorization header...")
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			log.Println("AuthMiddleware: Missing Authorization header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			log.Println("AuthMiddleware: Invalid Authorization header format")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header format must be Bearer {token}"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		log.Println("AuthMiddleware: Token extracted:", tokenString)

		cfg := config.LoadConfig()

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				log.Println("AuthMiddleware: Invalid signing method")
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil {
			log.Printf("AuthMiddleware: Error parsing token: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			log.Println("AuthMiddleware: Token is valid")
			if userID, exists := claims["user_id"]; exists {
				log.Println("AuthMiddleware: User ID found:", userID)
				c.Set("user_id", int(userID.(float64)))
			} else {
				log.Println("AuthMiddleware: Missing user_id in token")
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token: missing user_id"})
				c.Abort()
				return
			}

			if role, exists := claims["role"]; exists {
				log.Println("AuthMiddleware: Role found:", role)
				c.Set("role", role.(string))
			}

			c.Next()
		} else {
			log.Println("AuthMiddleware: Invalid token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
	}
}
