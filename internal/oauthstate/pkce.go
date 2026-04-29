package oauthstate

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const (
	CodeVerifierBytes     = 32
	CodeVerifierMinLength = 43
)

func GenerateCodeVerifier() (string, error) {
	buf := make([]byte, CodeVerifierBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func CodeChallengeS256(verifier string) string {
	sum := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}
