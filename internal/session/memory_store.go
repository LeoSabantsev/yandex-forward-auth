package session

import (
	"context"
	"sync"
	"time"
)

//TODO: think about pointer on Record

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		records: make(map[string]Record),
	}
}

func (s *MemoryStore) Get(ctx context.Context, sessionID string) (*Record, error) {
	_ = ctx

	sessionIDHash, err := Hash(sessionID)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[sessionIDHash]
	if !ok {
		return nil, ErrNotFound
	}

	return &record, nil
}

func (s *MemoryStore) Put(ctx context.Context, sessionID string, record Record) error {
	_ = ctx

	sessionIDHash, err := Hash(sessionID)
	if err != nil {
		return err
	}

	record.SessionIDHash = sessionIDHash

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[sessionIDHash] = record
	return nil
}

func (s *MemoryStore) Revoke(ctx context.Context, sessionID string, reason string) error {
	_ = ctx

	sessionIDHash, err := Hash(sessionID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[sessionIDHash]
	if !ok {
		return ErrNotFound
	}

	now := time.Now().UTC()
	record.RevokedAt = &now
	record.RevokedReason = reason

	s.records[sessionIDHash] = record
	return nil
}
