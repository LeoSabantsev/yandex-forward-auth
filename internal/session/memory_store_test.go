package session

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemoryStorePutAndGet(t *testing.T) {
	store := NewMemoryStore()

	sessionID, err := Generate()
	require.NoError(t, err)

	require.NoError(t, store.Put(context.Background(), sessionID, &Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}))

	record, err := store.Get(context.Background(), sessionID)
	require.NoError(t, err)

	require.NotEmpty(t, record.SessionIDHash)
	require.NotEqual(t, sessionID, record.SessionIDHash)
	require.Equal(t, "123456789", record.UserID)
	require.Equal(t, "alice", record.Login)
	require.Equal(t, "alice@example.com", record.Email)
}

func TestMemoryStoreGetMissing(t *testing.T) {
	store := NewMemoryStore()

	sessionID, err := Generate()
	require.NoError(t, err)

	_, err = store.Get(context.Background(), sessionID)
	require.ErrorIs(t, err, ErrNotFound)
}

func TestMemoryStoreRevoke(t *testing.T) {
	store := NewMemoryStore()

	sessionID, err := Generate()
	require.NoError(t, err)

	require.NoError(t, store.Put(context.Background(), sessionID, &Record{
		UserID:    "123456789",
		Login:     "alice",
		Email:     "alice@example.com",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(time.Hour),
	}))

	require.NoError(t, store.Revoke(context.Background(), sessionID, "test"))

	record, err := store.Get(context.Background(), sessionID)
	require.NoError(t, err)
	require.True(t, record.Revoked())
	require.Equal(t, "test", record.RevokedReason)
}
