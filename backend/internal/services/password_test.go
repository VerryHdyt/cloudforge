package services

import (
	"strings"
	"testing"
)

func TestHashAndCheckPassword(t *testing.T) {
	password := "cloudforge_secure_123"

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashedPassword == password {
		t.Fatal("Password should not be stored in plain text")
	}

	if !CheckPassword(password, hashedPassword) {
		t.Fatal("Correct password should be accepted")
	}

	if CheckPassword("wrong_password", hashedPassword) {
		t.Fatal("Incorrect password should be rejected")
	}
}

func TestHashEmptyPassword(t *testing.T) {
	_, err := HashPassword("")

	if err == nil {
		t.Fatal("Empty password should be rejected")
	}
}

func TestHashPasswordExceeding72Bytes(t *testing.T) {
	password := strings.Repeat("a", 73)

	_, err := HashPassword(password)

	if err == nil {
		t.Fatal("Password exceeding 72 bytes should be rejected")
	}
}

func TestHashPasswordExactly72Bytes(t *testing.T) {
	password := strings.Repeat("a", 72)

	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Password of exactly 72 bytes should be accepted: %v", err)
	}

	if !CheckPassword(password, hashedPassword) {
		t.Fatal("Password of exactly 72 bytes should be verified")
	}
}

func TestCheckInvalidPasswordHash(t *testing.T) {
	if CheckPassword("some_password", "invalid_hash") {
		t.Fatal("Invalid password hash should be rejected")
	}
}
