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
