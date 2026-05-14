package api

import (
	"net/http"
	"os"
)

// ClientOption is a functional option for configuring the Client
type ClientOption func(*Client)

// NewClient creates a new API client with the provided options
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		http:    http.DefaultClient,
		baseUrl: "https://api.porkbun.com/api/json/v3",
	}

	for _, opt := range opts {
		opt(c)
	}

	c.Dns = &dnsService{client: c}
	c.Util = &utilService{client: c}

	return c
}

// WithCredentials sets the API key and secret key for API authentication
func WithCredentials(apiKey, secretKey string) ClientOption {
	return func(c *Client) {
		c.apiKey = apiKey
		c.secretKey = secretKey
	}
}

// WithEnvironmentCredentials sets the API key and secret key for API authentication and reads their values from
// environment variables: PORKBUN_API_KEY and PORKBUN_SECRET_KEY
func WithEnvironmentCredentials() ClientOption {
	return WithCredentials(
		os.Getenv("PORKBUN_API_KEY"),
		os.Getenv("PORKBUN_SECRET_KEY"),
	)
}
