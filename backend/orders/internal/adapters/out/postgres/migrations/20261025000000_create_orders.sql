-- migrate:up
create table order_groups (
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
    updated_at        timestamptz not null default now(),
    constraint order_groups_direction_check check (direction in ('bid','ask')),
    constraint order_groups_order_type_check check (order_type in ('limit','market')),
    constraint order_groups_allocation_status_check check (allocation_status in ('pending','partial','allocated')),
    constraint order_groups_status_check check (order_group_status in ('active','allocated','cancelled','expired'))
);

create index order_groups_trader_id_idx on order_groups (trader_id);
create index order_groups_security_id_idx on order_groups (security_id);
create index order_groups_status_idx on order_groups (order_group_status);

create table orders (
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
    updated_at      timestamptz not null default now(),
    constraint orders_direction_check check (direction in ('bid','ask')),
    constraint orders_order_type_check check (order_type in ('limit','market')),
    constraint orders_status_check check (status in ('active','fully_allocated','cancelled','expired','killed')),
    constraint orders_quantity_check check (quantity > 0)
);

create index orders_order_group_id_idx on orders (order_group_id);
create index orders_trader_id_idx on orders (trader_id);
create index orders_security_id_idx on orders (security_id);
create index orders_status_idx on orders (status);

-- migrate:down
drop table orders;
drop table order_groups;
