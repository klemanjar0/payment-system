package client

import (
	"context"
	"fmt"

	pbaccount "github.com/klemanjar0/payment-system/generated/proto/account"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
	usecase "github.com/klemanjar0/payment-system/services/transaction/internal/use_case"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountClient struct {
	client pbaccount.AccountServiceClient
}

func NewAccountClient(conn *grpc.ClientConn) *AccountClient {
	return &AccountClient{client: pbaccount.NewAccountServiceClient(conn)}
}

func (c *AccountClient) GetAccount(ctx context.Context, accountID string) (*usecase.AccountInfo, error) {
	resp, err := c.client.GetAccount(ctx, &pbaccount.GetAccountRequest{AccountId: accountID})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			if st.Code() == codes.NotFound {
				return nil, domain.ErrAccountNotFound
			}
		}
		return nil, fmt.Errorf("account service GetAccount: %w", err)
	}

	accountStatus := "active"
	if resp.GetStatus() != pbaccount.AccountStatus_ACCOUNT_STATUS_ACTIVE {
		accountStatus = "inactive"
	}

	return &usecase.AccountInfo{
		ID:       resp.GetId(),
		UserID:   resp.GetUserId(),
		Currency: resp.GetCurrency(),
		Status:   accountStatus,
	}, nil
}
