package client

import (
	"context"
	"fmt"

	pbuser "github.com/klemanjar0/payment-system/generated/proto/user"
	usecase "github.com/klemanjar0/payment-system/services/transaction/internal/use_case"
	"google.golang.org/grpc"
)

type UserClient struct {
	client pbuser.UserServiceClient
}

func NewUserClient(conn *grpc.ClientConn) *UserClient {
	return &UserClient{client: pbuser.NewUserServiceClient(conn)}
}

func (c *UserClient) GetUser(ctx context.Context, userID string) (*usecase.UserInfo, error) {
	resp, err := c.client.GetUser(ctx, &pbuser.GetUserRequest{UserId: userID})
	if err != nil {
		return nil, fmt.Errorf("user service GetUser: %w", err)
	}

	userStatus := "inactive"
	if resp.GetStatus() == pbuser.UserStatus_USER_STATUS_ACTIVE {
		userStatus = "active"
	}

	return &usecase.UserInfo{
		ID:     resp.GetId(),
		Status: userStatus,
	}, nil
}
