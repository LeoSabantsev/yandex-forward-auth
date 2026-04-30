package allowlist

import (
	"strings"
)

type Subject struct {
	UserID string
	Login  string
	Email  string
}

type List struct {
	UserIDs      []string
	Emails       []string
	EmailDomains []string
	Logins       []string
	DevAllowAll  bool
}

func (l List) Allows(subject Subject) bool {
	if l.DevAllowAll {
		return true
	}

	if !l.configured() {
		return false
	}

	return containsExact(l.UserIDs, subject.UserID) ||
		containsNormalized(l.Emails, subject.Email) ||
		containsNormalized(l.Logins, subject.Login) ||
		containsExact(l.EmailDomains, emailDomain(subject.Email))
}

func (l List) configured() bool {
	return len(l.UserIDs) > 0 || len(l.Emails) > 0 || len(l.EmailDomains) > 0 || len(l.Logins) > 0
}

func containsExact(values []string, candidate string) bool {
	if candidate == "" {
		return false
	}

	expected := strings.TrimSpace(candidate)
	for _, value := range values {
		if strings.TrimSpace(value) == expected {
			return true
		}
	}

	return false
}

func containsNormalized(values []string, candidate string) bool {
	if candidate == "" {
		return false
	}

	expected := strings.ToLower(strings.TrimSpace(candidate))
	for _, value := range values {
		if strings.ToLower(strings.TrimSpace(value)) == expected {
			return true
		}
	}

	return false
}

func emailDomain(email string) string {
	before, after, ok := strings.Cut(email, "@")
	if !ok || before == "" || after == "" {
		return ""
	}

	return strings.TrimSuffix(after, ".")
}
