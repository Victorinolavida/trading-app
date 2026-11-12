import { ApiClient, RequestOptions } from "../../lib/http";
import {
  CreateOrderRequest,
  CreateOrderResponse,
  OrderBookEvent,
} from "./types";

const api = new ApiClient();

export const ordersApi = {
  getOrderBook: (securityId: string, options: RequestOptions = {}) =>
    api.get<OrderBookEvent>(
      `/v0/orders/securities/${securityId}/book`,
      options,
    ),
  createOrder: (
    orderRequest: CreateOrderRequest,
    options: RequestOptions = {},
  ) => api.post<CreateOrderResponse>("/v0/orders", orderRequest, options),
};
