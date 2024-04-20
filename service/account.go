package service

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5"

	db "transfers/db/sqlc"
	"transfers/util"

	"transfers/api/models"
)

type GetAccountService struct {
	db.Store
}

func (s *GetAccountService) Validate(ctx context.Context, request *models.GetAccountRequest) error {
	return nil
}

func (s *GetAccountService) Do(ctx context.Context, request *models.GetAccountRequest) (*models.GetAccountResponse, error) {
	account, err := s.GetAccount(ctx, request.AccountID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, util.NewAccountNotFoundError(request.AccountID)
		}
		return nil, util.NewDBError(err)
	}
	return &models.GetAccountResponse{
		AccountID: account.ID,
		Balance:   account.Balance,
	}, nil
}

type CreateAccountService struct {
	db.Store
}

func (s *CreateAccountService) Validate(ctx context.Context, request *models.CreateAccountRequest) error {
	if request.AccountID < 0 {
		return util.NewInvalidIDError(request.AccountID)
	}
	balance, err := util.StringToAmount(request.InitialBalance)
	if err != nil {
		return err
	}
	log.Printf("balance: %v", balance.FloatString(5))
	if balance.Sign() == -1 {
		return util.NewNegativeBalanceError(request.InitialBalance)
	}
	request.InitialBalance = util.AmountToString(balance)
	_, err = s.GetAccount(ctx, request.AccountID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil // no existing account
	}
	if err != nil {
		return util.NewDBError(err)
	}
	return util.NewAccountAlreadyExistsError(request.AccountID)
}

func (s *CreateAccountService) Do(ctx context.Context, request *models.CreateAccountRequest) (*models.CreateAccountResponse, error) {
	account, err := s.CreateAccount(ctx, &db.CreateAccountParams{
		ID:      request.AccountID,
		Balance: request.InitialBalance,
	})
	if err != nil {
		return nil, util.NewDBError(err)
	}
	return &models.CreateAccountResponse{
		AccountID:      account.ID,
		InitialBalance: account.Balance,
	}, nil

}
