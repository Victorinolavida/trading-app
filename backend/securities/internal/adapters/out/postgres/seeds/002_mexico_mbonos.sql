-- Seed: Mexico Government Bonds (M-Bono) - additional issues
--
-- Additional active securities (type 'M', M-Bono) added to the securities seed
-- so it ends up with 10 active securities: the existing Mexico 7.75% 23-Nov-2034
-- (M-Bono, serie 341123) plus the nine new ones below.
--
-- Source fields -> securities columns (securities/internal/domain/security.go):
--   Name          "Mexico 6.35% 28-Feb-2029" -> name
--   ISIN          MX0MGO0000U4               -> isin (char(12))
--   Issuer        United Mexican States      -> (no column; kept in name/header)
--   Bond type     Government Bond (M-Bono)   -> type = 'M' (domain.MBond)
--   Ticker        M341233                    -> type 'M' + serie '341233'
--   Currency      MXN                        -> currency = 'MXN'
--   Coupon        6.35% fixed, semiannual    -> (no coupon column; kept in name)
--   Maturity date 28 February 2029           -> maturity_date = '2029-02-28'
--   Issue date    19 July 2016               -> issue_date = '2016-07-19'
--   Nominal value 100 MXN                    -> nominal_value = 100, fx = 1
--   Enable                                 -> enable = true
--
-- Idempotent: re-running updates the row matched by UNIQUE (type, serie),
-- so it is safe to run repeatedly (e.g. in CI/local bootstrap).

-- Seed: Mexico 6.35% 28-Feb-2029 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U4',
    'Mexico 6.35% 28-Feb-2029',
    'M', '341233',
    DATE '2016-07-19', DATE '2029-02-28',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 5.75% 31-Dec-2027 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U5',
    'Mexico 5.75% 31-Dec-2027',
    'M', '341234',
    DATE '2012-05-18', DATE '2027-12-31',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 7.00% 15-Jun-2033 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U6',
    'Mexico 7.00% 15-Jun-2033',
    'M', '341235',
    DATE '2008-03-11', DATE '2033-06-15',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 4.90% 31-Dec-2025 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U7',
    'Mexico 4.90% 31-Dec-2025',
    'M', '341236',
    DATE '2011-02-14', DATE '2025-12-31',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 6.75% 20-Nov-2036 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U8',
    'Mexico 6.75% 20-Nov-2036',
    'M', '341237',
    DATE '2013-06-18', DATE '2036-11-20',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 5.50% 28-Feb-2031 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U9',
    'Mexico 5.50% 28-Feb-2031',
    'M', '341238',
    DATE '2017-09-25', DATE '2031-02-28',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 8.00% 31-Dec-2038 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U0',
    'Mexico 8.00% 31-Dec-2038',
    'M', '341239',
    DATE '2015-03-10', DATE '2038-12-31',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 6.50% 15-Aug-2032 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U1',
    'Mexico 6.50% 15-Aug-2032',
    'M', '341240',
    DATE '2017-01-22', DATE '2032-08-15',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();

-- Seed: Mexico 7.25% 30-Jun-2035 (M-Bono)
INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U3',
    'Mexico 7.25% 30-Jun-2035',
    'M', '341241',
    DATE '2018-09-14', DATE '2035-06-30',
    100, 'MXN', 1, true
)
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = true,
    updated_at    = now();
