package oauthstate

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryStorePutAndGet(t *testing.T) {
	store := NewMemoryStore()

	require.NoError(t, store.Put(context.Background(), "state-id", Record{
		Nonce:        "nonce",
		CodeVerifier: "verifier",
		ReturnURL:    "https://app.example.com/",
		CreatedAt:    time.Now().UTC(),
		ExpiresAt:    time.Now().UTC().Add(5 * time.Minute),
	}))

	record, err := store.Get(context.Background(), "state-id")
	require.NoError(t, err)

	require.NotEmpty(t, record.StateIDHash)
	require.NotEqual(t, "state-id", record.StateIDHash)
	require.Equal(t, "nonce", record.Nonce)
	require.Equal(t, "verifier", record.CodeVerifier)
	require.Equal(t, "https://app.example.com/", record.ReturnURL)
}

func TestMemoryStoreGetMissing(t *testing.T) {
	store := NewMemoryStore()

	_, err := store.Get(context.Background(), "missing")

	require.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStoreMarkUsed(t *testing.T) {
	store := NewMemoryStore()

	require.NoError(t, store.Put(context.Background(), "state-id", Record{
		Nonce:     "nonce",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(5 * time.Minute),
	}))

	require.NoError(t, store.MarkUsed(context.Background(), "state-id"))

	record, err := store.Get(context.Background(), "state-id")
	require.NoError(t, err)
	require.True(t, record.Used())
}

func TestRecordExpired(t *testing.T) {
	record := Record{ExpiresAt: time.Now().UTC().Add(-time.Minute)}

	require.True(t, record.Expired(time.Now().UTC()))
}
