package api

import "context"

type IpResponse struct {
	Status        string `json:"status"`
	YourIp        string `json:"yourIp"`
	XForwardedFor string `json:"xForwardedFor"`
}

type UtilService interface {
	Ip(context.Context) (IpResponse, error)
}

type utilService struct {
	client *Client
}

func (s *utilService) Ip(ctx context.Context) (IpResponse, error) {
	var resp IpResponse
	err := s.client.PostJson(ctx, "/ip", EmptyPayload, &resp)
	return resp, err
}

type PingResponse struct {
	Status string `json:"status"`
	YourIp string `json:"yourIp"`
}

// Ping calls the authenticated /ping endpoint, which can be used to verify the API key and secret key.
func (c *Client) Ping(ctx context.Context) (PingResponse, error) {
	var resp PingResponse
	err := c.PostJson(ctx, "/ping", EmptyPayload, &resp)
	return resp, err
}
