package allowlist

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListAllowsUserID(t *testing.T) {
	list := List{UserIDs: []string{"123456789"}}

	require.True(t, list.Allows(Subject{UserID: "123456789"}))
}

func TestListAllowsEmailCaseInsensitively(t *testing.T) {
	list := List{Emails: []string{"alice@example.com"}}

	require.True(t, list.Allows(Subject{Email: "Alice@Example.com"}))
}

func TestListAllowsLoginCaseInsensitively(t *testing.T) {
	list := List{Logins: []string{"alice"}}

	require.True(t, list.Allows(Subject{Login: "Alice"}))
}

func TestListAllowsEmailDomain(t *testing.T) {
	list := List{EmailDomains: []string{"example.com"}}

	require.True(t, list.Allows(Subject{Email: "alice@example.com"}))
}

func TestListDeniesWhenNotConfigured(t *testing.T) {
	var list List

	require.False(t, list.Allows(Subject{
		UserID: "123456789",
		Email:  "alice@example.com",
		Login:  "alice",
	}))
}

func TestListDeniesUnknownSubject(t *testing.T) {
	list := List{UserIDs: []string{"123456789"}}

	require.False(t, list.Allows(Subject{UserID: "987654321"}))
}

func TestListAllowsAllOnlyWhenExplicitlyEnabled(t *testing.T) {
	list := List{DevAllowAll: true}

	require.True(t, list.Allows(Subject{}))
}
