package services

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestGenerateAndValidateToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "cloudforge_test_secret_12345678901234567890")

	token, err := GenerateToken(1, "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	claims, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.Subject != "1" {
		t.Fatalf("Expected user ID 1, got %s", claims.Subject)
	}

	if claims.Role != "admin" {
		t.Fatalf("Expected role admin, got %s", claims.Role)
	}
}

func TestInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "cloudforge_test_secret_12345678901234567890")

	_, err := ValidateToken("invalid.token.value")
	if err == nil {
		t.Fatal("Invalid token should be rejected")
	}
}

func TestExpiredToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "cloudforge_test_secret_12345678901234567890")

	claims := Claims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "cloudforge",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("cloudforge_test_secret_12345678901234567890"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Expired token should be rejected")
	}
}

func TestFutureIssuedAtToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "cloudforge_test_secret_12345678901234567890")

	claims := Claims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "1",
			Issuer:    "cloudforge",
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(time.Hour)),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(2 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("cloudforge_test_secret_12345678901234567890"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Token issued in the future should be rejected")
	}
}

func TestInvalidSubjectToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "cloudforge_test_secret_12345678901234567890")

	claims := Claims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "invalid-user-id",
			Issuer:    "cloudforge",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("cloudforge_test_secret_12345678901234567890"))
	if err != nil {
		t.Fatal(err)
	}

	_, err = ValidateToken(tokenString)
	if err == nil {
		t.Fatal("Token with invalid subject should be rejected")
	}
}
