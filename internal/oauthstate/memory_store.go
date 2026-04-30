package oauthstate

import (
	"context"
	"sync"
	"time"
)

type MemoryStore struct {
	mu      sync.RWMutex
	records map[string]Record
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		records: make(map[string]Record),
	}
}

func (s *MemoryStore) Put(ctx context.Context, stateID string, record Record) error {
	_ = ctx

	copy := record
	copy.StateIDHash = Hash(stateID)

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records[copy.StateIDHash] = copy
	return nil
}

func (s *MemoryStore) Get(ctx context.Context, stateID string) (*Record, error) {
	_ = ctx

	stateIDHash := Hash(stateID)

	s.mu.RLock()
	defer s.mu.RUnlock()

	record, ok := s.records[stateIDHash]
	if !ok {
		return nil, ErrNotFound
	}

	return &record, nil
}

func (s *MemoryStore) MarkUsed(ctx context.Context, stateID string) error {
	_ = ctx

	stateIDHash := Hash(stateID)

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.records[stateIDHash]
	if !ok {
		return ErrNotFound
	}

	now := time.Now().UTC()
	record.UsedAt = &now

	s.records[stateIDHash] = record
	return nil
}
