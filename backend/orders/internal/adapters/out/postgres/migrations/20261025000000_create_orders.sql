-- migrate:up
create table orders (
    id              uuid        primary key,
    public_id       uuid        not null unique,
    trader_id       uuid        not null,
    security_id     uuid        not null,
    direction       text        not null,
    order_type      text        not null,
    price           numeric     ,
    yield           numeric     ,
    allocation_ptc           numeric,
    allocation_qty           int not null,
    quantity        int     not null,
    nominal_value   numeric     not null,
    fx              numeric     not null,
    status          text        not null,
    created_at      timestamptz not null default now(),
    updated_at      timestamptz not null default now(),
    constraint orders_direction_check check (direction in ('bid','ask')),
    constraint orders_order_type_check check (order_type in ('limit','market')),
    constraint orders_status_check check (status in ('active','fully_allocated','cancelled','expired','killed'))
);

create index orders_trader_id_idx on orders (trader_id);
create index orders_security_id_idx on orders (security_id);
create index orders_status_idx on orders (status);

-- migrate:down
drop table orders;
