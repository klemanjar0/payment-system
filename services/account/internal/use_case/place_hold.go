package usecase

import (
	"context"

	"github.com/klemanjar0/payment-system/pkg/money"
	"github.com/klemanjar0/payment-system/services/account/internal/domain"
)

type PlaceHoldUseCaseParams struct {
	AccountID, TransactionID string
	Amount                   money.Amount
}

type PlaceHoldUseCase struct {
	manager domain.TransactionManager
}

func NewCreditUseCase(manager domain.TransactionManager) *PlaceHoldUseCase {
	return &PlaceHoldUseCase{
		manager: manager,
	}
}

func (uc *PlaceHoldUseCase) Execute(
	ctx context.Context,
	params *PlaceHoldUseCaseParams,
) error {
	err := uc.manager.ExecTx(ctx, func(tx domain.TxRepositories) error {
		account, err := tx.Accounts.GetByIDForUpdate(ctx, params.AccountID)

		if err != nil {
			return err
		}

		account.Hold(params.Amount)
		err = tx.Accounts.Update(ctx, account)

		if err != nil {
			return err
		}

		err = tx.Holds.Create(ctx, domain.NewHold(
			params.AccountID,
			params.TransactionID,
			params.Amount,
			"system",
		))

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}
