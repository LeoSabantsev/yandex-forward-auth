package session

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrNotFound = errors.New("session not found")

type Store interface {
	Get(ctx context.Context, sessionID string) (*Record, error)
	Put(ctx context.Context, sessionID string, record Record) error
	Revoke(ctx context.Context, sessionID string, reason string) error
}

func Hash(sessionID string) (string, error) {
	if err := Validate(sessionID); err != nil {
		return "", err
	}

	sum := sha256.Sum256([]byte(sessionID))
	return hex.EncodeToString(sum[:]), nil
}
