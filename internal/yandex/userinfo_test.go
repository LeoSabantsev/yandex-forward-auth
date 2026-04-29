package yandex

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClientUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodGet, r.Method)
		require.Equal(t, "OAuth access-token", r.Header.Get("Authorization"))
		require.Equal(t, "json", r.URL.Query().Get("format"))

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(UserInfo{
			ID:           "123456789",
			Login:        "alice",
			ClientID:     "client-id",
			DefaultEmail: "alice@example.com",
			Emails:       []string{"other@example.com"},
		}))
	}))
	defer server.Close()

	client := Client{UserInfoEndpoint: server.URL}

	info, err := client.UserInfo(context.Background(), "access-token")
	require.NoError(t, err)

	require.Equal(t, "123456789", info.ID)
	require.Equal(t, "alice", info.Login)
	require.Equal(t, "client-id", info.ClientID)
	require.Equal(t, "alice@example.com", info.Email())
}

func TestUserInfoEmailFallsBackToFirstEmail(t *testing.T) {
	info := UserInfo{Emails: []string{"first@example.com"}}

	require.Equal(t, "first@example.com", info.Email())
}

func TestClientUserInfoReturnsErrorForMissingToken(t *testing.T) {
	client := Client{}

	_, err := client.UserInfo(context.Background(), "")

	require.ErrorIs(t, err, ErrUserInfoFailed)
}

func TestClientUserInfoReturnsErrorForNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"invalid_token"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	client := Client{UserInfoEndpoint: server.URL}

	_, err := client.UserInfo(context.Background(), "access-token")

	require.ErrorIs(t, err, ErrUserInfoFailed)
}

func TestClientUserInfoReturnsErrorForMissingRequiredFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewEncoder(w).Encode(UserInfo{
			Login:    "alice",
			ClientID: "client-id",
		}))
	}))
	defer server.Close()

	client := Client{UserInfoEndpoint: server.URL}

	_, err := client.UserInfo(context.Background(), "access-token")

	require.ErrorIs(t, err, ErrUserInfoFailed)
}
