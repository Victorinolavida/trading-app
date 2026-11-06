-- Seed: United Mexican States 7.75% 2034 (M-Bono)
--
-- Source fields -> securities columns (securities/internal/domain/security.go):
--   Name          "Mexico 7.75% 23-Nov-2034" -> name
--   ISIN          MX0MGO0000U2               -> isin (char(12))
--   Issuer        United Mexican States      -> (no column; kept in name/header)
--   Bond type     Government Bond (M-Bono)   -> type = 'M' (domain.MBond)
--   Ticker        M341123                    -> type 'M' + serie '341123'
--   Currency      MXN                        -> currency = 'MXN'
--   Coupon        7.75% fixed, semiannual    -> (no coupon column; kept in name)
--   Maturity date 23 November 2034           -> maturity_date = '2034-11-23'
--   Issue date    19 December 2013           -> issue_date = '2013-12-19'
--   Nominal value 100 MXN                    -> nominal_value = 100, fx = 1
--   Enable                                 -> enable = true
--
-- Idempotent: re-running updates the row matched by UNIQUE (type, serie),
-- so it is safe to run repeatedly (e.g. in CI/local bootstrap).

INSERT INTO securities (
    id, public_id, isin, name, type, serie,
    issue_date, maturity_date, nominal_value, currency, fx, enable
) VALUES (
    gen_random_uuid(), gen_random_uuid(),
    'MX0MGO0000U2',
    'Mexico 7.75% 23-Nov-2034',
    'M', '341123',
    DATE '2013-12-19', DATE '2034-11-23',
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
