package grpc

import (
	"context"
	"errors"

	//"github.com/klemanjar0/payment-system/pkg/grpcutil"
	pb "github.com/klemanjar0/payment-system/generated/proto/user"
	"github.com/klemanjar0/payment-system/services/user/internal/domain"
	"github.com/klemanjar0/payment-system/services/user/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserHandler struct {
	pb.UnimplementedUserServiceServer
	createUser     *usecase.CreateUserUseCase
	authenticate   *usecase.AuthenticateUseCase
	getUser        *usecase.GetUserUseCase
	changePassword *usecase.ChangePasswordUseCase
}

func NewUserHandler(
	createUser *usecase.CreateUserUseCase,
	authenticate *usecase.AuthenticateUseCase,
	getUser *usecase.GetUserUseCase,
	changePassword *usecase.ChangePasswordUseCase,
) *UserHandler {
	return &UserHandler{
		createUser:     createUser,
		authenticate:   authenticate,
		getUser:        getUser,
		changePassword: changePassword,
	}
}

func (h *UserHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	output, err := h.createUser.Execute(ctx, usecase.CreateUserInput{
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.RegisterResponse{
		UserId:       output.UserID,
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

func (h *UserHandler) Authenticate(ctx context.Context, req *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	output, err := h.authenticate.Execute(ctx, usecase.AuthenticateInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.AuthenticateResponse{
		UserId:       output.UserID,
		AccessToken:  output.AccessToken,
		RefreshToken: output.RefreshToken,
	}, nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	user, err := h.getUser.ExecuteByID(ctx, req.UserId)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return userToProto(user), nil
}

func (h *UserHandler) GetUserByEmail(ctx context.Context, req *pb.GetUserByEmailRequest) (*pb.User, error) {
	user, err := h.getUser.ExecuteByEmail(ctx, req.Email)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return userToProto(user), nil
}

func (h *UserHandler) ValidateUser(ctx context.Context, req *pb.ValidateUserRequest) (*pb.ValidateUserResponse, error) {
	user, err := h.getUser.ExecuteByID(ctx, req.UserId)
	if err != nil {
		return &pb.ValidateUserResponse{
			Valid:     false,
			Status:    pb.UserStatus_USER_STATUS_UNSPECIFIED,
			KycStatus: pb.KYCStatus_KYC_STATUS_UNSPECIFIED,
		}, nil
	}

	return &pb.ValidateUserResponse{
		Valid:     user.IsActive(),
		Status:    statusToProto(user.Status),
		KycStatus: kycStatusToProto(user.KYCStatus),
	}, nil
}

func (h *UserHandler) ChangePassword(ctx context.Context, req *pb.ChangePasswordRequest) (*pb.ChangePasswordResponse, error) {
	err := h.changePassword.Execute(ctx, usecase.ChangePasswordInput{
		UserID:      req.UserId,
		OldPassword: req.OldPassword,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &pb.ChangePasswordResponse{Success: true}, nil
}

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrUserBlocked):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidPhone),
		errors.Is(err, domain.ErrPasswordTooShort),
		errors.Is(err, domain.ErrPasswordTooLong),
		errors.Is(err, domain.ErrPasswordTooWeak):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
