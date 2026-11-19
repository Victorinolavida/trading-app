# Calculator — Implementation Plan

Source spec: bond-analytics service for Mexican government bonds.

## Goal

Compute analytics for Mexican government bonds and Udibonos.

Given a security, a settlement date and a yield, return clean price, accrued interest and dirty price.

Can also solve for yield from price, return coupon schedule, and convert Udibonos to MXN using the UDI value for the settlement date.

Owns no securities data. Fetches security details from the existing securities service.

Non-goals for v0: CRUD for securities, live market prices, auction results, portfolio tracking, authentication, Cetes and Bondes.

## Boundary with the securities service

- Gets security by public_id.
- Reads ISIN, type, maturity date, nominal value, currency, coupon_rate, coupon_period_days and day_count_basis.
- Never reads the securities DB directly, so services stay independently deployable.
- Coupon fields must exist in the securities service; if not, that is the one change on that side.

## Data owned by this service

```sql
CREATE TABLE securities_udi_values (
    date       date PRIMARY KEY,
    value      numeric(12,6) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
```

- The calculator owns no securities data.

- The securities service keeps a separate UDI values table: `securities_udi_values`.
- UDI values are seeded externally for now.

- `securities.fx` is ignored for Udibonos.
- UDI is looked up by date instead.

## API (v0)

- `POST /v0/analytics/price` — `{public_id, settlement_date, yield}` → clean price, accrued interest, dirty price, and MXN value for UDI securities.
- `POST /v0/analytics/yield` — `{public_id, settlement_date, clean_price}` → yield.
- `POST /v0/analytics/cashflows` — coupon schedule with dates and amounts.
- `GET /v0/udi-values?date=` — UDI value, falling back to the latest earlier date.
- `GET /v0/healthz`, `GET /v0/readyz` — probes (readiness checks the DB).

## Pricing logic

1. Build coupon schedule backward from `maturity_date` in `coupon_period_days` steps (182 for these bonds).
2. Accrued interest = nominal × coupon × days_since_last_coupon / day_count_basis.
3. Dirty price = sum of discounted cash flows at `1 + y × period/basis` per period, with fractional exponent for first stub.
4. Clean price = dirty price − accrued interest.
5. Yield from price: Newton-Raphson, with bisection as fallback.
6. For Udibonos, result is in UDI; multiply by UDI value on settlement date for MXN.

## Testing

- Property tests: price falls as yield rises, yield(price(y)) ≈ y, and price at maturity equals nominal.

## Decisions

- Module lives inside the `backend/` folder.
- Runs with the other backend services in the same process.
- Securities client is HTTP only for now; no gRPC.
- `securities_udi_values` is the securities UDI values table.
- UDI values are seeded externally for now.
- Coupon fields still need the `ALTER TABLE`.

## Task list

See `docs/plans/calculator-tasks.md`.
