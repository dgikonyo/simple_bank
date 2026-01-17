-- TRANSFERS
-- name: GetTransfer :one
SELECT * FROM tranfers
WHERE id = $1 LIMIT 1;

-- name: ListTransfers :many
SELECT * FROM tranfers
ORDER BY created_at DESC;

-- name: CreateTransfer :one
INSERT INTO tranfers (
  from_account_id, to_account_id, amount
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: DeleteTransfer :exec
DELETE FROM tranfers
WHERE id = $1;