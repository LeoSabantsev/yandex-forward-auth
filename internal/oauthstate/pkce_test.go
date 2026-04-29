package oauthstate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateCodeVerifier(t *testing.T) {
	verifier, err := GenerateCodeVerifier()
	require.NoError(t, err)

	require.Len(t, verifier, CodeVerifierMinLength)
}

func TestCodeChallengeS256(t *testing.T) {
	challenge := CodeChallengeS256("test-verifier")

	require.Equal(t, "JBbiqONGWPaAmwXk_8bT6UnlPfrn65D32eZlJS-zGG0", challenge)
}
