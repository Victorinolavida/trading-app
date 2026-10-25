CREATE TABLE order_groups (
    id                uuid        primary key,
    public_id         uuid        not null unique,
    trader_id         uuid        not null,
    security_id       uuid        not null,
    direction         text        not null,
    order_type        text        not null,
    allocation_pc     numeric     not null default 0,
    allocation_status text        not null,
    order_group_status text       not null,
    weighted_price    numeric     not null,
    weighted_yield    numeric     not null,
    weighted_fx       numeric     not null,
    created_at        timestamptz not null default now(),
    updated_at        timestamptz not null default now()
);

CREATE TABLE orders (
    id              uuid        primary key,
    public_id       uuid        not null unique,
    order_group_id  uuid        not null references order_groups(id),
    trader_id       uuid        not null,
    security_id     uuid        not null,
    direction       text        not null,
    order_type      text        not null,
    price           numeric     not null,
    yield           numeric     not null,
    quantity        numeric     not null,
    nominal_value   numeric     not null,
    fx              numeric     not null,
    status          text        not null,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now()
);
