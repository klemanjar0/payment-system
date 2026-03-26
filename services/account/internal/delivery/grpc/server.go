package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/klemanjar0/payment-system/generated/proto/account"
	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
	usecase "github.com/klemanjar0/payment-system/services/account/internal/use_case"
)

type Server struct {
	pb.UnimplementedAccountServiceServer

	createAccount    usecase.AccountCreator
	getAccount       usecase.AccountGetter
	getAccountByUser usecase.AccountsByUserGetter
}

func NewServer(
	createAccount usecase.AccountCreator,
	getAccount usecase.AccountGetter,
	getAccountByUser usecase.AccountsByUserGetter,
) *Server {
	return &Server{
		createAccount:    createAccount,
		getAccount:       getAccount,
		getAccountByUser: getAccountByUser,
	}
}

func (s *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.Account, error) {
	logger.Info("CreateAccount request", "user_id", req.UserId, "currency", req.Currency)

	account, err := s.createAccount.Execute(ctx, &usecase.CreateAccountUseCaseParams{
		UserID:   req.UserId,
		Currency: domain.Currency(req.Currency),
	})
	if err != nil {
		logGrpcError("CreateAccount", err, "user_id", req.UserId)
		return nil, domainErrToStatus(err)
	}

	logger.Info("CreateAccount success", "account_id", account.ID, "user_id", req.UserId)
	return domainAccountToProto(account), nil
}

func (s *Server) GetAccount(ctx context.Context, req *pb.GetAccountRequest) (*pb.Account, error) {
	logger.Info("GetAccount request", "account_id", req.AccountId)

	account, err := s.getAccount.Execute(ctx, req.AccountId)
	if err != nil {
		logGrpcError("GetAccount", err, "account_id", req.AccountId)
		return nil, domainErrToStatus(err)
	}

	return domainAccountToProto(account), nil
}

func (s *Server) GetAccountsByUser(ctx context.Context, req *pb.GetAccountsByUserRequest) (*pb.AccountList, error) {
	logger.Info("GetAccountsByUser request", "user_id", req.UserId)

	accounts, err := s.getAccountByUser.Execute(ctx, req.UserId)
	if err != nil {
		logGrpcError("GetAccountsByUser", err, "user_id", req.UserId)
		return nil, domainErrToStatus(err)
	}

	list := make([]*pb.Account, len(accounts))
	for i, a := range accounts {
		list[i] = domainAccountToProto(a)
	}

	return &pb.AccountList{Accounts: list}, nil
}

// logGrpcError logs only internal (unexpected) errors — not domain errors like NotFound.
func logGrpcError(method string, err error, keysAndValues ...any) {
	if isInternalErr(err) {
		logger.Error("grpc handler error", append([]any{"method", method, "err", err}, keysAndValues...)...)
	}
}

func isInternalErr(err error) bool {
	return !errors.Is(err, domain.ErrAccountNotFound) &&
		!errors.Is(err, domain.ErrAccountExists) &&
		!errors.Is(err, domain.ErrInvalidCurrency) &&
		!errors.Is(err, domain.ErrAccountNotActive)
}

func domainErrToStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrAccountNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrAccountExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidCurrency):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrAccountNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
