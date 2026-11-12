-- migrate:up

ALTER TABLE orders ADD COLUMN expires_at  timestamptz not null default now();


-- migrate:down
ALTER TABLE orders DROP COLUMN  expires_at;

