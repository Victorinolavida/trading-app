import { useQuery } from "@tanstack/react-query";
import { fetchOrders } from "../features/blotter/mockOrders";
import type { OrderDirection, OrderStatus } from "../features/blotter/types";

export function useOrders(
  params: { status?: OrderStatus; direction?: OrderDirection } = {},
) {
  return useQuery({
    queryKey: ["orders", params],
    queryFn: async () => {
      const orders = await fetchOrders();
      return orders.filter(
        (order) =>
          (!params.status || order.status === params.status) &&
          (!params.direction || order.direction === params.direction),
      );
    },
    staleTime: 5_000,
  });
}
