package grpc

import (
	"context"
	"errors"

	pb "github.com/klemanjar0/payment-system/generated/proto/transaction"
	"github.com/klemanjar0/payment-system/pkg/logger"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
	usecase "github.com/klemanjar0/payment-system/services/transaction/internal/use_case"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedTransactionServiceServer
	creator usecase.TransactionCreator
	getter  usecase.TransactionGetter
	lister  usecase.TransactionsByAccountGetter
}

func NewServer(
	creator usecase.TransactionCreator,
	getter usecase.TransactionGetter,
	lister usecase.TransactionsByAccountGetter,
) *Server {
	return &Server{
		creator: creator,
		getter:  getter,
		lister:  lister,
	}
}

func (s *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {
	tx, err := s.creator.Execute(ctx, &usecase.CreateTransferParams{
		IdempotencyKey: req.GetIdempotencyKey(),
		FromAccountID:  req.GetFromAccountId(),
		ToAccountID:    req.GetToAccountId(),
		Amount:         req.GetAmount(),
		Currency:       req.GetCurrency(),
		Description:    req.GetDescription(),
	})
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	return &pb.CreateTransferResponse{
		TransactionId: tx.ID,
		Status:        domainStatusToProto(tx.Status),
	}, nil
}

func (s *Server) GetTransaction(ctx context.Context, req *pb.GetTransactionRequest) (*pb.Transaction, error) {
	tx, err := s.getter.Execute(ctx, req.GetTransactionId())
	if err != nil {
		return nil, domainErrToStatus(err)
	}
	return domainTransactionToProto(tx), nil
}

func (s *Server) GetTransactionsByAccount(ctx context.Context, req *pb.GetTransactionsByAccountRequest) (*pb.TransactionList, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 20
	}
	offset := int(req.GetOffset())

	txs, total, err := s.lister.Execute(ctx, req.GetAccountId(), limit, offset)
	if err != nil {
		return nil, domainErrToStatus(err)
	}

	items := make([]*pb.Transaction, 0, len(txs))
	for _, tx := range txs {
		items = append(items, domainTransactionToProto(tx))
	}

	return &pb.TransactionList{
		Transactions: items,
		Total:        int32(total),
	}, nil
}

func domainErrToStatus(err error) error {
	switch {
	case errors.Is(err, domain.ErrTransactionNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrTransactionExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrSameAccount),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidCurrency):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrAccountNotActive),
		errors.Is(err, domain.ErrCurrencyMismatch),
		errors.Is(err, domain.ErrUserNotActive):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrAccountNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		logger.Error("transaction grpc: internal error", "err", err)
		return status.Error(codes.Internal, "internal error")
	}
}
