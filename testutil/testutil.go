package testutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"

	"transfers/db/sqlc"
)

func GenerateAccount() *db.Account {
	balance := big.NewRat(int64(rand.Intn(1000000000)), 10000)
	return &db.Account{
		ID:      int64(rand.Intn(1000)),
		Balance: balance.FloatString(5),
	}
}

func UnmarshalToResp[T any](t *testing.T, body *bytes.Buffer, resp *T) {
	data, err := ioutil.ReadAll(body)
	require.NoError(t, err)
	err = json.Unmarshal(data, resp)
	require.NoError(t, err)
}
