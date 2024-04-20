package db

import (
	"context"
	"math/big"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"

	"transfers/util"
)

func TestPgxStore_CreateTransactionTx(t *testing.T) {
	accountA := &CreateAccountParams{
		ID:      1,
		Balance: "100.0",
	}
	accountB := &CreateAccountParams{
		ID:      2,
		Balance: "100.0",
	}
	accounts := []*CreateAccountParams{accountA, accountB}

	tests := []struct {
		name           string
		param          *CreateTransactionParams
		want           *Transaction
		wantTransacted string
		wantErr        bool
	}{
		{
			name: "success",
			param: &CreateTransactionParams{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "100.00000",
			},
			want: &Transaction{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "100.00000",
			},
			wantTransacted: "100.00000",
		},
		{
			name: "insufficient balance",
			param: &CreateTransactionParams{
				SourceAccountID:      1,
				DestinationAccountID: 2,
				Amount:               "100.00001",
			},
			wantErr:        true,
			wantTransacted: "0.00000",
		},
		{
			name: "missing destination account",
			param: &CreateTransactionParams{
				SourceAccountID:      1,
				DestinationAccountID: 0,
				Amount:               "100.00001",
			},
			wantErr:        true,
			wantTransacted: "0.00000",
		},
		{
			name: "missing source account",
			param: &CreateTransactionParams{
				SourceAccountID:      0,
				DestinationAccountID: 2,
				Amount:               "100.00001",
			},
			wantErr:        true,
			wantTransacted: "0.00000",
		},
	}
	ctx := context.Background()
	s := testStore

	for _, tt := range tests {
		for _, fn := range []CreateTransactionFunc{
			s.CreateTransactionWithLock,
			s.CreateTransactionWithSSI,
		} {
			t.Run(tt.name, func(t *testing.T) {
				setup(t, accounts)
				defer teardown(t)
				got, err := fn(ctx, tt.param)
				require.Equal(t, tt.wantErr, err != nil)
				if err == nil {
					require.Equal(t, got.Amount, tt.want.Amount)
					require.Equal(t, got.SourceAccountID, tt.want.SourceAccountID)
					require.Equal(t, got.DestinationAccountID, tt.want.DestinationAccountID)
				}
				accA, err := s.GetAccount(ctx, accountA.ID)
				require.NoError(t, err)
				requireBalanceChange(t, accountA.Balance, accA.Balance, "-"+tt.wantTransacted)
				accB, err := s.GetAccount(ctx, accountB.ID)
				require.NoError(t, err)
				requireBalanceChange(t, accountB.Balance, accB.Balance, tt.wantTransacted)
			})
		}
	}
}

func TestPgxStore_CreateTransactionDeadlock(t *testing.T) {
	numTransactions := 50
	initialBalance := strconv.Itoa(numTransactions * 2)
	amountTransacted := strconv.Itoa(numTransactions)

	accountA := &CreateAccountParams{
		ID:      1,
		Balance: initialBalance,
	}
	accountB := &CreateAccountParams{
		ID:      2,
		Balance: initialBalance,
	}
	accounts := []*CreateAccountParams{accountA, accountB}

	debit := &CreateTransactionParams{
		SourceAccountID:      accountA.ID,
		DestinationAccountID: accountB.ID,
		Amount:               "2.00000",
	}
	credit := &CreateTransactionParams{
		SourceAccountID:      accountB.ID,
		DestinationAccountID: accountA.ID,
		Amount:               "1.0000",
	}
	ctx := context.Background()
	s := testStore

	for _, fn := range []CreateTransactionFunc{
		s.CreateTransactionWithLock,
		s.CreateTransactionWithSSI,
	} {
		t.Run("Test Transaction Lock", func(t *testing.T) {
			{
				setup(t, accounts)
				defer teardown(t)

				var errCnt atomic.Int64
				var wg sync.WaitGroup
				wg.Add(numTransactions * 2)
				for i := 0; i < numTransactions; i++ {
					go func() {
						defer wg.Done()
						_, err := fn(ctx, debit)
						if err != nil {
							errCnt.Add(1)
						}
					}()
					go func() {
						defer wg.Done()
						_, err := fn(ctx, credit)
						if err != nil {
							errCnt.Add(1)
						}
					}()
				}
				wg.Wait()
				require.Equal(t, errCnt.Load(), int64(0))
				accA, err := s.GetAccount(ctx, accountA.ID)
				require.NoError(t, err)
				requireBalanceChange(t, initialBalance, accA.Balance, "-"+amountTransacted)
				accB, err := s.GetAccount(ctx, accountB.ID)
				require.NoError(t, err)
				requireBalanceChange(t, initialBalance, accB.Balance, amountTransacted)
			}
		})
	}
}

func setup(t *testing.T, accounts []*CreateAccountParams) {
	ctx := context.Background()
	s := testStore
	for _, account := range accounts {
		_, err := s.CreateAccount(ctx, account)
		require.NoError(t, err)
	}
}

func teardown(t *testing.T) {
	ctx := context.Background()
	s := testStore
	require.NoError(t, s.DeleteAllTransactions(ctx))
	require.NoError(t, s.DeleteAllAccounts(ctx))
}

func requireBalanceChange(t *testing.T, initialBalance string, finalBalance string, amountTransacted string) {
	initialV, err := util.StringToAmount(initialBalance)
	require.NoError(t, err)
	finalV, err := util.StringToAmount(finalBalance)
	require.NoError(t, err)
	amountTransactedV, err := util.StringToAmount(amountTransacted)
	require.NoError(t, err)
	expectedV := *new(big.Rat).Add(&initialV, &amountTransactedV)
	require.Equal(t, util.AmountToString(expectedV), util.AmountToString(finalV))
}
