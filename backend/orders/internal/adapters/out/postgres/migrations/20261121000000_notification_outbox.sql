-- migrate:up
CREATE TABLE notification_sequences (
    user_id uuid PRIMARY KEY,
    sequence bigint NOT NULL CHECK (sequence > 0)
);

CREATE TABLE notification_outbox (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    sequence bigint NOT NULL CHECK (sequence > 0),
    routing_key text NOT NULL,
    occurred_at timestamptz NOT NULL,
    payload jsonb NOT NULL,
    published_at timestamptz,
    UNIQUE (user_id, sequence)
);
CREATE INDEX notification_outbox_pending_idx ON notification_outbox (occurred_at, id)
    WHERE published_at IS NULL;

-- migrate:down
DROP TABLE notification_outbox;
DROP TABLE notification_sequences;
