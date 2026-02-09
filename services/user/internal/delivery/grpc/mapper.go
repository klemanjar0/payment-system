package grpc

import (
	pb "github.com/klemanjar0/payment-system/generated/proto/user"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func userToProto(u *domain.User) *pb.User {
	return &pb.User{
		Id:        u.ID,
		Email:     u.Email,
		Phone:     u.Phone,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Status:    statusToProto(u.Status),
		KycStatus: kycStatusToProto(u.KYCStatus),
		CreatedAt: timestamppb.New(u.CreatedAt),
		UpdatedAt: timestamppb.New(u.UpdatedAt),
	}
}

func statusToProto(s domain.UserStatus) pb.UserStatus {
	switch s {
	case domain.UserStatusPending:
		return pb.UserStatus_USER_STATUS_PENDING
	case domain.UserStatusActive:
		return pb.UserStatus_USER_STATUS_ACTIVE
	case domain.UserStatusBlocked:
		return pb.UserStatus_USER_STATUS_BLOCKED
	case domain.UserStatusDeleted:
		return pb.UserStatus_USER_STATUS_DELETED
	default:
		return pb.UserStatus_USER_STATUS_UNSPECIFIED
	}
}

func kycStatusToProto(s domain.KYCStatus) pb.KYCStatus {
	switch s {
	case domain.KYCStatusNone:
		return pb.KYCStatus_KYC_STATUS_NONE
	case domain.KYCStatusPending:
		return pb.KYCStatus_KYC_STATUS_PENDING
	case domain.KYCStatusVerified:
		return pb.KYCStatus_KYC_STATUS_VERIFIED
	case domain.KYCStatusRejected:
		return pb.KYCStatus_KYC_STATUS_REJECTED
	default:
		return pb.KYCStatus_KYC_STATUS_UNSPECIFIED
	}
}
