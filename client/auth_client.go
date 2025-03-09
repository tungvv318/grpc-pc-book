package client

import (
	"pcbook/pb"

	"google.golang.org/grpc"
)

type AuthClient struct {
	service pb.AuthServiceClient,
	username string,
	password string,
}

func NewAuthClient(cc *grpc.ClientConn, username string, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

func (c *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	req := &pb.LoginRequest{
		Username:   c.username,
		Password: c.password,
	}
	resp, err := c.service.Login(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.GetAccessToken(), nil
}
