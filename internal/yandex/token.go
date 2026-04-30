package yandex

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const TokenEndpoint = "https://oauth.yandex.com/token"

var ErrTokenExchangeFailed = errors.New("yandex token exchange failed")

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	HTTPClient       HTTPClient
	TokenEndpoint    string
	UserInfoEndpoint string
}

type TokenRequest struct {
	ClientID     string
	ClientSecret string
	Code         string
	CodeVerifier string
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

func (c Client) ExchangeCode(ctx context.Context, params TokenRequest) (*TokenResponse, error) {
	values := url.Values{}
	values.Set("grant_type", "authorization_code")
	values.Set("code", params.Code)
	values.Set("code_verifier", params.CodeVerifier)

	if params.ClientSecret == "" {
		values.Set("client_id", params.ClientID)
	}

	endpoint := c.TokenEndpoint
	if endpoint == "" {
		endpoint = TokenEndpoint
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	if params.ClientSecret != "" {
		req.SetBasicAuth(params.ClientID, params.ClientSecret)
	}

	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	res, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode < 200 || res.StatusCode > 299 {
		_, _ = io.Copy(io.Discard, io.LimitReader(res.Body, 1024))
		return nil, ErrTokenExchangeFailed
	}

	var token TokenResponse
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&token); err != nil {
		return nil, err
	}

	if token.AccessToken == "" || !strings.EqualFold(token.TokenType, "bearer") {
		return nil, ErrTokenExchangeFailed
	}

	return &token, nil
}
