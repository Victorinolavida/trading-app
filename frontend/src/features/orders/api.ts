import { ApiClient, RequestOptions } from "../../lib/http";
import { OrderBookEvent } from "./types";

const api = new ApiClient()


export const ordersApi =  {
    getOrderBook: (securityId:string, options: RequestOptions = {}) =>
         api.get<OrderBookEvent>(`/v0/orders/securities/${securityId}/book`,options)
}