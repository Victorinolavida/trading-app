-- migrate:up
create table password_reset_tokens (
    user_id uuid primary key references users(id) on delete cascade,
    token_hash bytea not null unique,
    expires_at timestamptz not null
);

-- migrate:down
drop table password_reset_tokens;
