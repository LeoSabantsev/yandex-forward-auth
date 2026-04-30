package oauthstate

import "time"

type Record struct {
	StateIDHash  string
	Nonce        string
	CodeVerifier string
	ReturnURL    string
	CreatedAt    time.Time
	ExpiresAt    time.Time
	UsedAt       *time.Time
}

func (r Record) Expired(now time.Time) bool {
	if r.ExpiresAt.IsZero() {
		return true
	}

	return !now.Before(r.ExpiresAt)
}

func (r Record) Used() bool {
	return r.UsedAt != nil && !r.UsedAt.IsZero()
}
