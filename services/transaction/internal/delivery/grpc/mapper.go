package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/klemanjar0/payment-system/generated/proto/transaction"
	"github.com/klemanjar0/payment-system/services/transaction/internal/domain"
)

func domainStatusToProto(s domain.TransactionStatus) pb.TransactionStatus {
	switch s {
	case domain.StatusPending:
		return pb.TransactionStatus_TRANSACTION_STATUS_PENDING
	case domain.StatusProcessing:
		return pb.TransactionStatus_TRANSACTION_STATUS_PROCESSING
	case domain.StatusCompleted:
		return pb.TransactionStatus_TRANSACTION_STATUS_COMPLETED
	case domain.StatusFailed:
		return pb.TransactionStatus_TRANSACTION_STATUS_FAILED
	case domain.StatusReversed:
		return pb.TransactionStatus_TRANSACTION_STATUS_REVERSED
	default:
		return pb.TransactionStatus_TRANSACTION_STATUS_UNSPECIFIED
	}
}

func domainTransactionToProto(tx *domain.Transaction) *pb.Transaction {
	return &pb.Transaction{
		Id:             tx.ID,
		IdempotencyKey: tx.IdempotencyKey,
		FromAccountId:  tx.FromAccountID,
		ToAccountId:    tx.ToAccountID,
		Amount:         tx.Amount.ToInt(),
		Currency:       tx.Currency,
		Description:    tx.Description,
		Status:         domainStatusToProto(tx.Status),
		FailureReason:  tx.FailureReason,
		CreatedAt:      timestamppb.New(tx.CreatedAt),
		UpdatedAt:      timestamppb.New(tx.UpdatedAt),
	}
}
