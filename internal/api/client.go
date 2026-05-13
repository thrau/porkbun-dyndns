package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

var EmptyPayload = struct{}{}

type Client struct {
	http      *http.Client
	baseUrl   string
	secretKey string
	apiKey    string

	Dns  DNSService
	Util UtilService
}

type ErrorResponse struct {
	HTTPStatus int    `json:"-"`
	Status     string `json:"status"`
	Message    string `json:"message"`
	Code       string `json:"code"`
}

func (e ErrorResponse) Error() string {
	return fmt.Sprintf(
		"porkbun api error (%d): status=%s code=%s message=%s",
		e.HTTPStatus,
		e.Status,
		e.Code,
		e.Message,
	)
}

func (c *Client) PostJson(ctx context.Context, path string, payload any, out any) error {
	if payload == nil {
		payload = struct{}{}
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseUrl+path,
		bytes.NewReader(body),
	)

	if err != nil {
		return err
	}

	// authentication
	if c.apiKey != "" {
		req.Header.Set("X-API-Key", c.apiKey)
	}
	if c.secretKey != "" {
		req.Header.Set("X-Secret-API-Key", c.secretKey)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var apiErr ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return err
		}
		apiErr.HTTPStatus = resp.StatusCode
		return apiErr
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
