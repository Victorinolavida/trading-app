-- migrate:up
ALTER TABLE securities ADD COLUMN enable BOOLEAN  default true;


-- migrate:down
ALTER TABLE securities DROP COLUMN IF EXISTS enable;
