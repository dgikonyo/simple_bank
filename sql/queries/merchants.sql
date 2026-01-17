-- MERCHANTS
-- name: GetMerchant :one
SELECT * FROM merchants
WHERE id = $1 LIMIT 1;

-- name: ListMerchants :many
SELECT * FROM merchants
ORDER BY merchant_name;

-- name: CreateMerchant :one
INSERT INTO merchants (
  merchant_name, country_code, admin_id
) VALUES (
  $1, $2, $3
)
RETURNING *;

-- name: UpdateMerchant :exec
UPDATE merchants
SET merchant_name = $2,
    country_code = $3,
    admin_id = $4
WHERE id = $1;

-- name: DeleteMerchant :exec
DELETE FROM merchants
WHERE id = $1;
