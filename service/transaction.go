package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"

	"transfers/api/models"
	db "transfers/db/sqlc"
	"transfers/util"
)

type CreateTransactionService struct {
	db.Store
}

func (s *CreateTransactionService) Validate(ctx context.Context, request *models.CreateTransactionRequest) error {
	amount, err := util.StringToAmount(request.Amount)
	if err != nil {
		return err
	}
	if amount.Sign() <= 0 { // include 0 value as invalid
		return util.NewInvalidAmountError(request.Amount)
	}
	request.Amount = util.AmountToString(amount)
	if request.SourceAccountID == request.DestinationAccountID {
		return util.NewTransactionToSameAccountError(request.SourceAccountID)
	}
	_, err = s.GetAccount(ctx, request.SourceAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return util.NewAccountNotFoundError(request.SourceAccountID)
		}
		return util.NewDBError(err)
	}
	_, err = s.GetAccount(ctx, request.DestinationAccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return util.NewAccountNotFoundError(request.DestinationAccountID)
		}
		return util.NewDBError(err)
	}
	return nil
}

func (s *CreateTransactionService) Do(ctx context.Context, request *models.CreateTransactionRequest) (*models.CreateTransactionResponse, error) {
	transaction, err := s.CreateTransactionWithSSI(ctx, &db.CreateTransactionParams{
		SourceAccountID:      request.SourceAccountID,
		DestinationAccountID: request.DestinationAccountID,
		Amount:               request.Amount,
	})
	if err != nil {
		return nil, err
	}
	return &models.CreateTransactionResponse{
		SourceAccountID:      transaction.SourceAccountID,
		DestinationAccountID: transaction.DestinationAccountID,
		Amount:               transaction.Amount,
	}, nil
}
