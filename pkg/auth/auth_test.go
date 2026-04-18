package auth

import (
	"github.com/ccy/devices-monitor/internal/common"
	"github.com/ccy/devices-monitor/pkg/password"
	"github.com/golang-jwt/jwt/v5"
	"testing"
	"time"
)

func TestGenerateToken(t *testing.T) {
	auth := NewAuth("test-secret-key")

	user := &common.User{
		ID:    "test-user-id",
		Email: "test@example.com",
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	if token == "" {
		t.Fatal("Generated token is empty")
	}

	t.Logf("Generated token: %s", token)
}

func TestValidateToken(t *testing.T) {
	auth := NewAuth("test-secret-key")

	user := &common.User{
		ID:    "test-user-id",
		Email: "test@example.com",
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != user.ID {
		t.Errorf("Expected user ID %s, got %s", user.ID, claims.UserID)
	}
}

func TestValidateInvalidToken(t *testing.T) {
	auth := NewAuth("test-secret-key")

	invalidToken := "invalid-token-string"
	_, err := auth.ValidateToken(invalidToken)

	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}

	if err != ErrInvalidToken {
		t.Errorf("Expected ErrInvalidToken, got %v", err)
	}
}

func TestExpiredToken(t *testing.T) {
	shortLivedAuth := &Auth{
		secretKey:      []byte("test-secret-key"),
		passwordHasher: password.NewHasher(10),
	}

	user := &common.User{
		ID:    "test-user-id",
		Email: "test@example.com",
	}

	claims := Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-25 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(shortLivedAuth.secretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	_, err = shortLivedAuth.ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expected error for expired token, got nil")
	}

	if err != ErrExpiredToken {
		t.Errorf("Expected ErrExpiredToken, got %v", err)
	}
}

func TestHashPassword(t *testing.T) {
	auth := NewAuth("test-secret-key")

	password := "test-password-123"
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashedPassword == "" {
		t.Fatal("Hashed password is empty")
	}

	if hashedPassword == password {
		t.Fatal("Hashed password should not be equal to plain password")
	}

	t.Logf("Hashed password: %s", hashedPassword)
}

func TestCheckPassword(t *testing.T) {
	auth := NewAuth("test-secret-key")

	password := "test-password-123"
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if !auth.CheckPassword(password, hashedPassword) {
		t.Fatal("Password check failed for correct password")
	}

	if auth.CheckPassword("wrong-password", hashedPassword) {
		t.Fatal("Password check should fail for wrong password")
	}
}
