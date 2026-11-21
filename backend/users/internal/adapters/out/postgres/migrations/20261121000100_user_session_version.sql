-- migrate:up
ALTER TABLE users ADD COLUMN session_version bigint NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE users DROP COLUMN session_version;
