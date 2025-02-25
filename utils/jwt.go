package utils

import (
	"time"

	"github.com/dgrijalva/jwt-go"
)

// Secret key (use environment variables in production)
var jwtSecret = []byte("supersecretkey")

// GenerateJWT generates a JWT token for a user
func GenerateJWT(username string, role string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"role":     role,
		"exp":      time.Now().Add(time.Hour * 1).Unix(), // Token expires in 1 hour
	})

	return token.SignedString(jwtSecret)
}

// ParseJWT parses and validates a JWT token
func ParseJWT(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, err
	}

	return claims, nil
}



