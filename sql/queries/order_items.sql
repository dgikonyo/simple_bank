-- Order Items (no primary key → composite operations)
-- name: ListOrderItems :many
SELECT * FROM order_items
WHERE order_id = $1;

-- name: CreateOrderItem :exec
INSERT INTO order_items (
  order_id, product_id, quantity
) VALUES (
  $1, $2, $3
);

-- name: UpdateOrderItem :exec
UPDATE order_items
SET quantity = $3
WHERE order_id = $1
  AND product_id = $2;

-- name: DeleteOrderItem :exec
DELETE FROM order_items
WHERE order_id = $1
  AND product_id = $2;
