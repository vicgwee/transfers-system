package util

import (
	"github.com/joomcode/errorx"
)

var (
	TransfersSystemErrors = errorx.NewNamespace("transfers", PaymentRequired)

	// Traits
	PaymentRequired = errorx.RegisterTrait("payment_required")

	// Types
	ErrAccountNotFound     = TransfersSystemErrors.NewType("account_not_found", errorx.NotFound())
	ErrDuplicateAccount    = TransfersSystemErrors.NewType("duplicate_account", errorx.Duplicate())
	ErrInsufficientBalance = TransfersSystemErrors.NewType("insufficient_balance", PaymentRequired)
)

func NewDBError(err error) *errorx.Error {
	return errorx.ExternalError.Wrap(err, "DB Error")
}

func NewInvalidAmountError(val string) *errorx.Error {
	return errorx.IllegalArgument.New("invalid amount: %s", val)
}

func NewNegativeBalanceError(val string) *errorx.Error {
	return errorx.IllegalArgument.New("negative balance: %s", val)
}

func NewAccountAlreadyExistsError(id int64) *errorx.Error {
	return ErrDuplicateAccount.New("account already exists: %d", id)
}

func NewAccountNotFoundError(id int64) *errorx.Error {
	return ErrAccountNotFound.New("account not found: %d", id)
}

func NewInvalidIDError(id int64) *errorx.Error {
	return errorx.IllegalArgument.New("invalid ID: %d", id)
}

func NewTransactionToSameAccountError(accountId int64) *errorx.Error {
	return errorx.IllegalArgument.New("invalid transaction with same source and destination account: %d", accountId)
}

func NewInsufficientBalanceError() *errorx.Error {
	return ErrInsufficientBalance.New("insufficient balance in debiting account")
}
