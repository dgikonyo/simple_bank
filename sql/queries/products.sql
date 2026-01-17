-- PRODUCTS
-- name: GetProduct :one
SELECT * FROM products
WHERE id = $1 LIMIT 1;

-- name: ListProductsByMerchant :many
SELECT * FROM products
WHERE merchant_id = $1
ORDER BY created_at DESC;

-- name: CreateProduct :one
INSERT INTO products (
  id, name, merchant_id, price, status
) VALUES (
  $1, $2, $3, $4, $5
)
RETURNING *;

-- name: UpdateProduct :exec
UPDATE products
SET name = $2,
    merchant_id = $3,
    price = $4,
    status = $5
WHERE id = $1;

-- name: DeleteProduct :exec
DELETE FROM products
WHERE id = $1;