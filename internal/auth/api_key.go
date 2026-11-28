package auth

import (
	"errors"
	"net/http"
)

func GetAPIKey(headers http.Header) (string, error) {
	authorization := headers.Get("Authorization")

	if authorization == "" {
		return "", errors.New("authorization header missing")
	}
	if authorization[:7] != "ApiKey " {
		return "", errors.New("token must start with 'ApiKey '")
	}
	if len(authorization) < 7 {
		return "", errors.New("invalid input(too short for a token)")
	}

	token := authorization[7:]
	return token, nil
}
