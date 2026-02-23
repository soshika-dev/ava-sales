package auth

import (
	"testing"
	"time"
)

func TestGenerateAndParseAccessToken_Success(t *testing.T) {
	secret := "test-secret"
	token, _, err := GenerateAccessToken(secret, "11111111-1111-1111-1111-111111111111", "CUSTOMER", 15*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}

	claims, err := ParseAccessToken(secret, token)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}
	if claims.Subject != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected subject: %s", claims.Subject)
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	secret := "test-secret"
	token, _, err := GenerateAccessToken(secret, "11111111-1111-1111-1111-111111111111", "CUSTOMER", -1*time.Minute)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}

	_, err = ParseAccessToken(secret, token)
	if err == nil {
		t.Fatal("expected expired token parse to fail")
	}
}
