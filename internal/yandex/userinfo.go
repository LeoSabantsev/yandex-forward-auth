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

const UserInfoEndpoint = "https://login.yandex.ru/info"

var ErrUserInfoFailed = errors.New("yandex user info request failed")

type UserInfo struct {
	ID           string   `json:"id"`
	Login        string   `json:"login"`
	ClientID     string   `json:"client_id"`
	DefaultEmail string   `json:"default_email"`
	Emails       []string `json:"emails"`
}

func (u UserInfo) Email() string {
	if u.DefaultEmail != "" {
		return u.DefaultEmail
	}

	if len(u.Emails) > 0 {
		return u.Emails[0]
	}

	return ""
}

func (c Client) UserInfo(ctx context.Context, accessToken string) (*UserInfo, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, ErrUserInfoFailed
	}

	endpoint := c.UserInfoEndpoint
	if endpoint == "" {
		endpoint = UserInfoEndpoint
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	query := parsed.Query()
	query.Set("format", "json")
	parsed.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "OAuth "+accessToken)

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
		return nil, ErrUserInfoFailed
	}

	var info UserInfo
	if err := json.NewDecoder(io.LimitReader(res.Body, 1<<20)).Decode(&info); err != nil {
		return nil, err
	}

	if info.ID == "" || info.Login == "" || info.ClientID == "" {
		return nil, ErrUserInfoFailed
	}

	return &info, nil
}
