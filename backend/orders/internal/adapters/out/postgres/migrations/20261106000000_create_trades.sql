-- migrate:up
create table trades (
    id               uuid        primary key,
    public_id        uuid        not null unique,
    maker_order_id   uuid        not null references orders(id),
    taker_order_id   uuid        not null references orders(id),
    price            bigint      not null,
    quantity         bigint      not null,
    taker_side       text        not null,
    executed_at      timestamptz not null default now(),

    constraint trades_price_check
        check (price > 0),
    constraint trades_quantity_check
        check (quantity > 0),
    constraint trades_taker_side_check
        check (taker_side in ('bid','ask'))
);

create index trades_maker_order_id_idx
    on trades (maker_order_id);

create index trades_taker_order_id_idx
    on trades (taker_order_id);

-- migrate:down
drop table trades;
