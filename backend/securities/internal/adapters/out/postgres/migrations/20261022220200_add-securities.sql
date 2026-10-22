-- migrate:up
create table securities (
    id            uuid          primary key, -- v7
    public_id     uuid          not null unique, -- v4
    isin          char(12)      not null,
    type    varchar(4)    not null check (type in ('M', 'S')),
    serie         varchar(16)   not null,
    issue_date    date          not null,
    maturity_date date          not null,
    nominal_value numeric(10,6) not null,
    currency      varchar(3)    not null check (currency in ('MXN', 'UDI')),
    fx            numeric(10,6) not null default 1,
    created_at    timestamptz   not null default now(),
    updated_at    timestamptz,
    unique (type, serie)
);

-- migrate:down
drop table securities;
