-- name: SaveOrderGroup :exec
INSERT INTO order_groups(
    id, public_id, trader_id, security_id, direction, order_type,
    allocation_pc, allocation_status, order_group_status,
    weighted_price, weighted_yield, weighted_fx,
    created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14);

-- name: GetOrderGroupByID :one
SELECT * FROM order_groups
WHERE id = $1;

-- name: GetOrderGroupByPublicID :one
SELECT * FROM order_groups
WHERE public_id = $1;

-- name: UpdateOrderGroup :exec
UPDATE order_groups
SET allocation_pc = $1,
    allocation_status = $2,
    order_group_status = $3,
    weighted_price = $4,
    weighted_yield = $5,
    weighted_fx = $6,
    updated_at = $7
WHERE id = $8;

-- name: SaveOrder :exec
INSERT INTO orders(
    id, public_id, order_group_id, trader_id, security_id,
    direction, order_type, price, yield, quantity,
    nominal_value, fx, status, created_at, updated_at
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15);

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrderByPublicID :one
SELECT * FROM orders
WHERE public_id = $1;

-- name: GetOrdersByOrderGroupID :many
SELECT * FROM orders
WHERE order_group_id = $1;

-- name: UpdateOrder :exec
UPDATE orders
SET status = $1,
    updated_at = $2
WHERE id = $3;
