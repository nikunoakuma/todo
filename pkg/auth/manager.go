package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"os"
	"strconv"
	"time"
)

// TODO: добавить refresh token и эндпоинт, где пользователи могут обновлять access token

const alg = "HS256"

var (
	ErrSubEmpty    = errors.New("subject is empty")
	ErrSecretEmpty = errors.New("secret is empty")
)

type Manager struct {
	jwtSecret []byte
}

func NewManager() (*Manager, error) {
	const op = "auth.NewManager"
	jwtSecret := os.Getenv("JWT_SECRET")

	if jwtSecret == "" {
		return nil, fmt.Errorf("%s: %w", op, ErrSecretEmpty)
	}

	return &Manager{[]byte(jwtSecret)}, nil
}

func (m *Manager) GenerateAccessToken(userID int, tokenTTL time.Duration) (string, error) {
	const op = "auth.GenerateAccessToken"

	generatedJWT := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenTTL)),
		},
	)

	rawJWT, err := generatedJWT.SignedString(m.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return rawJWT, nil
}

func (m *Manager) ParseToken(receivedJWT string) (int, error) {
	const op = "auth.ParseToken"
	parsedJWT, err := jwt.Parse(
		receivedJWT,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, fmt.Errorf("unexpected alg type: %s", token.Header["alg"])
			}

			return m.jwtSecret, nil
		},
		jwt.WithValidMethods([]string{alg}),
	)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	sub, err := parsedJWT.Claims.GetSubject()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	if sub == "" {
		return 0, fmt.Errorf("%s: %w", op, ErrSubEmpty)
	}

	userID, err := strconv.Atoi(sub)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}
