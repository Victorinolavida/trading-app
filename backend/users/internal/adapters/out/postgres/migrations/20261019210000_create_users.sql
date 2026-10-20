-- migrate:up
create table users (
    id            uuid        primary key,
    email         text        not null unique,
    first_name    text        not null,
    last_name     text        not null,
    password_hash text        not null,
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now(),
    constraint users_first_name_len check (char_length(first_name) between 3 and 49),
    constraint users_last_name_len  check (char_length(last_name) between 3 and 49)
);

create index users_created_at_idx on users (created_at);

-- migrate:down
drop table users;
