package session

import "time"

type Record struct {
	SessionIDHash string
	UserID        string
	Login         string
	Email         string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	LastSeenAt    time.Time

	RevokedAt     *time.Time
	RevokedReason string
}

func (r Record) Expired(now time.Time) bool {
	if r.ExpiresAt.IsZero() {
		return true
	}

	return !now.Before(r.ExpiresAt)
}

func (r Record) Revoked() bool {
	return r.RevokedAt != nil && !r.RevokedAt.IsZero()
}
