package jwtutils

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const SECRET_KEY = "temp"

var (
	INVALID_TOKEN error = errors.New("invalid token")
)

func CreateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(time.Minute * 20).Unix(),
	})
	return token.SignedString([]byte(SECRET_KEY))
}

func VerifyJWT(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (any, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse jwt: %v", err)
	}

	if !token.Valid {
		return nil, INVALID_TOKEN
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, INVALID_TOKEN
	}

	return claims, nil
}
