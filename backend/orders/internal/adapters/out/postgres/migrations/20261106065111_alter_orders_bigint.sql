-- migrate:up
alter table orders
    alter column price type bigint using price::bigint,
    alter column yield type bigint using yield::bigint,
    alter column quantity type bigint using quantity::bigint,
    alter column allocation_qty type bigint using allocation_qty::bigint,
    alter column nominal_value type bigint using nominal_value::bigint,
    alter column fx type bigint using fx::bigint;

alter table orders drop column if exists allocation_ptc;

alter table orders add column if not exists allocation_status text not null default 'none';

update orders
set allocation_status = case
    when allocation_qty = 0 then 'none'
    when allocation_qty < quantity then 'partial'
    else 'allocated'
end;

do $$ begin
    alter table orders add constraint orders_alloc_status_check
        check (allocation_status in ('none','partial','allocated'));
exception when duplicate_object then null;
end $$;

do $$ begin
    alter table orders add constraint orders_quantity_check
        check (quantity > 0);
exception when duplicate_object then null;
end $$;

do $$ begin
    alter table orders add constraint orders_alloc_qty_check
        check (allocation_qty >= 0 and allocation_qty <= quantity);
exception when duplicate_object then null;
end $$;

do $$ begin
    alter table orders add constraint orders_price_check
        check (price is null or price > 0);
exception when duplicate_object then null;
end $$;

do $$ begin
    alter table orders add constraint orders_limit_price_check
        check (order_type <> 'limit' or price is not null);
exception when duplicate_object then null;
end $$;

do $$ begin
    alter table orders add constraint orders_alloc_consistency_check check (
        (allocation_status = 'none'      and allocation_qty = 0) or
        (allocation_status = 'partial'   and allocation_qty > 0 and allocation_qty < quantity) or
        (allocation_status = 'allocated' and allocation_qty = quantity)
    );
exception when duplicate_object then null;
end $$;

create index if not exists orders_active_book_idx
    on orders (security_id, direction, price)
    where status = 'active';

drop index if exists orders_trader_id_idx;
drop index if exists orders_security_id_idx;
drop index if exists orders_status_idx;

-- migrate:down
drop index if exists orders_active_book_idx;

create index if not exists orders_trader_id_idx on orders (trader_id);
create index if not exists orders_security_id_idx on orders (security_id);
create index if not exists orders_status_idx on orders (status);

alter table orders
    drop constraint if exists orders_alloc_consistency_check,
    drop constraint if exists orders_limit_price_check,
    drop constraint if exists orders_price_check,
    drop constraint if exists orders_alloc_qty_check,
    drop constraint if exists orders_quantity_check,
    drop constraint if exists orders_alloc_status_check;

alter table orders drop column if exists allocation_status;

alter table orders add column if not exists allocation_ptc numeric;

alter table orders
    alter column price type numeric using price::numeric,
    alter column yield type numeric using yield::numeric,
    alter column quantity type int using quantity::int,
    alter column allocation_qty type int using allocation_qty::int,
    alter column nominal_value type numeric using nominal_value::numeric,
    alter column fx type numeric using fx::numeric;
