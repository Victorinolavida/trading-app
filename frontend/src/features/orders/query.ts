import { queryOptions } from "@tanstack/react-query";
import { ordersApi } from "./api";

export const orderKeys = {
    all: ["orders"] as const,
    events: (securityID:string) =>  [...orderKeys.all, "order-book", securityID] as const
}


export const OrderQueries = {
    orderBook: (securityID: string) => queryOptions({
        queryKey: orderKeys.events(securityID),
      queryFn: async ({ signal }) =>  ordersApi.getOrderBook(securityID, { signal }),
      enabled: securityID.trim().length > 0,

    })
};