package auth

import (
	"errors"
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/password"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type Auth struct {
	secretKey      []byte
	passwordHasher *password.Hasher
}

func NewAuth(secretKey string) *Auth {
	return &Auth{
		secretKey:      []byte(secretKey),
		passwordHasher: password.NewHasher(10),
	}
}

func (a *Auth) HashPassword(pwd string) (string, error) {
	return a.passwordHasher.HashPassword(pwd)
}

func (a *Auth) CheckPassword(pwd, hash string) bool {
	return a.passwordHasher.CheckPassword(pwd, hash)
}

func (a *Auth) GenerateToken(user *common.User) (string, error) {
	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

func (a *Auth) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return a.secretKey, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
