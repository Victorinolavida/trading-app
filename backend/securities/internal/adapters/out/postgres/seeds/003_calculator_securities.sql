-- securities_seed.sql
-- 40 rows: 14 real (ISIN + maturity from Cbonds) + 26 synthetic test rows.
--
-- Idempotent: rows are matched by UNIQUE (type, serie) and updated in place.
--
-- VERIFY BEFORE USING FOR REAL:
--   * issue_date on the real rows is APPROXIMATE (I could not find it in the sources).
--   * UDI rows use the latest published quote in udi_values on or before today.
--     They are skipped until an eligible quote is loaded.
--   * Real rows with enable = false have already matured.
--
-- Synthetic rows use fake ISINs prefixed MXTSTM / MXTSTS.
-- Delete with: DELETE FROM securities WHERE isin LIKE 'MXTST%';

BEGIN;

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM udi_values WHERE date <= CURRENT_DATE) THEN
        RAISE NOTICE 'Skipping Udibono seeds: load a UDI quote dated today or earlier and rerun the seed';
    END IF;
END;
$$;

-- ---------------------------------------------------------------------------
-- Real M-bonos (type 'M', MXN, nominal 100)
-- ---------------------------------------------------------------------------
INSERT INTO securities
    (id, public_id, isin, type, serie, issue_date, maturity_date, nominal_value, currency, fx, enable, name)
VALUES
    (gen_random_uuid(), gen_random_uuid(), 'MX0MGO0000B2', 'M', '361120', '2006-11-23', '2036-11-20', 100, 'MXN', 1, true,  'MBONO 10 11/20/36 M'),
    (gen_random_uuid(), gen_random_uuid(), 'MX0MGO0001L9', 'M', '360221', '2026-02-26', '2036-02-21', 100, 'MXN', 1, true,  'MBONO 8 02/21/36 M'),
    (gen_random_uuid(), gen_random_uuid(), 'MX0MGO0000D8', 'M', '270603', '2017-06-08', '2027-06-03', 100, 'MXN', 1, true,  'MBONO 7.5 06/03/27 M'),
    (gen_random_uuid(), gen_random_uuid(), 'MX0MGO0001O3', 'M', '320415', '2022-04-21', '2032-04-15', 100, 'MXN', 1, true,  'MBONO 8 04/15/32 M'),
    (gen_random_uuid(), gen_random_uuid(), 'MX0MGO000193', 'M', '260903', '2021-09-09', '2026-09-03', 100, 'MXN', 1, false, 'MBONO 7 09/03/26 M')
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = EXCLUDED.enable,
    updated_at    = now();

-- ---------------------------------------------------------------------------
-- Real Udibonos (type 'S', UDI, nominal 100 UDI, fx = UDI value in MXN)
-- ---------------------------------------------------------------------------
INSERT INTO securities
    (id, public_id, isin, type, serie, issue_date, maturity_date, nominal_value, currency, fx, enable, name)
SELECT gen_random_uuid(), gen_random_uuid(), bond.isin, 'S', bond.serie,
       bond.issue_date::date, bond.maturity_date::date, 100, 'UDI', quote.value,
       bond.enable, bond.name
FROM (
    VALUES
        ('MX0SGO000023', '351122', '2005-11-24', '2035-11-22', true,  'UDI 4.5 11/22/35 S'),
        ('MX0SGO0000K0', '311127', '2021-12-02', '2031-11-27', true,  'UDI 2.75 11/27/31 S'),
        ('MX0SGO0000E3', '461108', '2016-11-10', '2046-11-08', true,  'UDI 4 11/08/46 S'),
        ('MX0SGO0000N4', '431112', '2013-11-14', '2043-11-12', true,  'UDI 6.5 11/12/43 S'),
        ('MX0SGO0000V7', '470801', '2017-08-03', '2047-08-01', true,  'UDI 4 08/01/47 S'),
        ('MX0SGO0000Q7', '541029', '2024-10-31', '2054-10-29', true,  'UDI 4 10/29/54 S'),
        ('MX0SGO0000F0', '281130', '2018-12-06', '2028-11-30', true,  'UDI 4 11/30/28 S'),
        ('MX0SGO000098', '401115', '2010-11-18', '2040-11-15', true,  'UDI 4 11/15/40 S'),
        ('MX0SGO000015', '251204', '2015-12-10', '2025-12-04', false, 'UDI 4.5 12/04/25 S')
) AS bond(isin, serie, issue_date, maturity_date, enable, name)
CROSS JOIN (
    SELECT value FROM udi_values WHERE date <= CURRENT_DATE ORDER BY date DESC LIMIT 1
) AS quote
WHERE true
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = EXCLUDED.enable,
    updated_at    = now();

-- ---------------------------------------------------------------------------
-- SYNTHETIC test data (fake ISINs prefixed MXTSTM / MXTSTS). 13 M + 13 S.
-- Delete with: DELETE FROM securities WHERE isin LIKE 'MXTST%';
-- ---------------------------------------------------------------------------
INSERT INTO securities
    (id, public_id, isin, type, serie, issue_date, maturity_date, nominal_value, currency, fx, enable, name)
SELECT gen_random_uuid(), gen_random_uuid(),
       'MXTSTM' || lpad(g::text, 6, '0'),
       'M',
       to_char(mat, 'YYMMDD'),
       (mat - interval '10 years')::date,
       mat,
       100, 'MXN', 1, true,
       'MBONO ' || to_char(mat, 'MM/DD/YY')
FROM (
    SELECT g, (date '2027-03-04' + g * interval '1 year')::date AS mat
    FROM generate_series(1, 13) AS g
) t
WHERE true
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = EXCLUDED.enable,
    updated_at    = now();

INSERT INTO securities
    (id, public_id, isin, type, serie, issue_date, maturity_date, nominal_value, currency, fx, enable, name)
SELECT gen_random_uuid(), gen_random_uuid(),
       'MXTSTS' || lpad(g::text, 6, '0'),
       'S',
       to_char(mat, 'YYMMDD'),
       (mat - interval '10 years')::date,
       mat,
       100, 'UDI', quote.value, true,
       'UDI ' || to_char(mat, 'MM/DD/YY')
FROM (
    SELECT g, (date '2027-06-10' + g * interval '1 year')::date AS mat
    FROM generate_series(1, 13) AS g
) t
CROSS JOIN (
    SELECT value FROM udi_values WHERE date <= CURRENT_DATE ORDER BY date DESC LIMIT 1
) AS quote
WHERE true
ON CONFLICT (type, serie) DO UPDATE SET
    isin          = EXCLUDED.isin,
    name          = EXCLUDED.name,
    issue_date    = EXCLUDED.issue_date,
    maturity_date = EXCLUDED.maturity_date,
    nominal_value = EXCLUDED.nominal_value,
    currency      = EXCLUDED.currency,
    fx            = EXCLUDED.fx,
    enable        = EXCLUDED.enable,
    updated_at    = now();

COMMIT;

-- Sanity check
-- SELECT type, currency, count(*) FROM securities GROUP BY 1, 2;  -- expect M/MXN 18, S/UDI 22
