import { queryOptions } from "@tanstack/react-query";
import { blotterApi } from "./api";

export const blotterKeys = {
  all: ["orders", "blotter"] as const,
  list: () => [...blotterKeys.all, "list"] as const,
};

export const blotterQueries = {
  list: () =>
    queryOptions({
      queryKey: blotterKeys.list(),
      queryFn: async ({ signal }) => blotterApi.listOrders({ signal }),
      staleTime: 5_000,
      refetchInterval: 5_000,
    }),
};
