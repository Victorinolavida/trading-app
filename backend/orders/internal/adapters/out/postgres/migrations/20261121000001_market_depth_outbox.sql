-- migrate:up
CREATE TABLE market_depth_outbox (
    id uuid PRIMARY KEY,
    sequence bigint NOT NULL,
    routing_key text NOT NULL,
    occurred_at timestamptz NOT NULL,
    payload jsonb NOT NULL,
    published_at timestamptz
);
CREATE INDEX market_depth_outbox_pending_idx ON market_depth_outbox (occurred_at, id) WHERE published_at IS NULL;

-- migrate:down
DROP TABLE market_depth_outbox;
