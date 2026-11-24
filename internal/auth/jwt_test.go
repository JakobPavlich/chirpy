package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

var tokenSecrets = []string{
	"nsR9djhxuL2jIAWFxDT7wA0rTxWs4bTA+GVPhKU+FYI=",
	"uHtxHcswePXcMdrt6S+Ce9zO+XAU6E4aVwBxjN+lGnY=",
	"3WKycBUkUk7jcM7cEmuas+AUAKem4S8WZF81/vi7mj0=",
}

var uuids = []uuid.UUID{
	uuid.MustParse("11111111-1111-4111-8111-111111111111"),
	uuid.MustParse("22222222-2222-4222-8222-222222222222"),
	uuid.MustParse("33333333-3333-4333-8333-333333333333"),
}

func TestJWT(t *testing.T) {
	for _, tokenSecret := range tokenSecrets {
		inputID := uuids[0]
		token, err := MakeJWT(inputID, tokenSecret, 1*time.Hour)
		if err != nil {
			t.Errorf("error making JWT: %v", err)
		}

		id, err := ValidateJWT(token, tokenSecret)
		if err != nil {
			t.Errorf("error validating JWT: %v", err)
		}
		if inputID != id {
			t.Errorf("id's don't match: %v != %v", inputID, id)
		}
	}
}

func TestExpired(t *testing.T) {
	inputID := uuids[1]
	tokenSecret := tokenSecrets[1]
	token, err := MakeJWT(inputID, tokenSecret, 1*time.Millisecond)
	if err != nil {
		t.Errorf("error making JWT: %v", err)
	}
	time.Sleep(10 * time.Millisecond)
	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Errorf("JWT should have been expired")
	}
}

func TestWrongSecret(t *testing.T) {
	inputID := uuids[2]
	tokenSecret := tokenSecrets[2]
	wrongTokenSecret := tokenSecrets[0]
	token, err := MakeJWT(inputID, tokenSecret, 1*time.Hour)
	if err != nil {
		t.Errorf("error making JWT: %v", err)
	}

	_, err = ValidateJWT(token, wrongTokenSecret)
	if err == nil {
		t.Errorf("token should not have been valid (wrong token secret test input given)")
	}
}
