import { queryOptions } from "@tanstack/react-query";
import { ordersApi } from "./api";

export const orderKeys = {
  all: ["orders"] as const,
  events: (securityID: string) =>
    [...orderKeys.all, "order-book", securityID] as const,
  list: () => [...orderKeys.all, "list"] as const,
  detail: (orderID: string) =>
    [...orderKeys.all, "detail", orderID] as const,
};

export const OrderQueries = {
  orderBook: (securityID: string) =>
    queryOptions({
      queryKey: orderKeys.events(securityID),
      queryFn: async ({ signal }) =>
        ordersApi.getOrderBook(securityID, { signal }),
      enabled: securityID.trim().length > 0,
    }),
  list: () =>
    queryOptions({
      queryKey: orderKeys.list(),
      queryFn: async ({ signal }) => ordersApi.listOrders({ signal }),
      staleTime: 5_000,
      refetchInterval: 5_000,
    }),
  detail: (orderID: string) =>
    queryOptions({
      queryKey: orderKeys.detail(orderID),
      queryFn: async ({ signal }) => ordersApi.getOrder(orderID, { signal }),
      enabled: orderID.trim().length > 0,
      retry: false,
      staleTime: 5_000,
    }),
};
