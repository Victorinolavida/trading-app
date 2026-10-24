-- migrate:up
ALTER TABLE securities ADD COLUMN IF NOT EXISTS name character varying(128) NOT NULL DEFAULT '';

-- migrate:down
ALTER TABLE securities DROP COLUMN IF EXISTS name;
