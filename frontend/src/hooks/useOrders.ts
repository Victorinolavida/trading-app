import { useQuery } from "@tanstack/react-query";
import { getOrders, type OrderStatus } from "../lib/api";

export function useOrders(
  params: { status?: OrderStatus; direction?: "bid" | "ask" } = {},
) {
  return useQuery({
    queryKey: ["orders", params],
    queryFn: () => getOrders(params),
    staleTime: 5_000,
    refetchInterval: 5_000,
  });
}
