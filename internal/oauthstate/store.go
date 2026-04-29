package oauthstate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

var ErrNotFound = errors.New("oauth state not found")

type Store interface {
	Put(ctx context.Context, stateID string, record Record) error
	Get(ctx context.Context, stateID string) (*Record, error)
	MarkUsed(ctx context.Context, stateID string) error
}

func Hash(stateID string) string {
	sum := sha256.Sum256([]byte(stateID))
	return hex.EncodeToString(sum[:])
}
