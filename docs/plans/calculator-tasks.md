# Calculator Service — Task List

Source spec: bond-analytics service for Mexican government bonds.

## Phase 0 — Setup

- [x] Create module skeleton for `calculator` following `@backend/orders/` layout.
- [ ] Add module path / package naming decisions (`backend/calculator`).
- [ ] Wire module into workspace / go.mod replace strategy.

## Phase 1 — Domain & pricing

- [ ] Implement coupon schedule builder from `maturity_date` in `coupon_period_days` steps.
- [ ] Compute accrued interest.
- [ ] Compute dirty price via discounted cash flows.
- [ ] Compute clean price = dirty − accrued.
- [ ] Implement yield from price (Newton-Raphson, bisection fallback).

## Phase 2 — Udibono / UDI handling

- [ ] Multiply UDI-denominated result by UDI value on settlement date.
- [ ] Ignore `securities.fx` for Udibonos; use UDI lookup.

## Phase 3 — Securities boundary

- [ ] Define `SecurityProvider` interface.
- [ ] Fetch security by `public_id`.
- [ ] Read required security fields (ISIN, type, maturity, nominal, currency, coupon_rate, coupon_period_days, day_count_basis).

## Phase 4 — API

- [ ] `POST /v0/analytics/price`
- [ ] `POST /v0/analytics/yield`
- [ ] `POST /v0/analytics/cashflows`
- [ ] `GET /v0/udi-values?date=`
- [ ] `GET /v0/healthz`, `GET /v0/readyz`

## Phase 5 — Data & caching

- [ ] Use the separate `securities_udi_values` table for securities UDI values.
- [ ] Keep UDI values seeded externally for now.
- [ ] Cache securities lookups in memory for a few minutes.

## Phase 6 — Tests

- [ ] Property: price falls as yield rises.
- [ ] Property: `yield(price(y)) ≈ y`.
- [ ] Property: price at maturity equals nominal.

## Phase 7 — Plan / build handoff

- [ ] Final calculator plan created (`docs/plans/calculator.md`).
- [ ] Begin implementation only after plan approval.
