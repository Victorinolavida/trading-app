-- name: SaveOrder :exec
INSERT INTO orders (
    id, public_id, trader_id, security_id,
    direction, order_type, price, yield, quantity,
    allocation_qty, allocation_status,
    nominal_value, fx, status, created_at, updated_at, expires_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8, $9,
    $10, $11,
    $12, $13, $14, $15, $16, $17
);

-- name: GetOrderByID :one
SELECT * FROM orders
WHERE id = $1;

-- name: GetOrderByPublicID :one
SELECT * FROM orders
WHERE public_id = $1;

-- name: GetActiveOrders :many
SELECT * FROM orders
WHERE 
    security_id = $1
    AND trader_id = $2
    AND status = 'active'
    AND (
            sqlc.narg('direction')::text IS NULL
            OR direction = sqlc.narg('direction')::text
        )
ORDER BY created_at DESC, id ASC;

-- name: ListOrders :many
SELECT * FROM orders
WHERE
    trader_id = $1
    AND (
            sqlc.narg('order_ids')::uuid[] IS NULL
            OR cardinality(sqlc.narg('order_ids')::uuid[]) = 0
            OR public_id = ANY(sqlc.narg('order_ids')::uuid[])
        )
    AND (
            sqlc.narg('direction')::text IS NULL
            OR direction = sqlc.narg('direction')::text
        )
    AND (
            sqlc.narg('status')::text IS NULL
            OR status = sqlc.narg('status')::text
        )
ORDER BY created_at DESC, id ASC;

-- name: UpdateOrder :execrows
UPDATE orders
SET status            = $1,
    allocation_qty    = $2,
    allocation_status = $3,
    updated_at        = $4
WHERE id = $5;

-- name: UpdateOrders :execrows
UPDATE orders AS o
SET status            = u.status,
    allocation_qty    = u.allocation_qty,
    allocation_status = u.allocation_status,
    updated_at        = u.updated_at
FROM (
    SELECT unnest(sqlc.arg('ids')::uuid[])                AS id,
           unnest(sqlc.arg('statuses')::text[])            AS status,
           unnest(sqlc.arg('allocation_qtys')::bigint[])  AS allocation_qty,
           unnest(sqlc.arg('allocation_statuses')::text[]) AS allocation_status,
           unnest(sqlc.arg('updated_ats')::timestamptz[])  AS updated_at
) AS u
WHERE o.id = u.id;

-- name: SaveTrades :execrows
INSERT INTO trades (
    id, public_id, maker_order_id, taker_order_id,
    price, quantity, taker_side, executed_at
)
SELECT unnest(sqlc.arg('ids')::uuid[]),
       unnest(sqlc.arg('public_ids')::uuid[]),
       unnest(sqlc.arg('maker_order_ids')::uuid[]),
       unnest(sqlc.arg('taker_order_ids')::uuid[]),
       unnest(sqlc.arg('prices')::bigint[]),
       unnest(sqlc.arg('quantities')::bigint[]),
       unnest(sqlc.arg('taker_sides')::text[]),
       unnest(sqlc.arg('executed_ats')::timestamptz[]);

-- name: GetTradeByID :one
SELECT * FROM trades
WHERE id = $1;

-- name: GetTradesByOrderID :many
SELECT * FROM trades
WHERE maker_order_id = $1 OR taker_order_id = $1
ORDER BY executed_at ASC, id ASC;

-- name: GetBidsToTrade :many
SELECT * FROM orders
WHERE direction = 'bid'
  AND status = 'active'
  AND security_id = $1
ORDER BY COALESCE(price, 9223372036854775807) DESC, created_at ASC, id ASC;

-- name: GetAsksToTrade :many
SELECT * FROM orders
WHERE direction = 'ask'
  AND status = 'active'
  AND security_id = $1
ORDER BY COALESCE(price, 0) ASC, created_at ASC, id ASC;



-- name: AskMarketDepth :many
select
    price,
    sum(nominal_value)::numeric as agg_nominal_value,
    sum(quantity)::numeric      as agg_qty,
    count(*)                    as order_count
from orders
where status = 'active'
  and security_id =  $1
  and direction  = 'ask'
group by price
order by price asc;

-- name: BidMarketDepth :many
select
    price,
    sum(nominal_value)::numeric as agg_nominal_value,
    sum(quantity)::numeric      as agg_qty,
    count(*)                    as order_count
from orders
where status = 'active'
  and security_id =  $1
  and direction  = 'bid'
group by price
order by price desc;

-- name: ExpireOrders :many
UPDATE orders
SET status = 'expired', updated_at = now()
WHERE status = 'active' AND expires_at < now()
RETURNING *;
