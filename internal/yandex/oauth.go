package yandex

import (
	"net/url"
)

const (
	AuthorizeEndpoint = "https://oauth.yandex.com/authorize"
	PKCEMethodS256    = "S256"
)

type AuthCodeURLParams struct {
	ClientID      string
	RedirectURI   string
	State         string
	CodeChallenge string
}

func AuthCodeURL(params AuthCodeURLParams) string {
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", params.ClientID)
	values.Set("redirect_uri", params.RedirectURI)
	values.Set("state", params.State)
	values.Set("code_challenge", params.CodeChallenge)
	values.Set("code_challenge_method", PKCEMethodS256)

	return AuthorizeEndpoint + "?" + values.Encode()
}
