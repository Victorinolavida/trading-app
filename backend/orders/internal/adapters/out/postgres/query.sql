-- name: SaveOrder :exec
INSERT INTO orders(
    id, public_id, trader_id, security_id,
    direction, order_type, price, yield, quantity,
    allocation_ptc, allocation_qty,
    nominal_value, fx, status, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14, $15,$16);

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrderByPublicID :one
SELECT * FROM orders
WHERE public_id = $1;

-- name: UpdateOrder :exec
UPDATE orders
SET status = $1,
    updated_at = $2
WHERE id = $3;
