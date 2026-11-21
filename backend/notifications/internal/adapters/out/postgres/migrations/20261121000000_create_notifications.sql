-- migrate:up
CREATE TABLE notifications (
    id uuid PRIMARY KEY,
    user_id uuid NOT NULL,
    event_id uuid NOT NULL,
    event_type text NOT NULL,
    payload jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    read_at timestamptz,
    UNIQUE (event_id, user_id)
);
CREATE INDEX notifications_user_created_idx ON notifications (user_id, created_at DESC, id DESC);
CREATE INDEX notifications_user_unread_idx ON notifications (user_id, created_at DESC, id DESC) WHERE read_at IS NULL;

-- migrate:down
DROP TABLE notifications;
