package utils

import (
	"os"


	"github.com/golang-jwt/jwt/v5"
)


func VerifyToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	})
	if err != nil || !token.Valid {
		return "", err
	}

	claims := token.Claims.(jwt.MapClaims)
	userID := claims["id"].(string)

	return userID, nil
}