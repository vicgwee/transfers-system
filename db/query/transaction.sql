-- name: CreateTransaction :one
INSERT INTO transactions (
    source_account_id,
    destination_account_id,
    amount
) VALUES (
  $1, $2, $3
) RETURNING *;


-- name: DeleteAllTransactions :exec
DELETE FROM transactions;