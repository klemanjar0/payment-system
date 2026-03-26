package grpc

import (
	pb "github.com/klemanjar0/payment-system/generated/proto/account"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func domainStatusToProto(s domain.AccountStatus) pb.AccountStatus {
	switch s {
	case domain.AccountStatusActive:
		return pb.AccountStatus_ACCOUNT_STATUS_ACTIVE
	case domain.AccountStatusBlocked:
		return pb.AccountStatus_ACCOUNT_STATUS_BLOCKED
	case domain.AccountStatusClosed:
		return pb.AccountStatus_ACCOUNT_STATUS_CLOSED
	default:
		return pb.AccountStatus_ACCOUNT_STATUS_UNSPECIFIED
	}
}

func domainAccountToProto(a *domain.Account) *pb.Account {
	return &pb.Account{
		Id:         a.ID,
		UserId:     a.UserID,
		Currency:   string(a.Currency),
		Balance:    a.Balance.ToInt(),
		HoldAmount: a.HoldAmount.ToInt(),
		Available:  a.Available().ToInt(),
		Status:     domainStatusToProto(a.Status),
		CreatedAt:  timestamppb.New(a.CreatedAt),
		UpdatedAt:  timestamppb.New(a.UpdatedAt),
	}
}
