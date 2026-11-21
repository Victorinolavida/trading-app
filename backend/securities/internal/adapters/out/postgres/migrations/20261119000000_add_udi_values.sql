-- migrate:up
CREATE TABLE udi_values (
    date date PRIMARY KEY,
    value numeric(12,6) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);

-- migrate:down
DROP TABLE udi_values;
