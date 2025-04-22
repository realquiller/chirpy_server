package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	secret := "testsecret"
	userID := uuid.New()
	expires := time.Minute

	token, err := MakeJWT(userID, secret, expires)
	if err != nil {
		t.Fatalf("error creating JWT: %v", err)
	}

	parsedUserID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("error validating JWT: %v", err)
	}

	if parsedUserID != userID {
		t.Errorf("expected userID %v, got %v", userID, parsedUserID)
	}
}

func TestExpiredJWT(t *testing.T) {
	secret := "testsecret"
	userID := uuid.New()

	token, err := MakeJWT(userID, secret, -time.Minute) // already expired
	if err != nil {
		t.Fatalf("error creating JWT: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Errorf("expected error for expired token, got none")
	}
}

func TestJWTWithWrongSecret(t *testing.T) {
	secret := "testsecret"
	wrongSecret := "wrongsecret"
	userID := uuid.New()
	expires := time.Minute

	token, err := MakeJWT(userID, secret, expires)
	if err != nil {
		t.Fatalf("error creating JWT: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Errorf("expected error with wrong secret, got none")
	}
}
