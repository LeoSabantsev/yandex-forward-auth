package yandex

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientExchangeCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		username, password, ok := r.BasicAuth()
		require.True(t, ok)
		require.Equal(t, "client-id", username)
		require.Equal(t, "client-secret", password)

		require.NoError(t, r.ParseForm())
		require.Equal(t, "authorization_code", r.Form.Get("grant_type"))
		require.Equal(t, "auth-code", r.Form.Get("code"))
		require.Equal(t, "verifier", r.Form.Get("code_verifier"))
		require.Empty(t, r.Form.Get("client_secret"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(TokenResponse{
			AccessToken:  "access-token",
			TokenType:    "bearer",
			ExpiresIn:    3600,
			RefreshToken: "refresh-token",
			Scope:        "login:info login:email",
		}))
	}))
	defer server.Close()

	client := Client{TokenEndpoint: server.URL}

	token, err := client.ExchangeCode(context.Background(), TokenRequest{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Code:         "auth-code",
		CodeVerifier: "verifier",
	})
	require.NoError(t, err)

	require.Equal(t, "access-token", token.AccessToken)
	require.Equal(t, "bearer", token.TokenType)
	require.Equal(t, int64(3600), token.ExpiresIn)
	require.Equal(t, "refresh-token", token.RefreshToken)
	require.Equal(t, "login:info login:email", token.Scope)
}

func TestClientExchangeCodeSendsClientIDWhenSecretIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _, ok := r.BasicAuth()
		require.False(t, ok)

		require.NoError(t, r.ParseForm())
		require.Equal(t, "client-id", r.Form.Get("client_id"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(TokenResponse{
			AccessToken: "access-token",
			TokenType:   "bearer",
		}))
	}))
	defer server.Close()

	client := Client{TokenEndpoint: server.URL}

	_, err := client.ExchangeCode(context.Background(), TokenRequest{
		ClientID:     "client-id",
		Code:         "auth-code",
		CodeVerifier: "verifier",
	})
	require.NoError(t, err)
}

func TestClientExchangeCodeReturnsErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	client := Client{TokenEndpoint: server.URL}

	_, err := client.ExchangeCode(context.Background(), TokenRequest{
		ClientID:     "client-id",
		Code:         "auth-code",
		CodeVerifier: "verifier",
	})

	require.ErrorIs(t, err, ErrTokenExchangeFailed)
}

func TestClientExchangeCodeReturnsErrorWhenAccessTokenIsMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(TokenResponse{
			TokenType: "bearer",
		}))
	}))
	defer server.Close()

	client := Client{TokenEndpoint: server.URL}

	_, err := client.ExchangeCode(context.Background(), TokenRequest{
		ClientID:     "client-id",
		Code:         "auth-code",
		CodeVerifier: "verifier",
	})

	require.ErrorIs(t, err, ErrTokenExchangeFailed)
	require.False(t, errors.Is(err, context.Canceled))
}
