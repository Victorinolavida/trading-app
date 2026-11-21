import { queryOptions } from "@tanstack/react-query";
import { calculatorApi } from "./api";

export const calculatorKeys = {
  all: ["calculator"] as const,
  price: (
    securityId: string,
    settlementDate: string,
    yieldValue: string,
  ) =>
    [
      ...calculatorKeys.all,
      "price",
      securityId,
      settlementDate,
      yieldValue,
    ] as const,
  yield: (
    securityId: string,
    settlementDate: string,
    cleanPrice: string,
  ) =>
    [
      ...calculatorKeys.all,
      "yield",
      securityId,
      settlementDate,
      cleanPrice,
    ] as const,
};

export const calculatorQueries = {
  price: (params: {
    securityId: string;
    settlementDate: string;
    yield: string;
    enabled: boolean
  }) =>
    queryOptions({
      queryKey: calculatorKeys.price(
        params.securityId,
        params.settlementDate,
        params.yield,
      ),
      queryFn: ({ signal }) =>
        calculatorApi.calculatePrice(
          {
            security_id: params.securityId,
            settlement_date: params.settlementDate,
            yield: params.yield,
          },
          { signal },
        ),
      enabled:
      params.enabled &&
        params.securityId.trim().length > 0 && params.yield != null &&
        params.yield.trim().length > 0,
      staleTime: 5 * 60 * 1000,
    }),
  yield: (params: {
    securityId: string;
    settlementDate: string;
    cleanPrice: string;
    enabled: boolean
  }) =>
    queryOptions({
      queryKey: calculatorKeys.yield(
        params.securityId,
        params.settlementDate,
        params.cleanPrice,
      ),
      queryFn: ({ signal }) =>
        calculatorApi.calculateYield(
          {
            security_id: params.securityId,
            settlement_date: params.settlementDate,
            clean_price: params.cleanPrice,
          },
          { signal },
        ),
      enabled:
        params.enabled &&
        params.securityId.trim().length > 0 && params.cleanPrice != null &&
        params.cleanPrice.trim().length > 0,
      staleTime: 5 * 60 * 1000,
    }),
};
