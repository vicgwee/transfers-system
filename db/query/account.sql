-- name: CreateAccount :one
INSERT INTO accounts (
  id,
  balance
) VALUES (
  $1, $2
) RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccountForUpdate :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;

-- name: DeleteAllAccounts :exec
DELETE FROM accounts;
