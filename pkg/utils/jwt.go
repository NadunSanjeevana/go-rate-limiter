package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// Secret key for JWT signing
	jwtSecret = []byte("your-secret-key")
)

// SetJWTSecret sets the JWT secret key
func SetJWTSecret(secret []byte) {
	jwtSecret = secret
}

// GenerateJWT creates a new JWT token
func GenerateJWT(claims map[string]interface{}, expiresIn time.Duration) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	
	// Set claims
	tokenClaims := token.Claims.(jwt.MapClaims)
	for key, value := range claims {
		tokenClaims[key] = value
	}
	
	// Set expiration if provided
	if expiresIn > 0 {
		tokenClaims["exp"] = time.Now().Add(expiresIn).Unix()
	}
	
	// Sign and return token
	return token.SignedString(jwtSecret)
}

// ParseJWT validates and parses a JWT token
func ParseJWT(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return claims, nil
}