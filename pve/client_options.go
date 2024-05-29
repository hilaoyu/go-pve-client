package pve

import (
	"fmt"
	"github.com/hilaoyu/go-utils/utilLogger"
	"net/http"
)

type Option func(*Client)

func WithHttpClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

func WithAuthAccount(username, password string) Option {
	return func(c *Client) {
		c.credentials = &Credentials{
			Username: username,
			Password: password,
		}
	}
}

func WithAuthApiToken(tokenID, secret string) Option {
	return func(c *Client) {
		c.token = fmt.Sprintf("%s=%s", tokenID, secret)
	}
}

func WithUserAgent(ua string) Option {
	return func(c *Client) {
		c.userAgent = ua
	}
}

func WithLogger(logger *utilLogger.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}
