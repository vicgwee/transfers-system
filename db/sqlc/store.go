package db

import (
	"bytes"
	"context"
	"math/rand"
	"strings"
	"text/template"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"transfers/util"
)

type Store interface {
	Querier
	CreateTransactionWithLock(ctx context.Context, param *CreateTransactionParams) (*Transaction, error)
	CreateTransactionWithSSI(ctx context.Context, param *CreateTransactionParams) (*Transaction, error)
}

type CreateTransactionFunc func(context.Context, *CreateTransactionParams) (*Transaction, error)

type PgxStore struct {
	*Queries
	dbConn *pgxpool.Pool
}

func NewPgxStore(dbConn *pgxpool.Pool) *PgxStore {
	return &PgxStore{
		dbConn:  dbConn,
		Queries: New(dbConn),
	}
}

// TODO: Tune these config settings based on the performance of the server hardware
const (
	maxRetries     = 5
	initialRetryMs = 50
	randomRetryMs  = 200
)

// CreateTransactionWithLock handles creating transaction and updating account balances safely with DB locking
func (s *PgxStore) CreateTransactionWithLock(ctx context.Context, param *CreateTransactionParams) (*Transaction, error) {
	// Prevent deadlock by updating in consistent order based on accountID
	sqlParam := toCreateTransactionSqlParams(param)

	var transaction *Transaction
	var err error
	txOptions := pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	}
	err = s.doTx(ctx, txOptions, func(tx DBTX) error {
		q := New(tx)

		lowAccount, err := q.GetAccountForUpdate(ctx, sqlParam.LowAccountID)
		if err != nil {
			return util.NewDBError(err)
		}
		highAccount, err := q.GetAccountForUpdate(ctx, sqlParam.HighAccountID)
		if err != nil {
			return util.NewDBError(err)
		}
		sourceAccount, destinationAccount := lowAccount, highAccount
		if param.SourceAccountID == sqlParam.HighAccountID {
			sourceAccount, destinationAccount = highAccount, lowAccount
		}

		// Check and update balances
		transferAmount, err := util.StringToAmount(param.Amount)
		if err != nil {
			return err
		}
		sourceBalance, err := util.StringToAmount(sourceAccount.Balance)
		if err != nil {
			return err
		}
		destinationBalance, err := util.StringToAmount(destinationAccount.Balance)
		if err != nil {
			return err
		}
		if sourceBalance.Cmp(&transferAmount) < 0 {
			return util.NewInsufficientBalanceError()
		}
		sourceBalance.Sub(&sourceBalance, &transferAmount)
		destinationBalance.Add(&destinationBalance, &transferAmount)

		// Write updates to DB
		_, err = q.UpdateAccount(ctx, &UpdateAccountParams{
			ID:      sourceAccount.ID,
			Balance: util.AmountToString(sourceBalance),
		})
		if err != nil {
			return util.NewDBError(err)
		}
		_, err = q.UpdateAccount(ctx, &UpdateAccountParams{
			ID:      destinationAccount.ID,
			Balance: util.AmountToString(destinationBalance),
		})
		if err != nil {
			return util.NewDBError(err)
		}
		transaction, err = q.CreateTransaction(ctx, param)
		if err != nil {
			return util.NewDBError(err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return transaction, nil
}

// doTx executes fn within a DB transaction with txOptions
func (s *PgxStore) doTx(ctx context.Context, txOptions pgx.TxOptions, fn func(DBTX) error) error {
	tx, err := s.dbConn.BeginTx(ctx, txOptions)
	if err != nil {
		return util.NewDBError(err)
	}
	err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return util.NewDBError(rbErr).WithUnderlyingErrors(err)
		}
		return util.NewDBError(err)
	}
	return tx.Commit(ctx)
}

const createTransactionSqlTemplate = `
--name: CreateTransactionSql :exec
BEGIN TRANSACTION ISOLATION LEVEL SERIALIZABLE; 

UPDATE accounts
SET balance = balance + {{.AddLowAmount}}
WHERE id = {{.LowAccountID}};

UPDATE accounts
SET balance = balance + {{.AddHighAmount}}
WHERE id = {{.HighAccountID}};

INSERT INTO transactions (
	source_account_id,
	destination_account_id,
	amount
) VALUES (
  {{.SourceAccountID}}, {{.DestinationAccountID}}, {{.Amount}}
);

COMMIT;`

/*
CreateTransactionWithSSI handles creating transaction and updating account balances safely with SSI
*/
func (s *PgxStore) CreateTransactionWithSSI(ctx context.Context, param *CreateTransactionParams) (*Transaction, error) {
	// Prevent deadlock by updating in consistent order based on accountID
	sqlParam := toCreateTransactionSqlParams(param)

	tmpl, err := template.New("CreateTransactionSql").Parse(createTransactionSqlTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, sqlParam)
	if err != nil {
		return nil, err
	}
	sql := buf.String()

	retryTime := initialRetryMs
	for i := 0; i < maxRetries; i++ {
		_, err = s.dbConn.Exec(ctx, sql)
		if err == nil {
			return &Transaction{
				SourceAccountID:      param.SourceAccountID,
				DestinationAccountID: param.DestinationAccountID,
				Amount:               param.Amount,
			}, nil
		}
		if strings.Contains(err.Error(), "(SQLSTATE 40001)") || // serialization failure
			strings.Contains(err.Error(), "(SQLSTATE 40P01)") { // deadlock detected
			time.Sleep(time.Millisecond * time.Duration(retryTime+rand.Intn(randomRetryMs)))
			retryTime *= 2
			continue
		}
		if strings.Contains(err.Error(), "(SQLSTATE 23514)") { // constraint violated
			// Since we have already checked for valid account ID before this, the other
			// DB constraint is balance >= 0
			return nil, util.NewInsufficientBalanceError()
		}
		return nil, util.NewDBError(err)
	}
	return nil, util.NewDBError(err)
}

type createTransactionSqlParam struct {
	Amount               string
	SourceAccountID      int64
	DestinationAccountID int64
	LowAccountID         int64
	HighAccountID        int64
	AddLowAmount         string
	AddHighAmount        string
}

func toCreateTransactionSqlParams(param *CreateTransactionParams) *createTransactionSqlParam {
	sqlParam := createTransactionSqlParam{
		Amount:               param.Amount,
		SourceAccountID:      param.SourceAccountID,
		DestinationAccountID: param.DestinationAccountID,
		LowAccountID:         param.SourceAccountID,
		HighAccountID:        param.DestinationAccountID,
		AddLowAmount:         "-" + param.Amount,
		AddHighAmount:        param.Amount,
	}
	// Swap
	if sqlParam.HighAccountID > sqlParam.LowAccountID {
		sqlParam.LowAccountID, sqlParam.HighAccountID = sqlParam.HighAccountID, sqlParam.LowAccountID
		sqlParam.AddLowAmount, sqlParam.AddHighAmount = sqlParam.AddHighAmount, sqlParam.AddLowAmount
	}
	return &sqlParam
}
