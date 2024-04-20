package models

type CreateAccountRequest struct {
	AccountID      int64  `json:"account_id" binding:"required,min=1"`
	InitialBalance string `json:"initial_balance" binding:"required"`
}
type CreateAccountResponse struct {
	AccountID      int64  `json:"account_id,omitempty"`
	InitialBalance string `json:"initial_balance,omitempty"`
}

type GetAccountRequest struct {
	AccountID int64 `uri:"account_id" binding:"required,min=1"`
}
type GetAccountResponse struct {
	AccountID int64  `json:"account_id,omitempty"`
	Balance   string `json:"balance,omitempty"`
}

type CreateTransactionRequest struct {
	SourceAccountID      int64  `json:"source_account_id" binding:"required,min=1"`
	DestinationAccountID int64  `json:"destination_account_id" binding:"required,min=1"`
	Amount               string `json:"amount" binding:"required"`
}
type CreateTransactionResponse struct {
	SourceAccountID      int64  `json:"source_account_id,omitempty"`
	DestinationAccountID int64  `json:"destination_account_id,omitempty"`
	Amount               string `json:"amount,omitempty"`
}
