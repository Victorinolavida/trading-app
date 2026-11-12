import { ApiClient, type RequestOptions } from "../../lib/http";
import type { components } from "../../types/generated/openapi";
import {
  ALLOCATION_STATUSES,
  ORDER_DIRECTIONS,
  ORDER_STATUSES,
  ORDER_TYPES,
} from "./constants";
import type {
  AllocationStatus,
  CreateOrderRequest,
  CreateOrderResponse,
  Order,
  OrderBookEvent,
  OrderDirection,
  OrderStatus,
  OrderType,
} from "./types";

type CancelOrderResponse = components["schemas"]["OrderResponse"];

// The gateway serializes the proto expires_at field, but the checked-in
// OpenAPI document predates that field. Extend the generated shape locally
// until the spec is regenerated.
type ApiOrder = components["schemas"]["OrderResponse"] & {
  expires_at?: string;
};

type ListOrdersResponse = Omit<
  components["schemas"]["ListOrdersResponse"],
  "orders"
> & {
  orders: ApiOrder[];
};

const api = new ApiClient();

function isOneOf<T extends string>(
  value: string | undefined,
  options: readonly T[],
): value is T {
  return value !== undefined && options.some((option) => option === value);
}

function readDirection(value: string | undefined): OrderDirection {
  return isOneOf(value, ORDER_DIRECTIONS) ? value : "bid";
}

function readOrderType(value: string | undefined): OrderType {
  return isOneOf(value, ORDER_TYPES) ? value : "limit";
}

function readStatus(value: string | undefined): OrderStatus {
  return isOneOf(value, ORDER_STATUSES) ? value : "killed";
}

function readAllocationStatus(value: string | undefined): AllocationStatus {
  return isOneOf(value, ALLOCATION_STATUSES) ? value : "none";
}

function toOrder(apiOrder: ApiOrder): Order {
  const createdAt = apiOrder.created_at ?? "";

  return {
    id: apiOrder.id ?? "",
    security_id: apiOrder.security_id ?? "",
    direction: readDirection(apiOrder.direction),
    order_type: readOrderType(apiOrder.order_type),
    price: apiOrder.price ?? "0",
    yield: apiOrder.yield ?? "0",
    quantity: apiOrder.quantity ?? 0,
    nominal_value: apiOrder.nominal_value ?? "0",
    allocated_quantity: apiOrder.allocated_quantity ?? 0,
    allocation_ptc: apiOrder.allocation_ptc ?? "0",
    allocation_status: readAllocationStatus(apiOrder.allocation_status),
    fx: apiOrder.fx ?? "1",
    status: readStatus(apiOrder.status),
    created_at: createdAt,
    updated_at: apiOrder.updated_at ?? createdAt,
    expires_at: apiOrder.expires_at ?? "",
  };
}

export const ordersApi = {
  getOrderBook: (securityId: string, options: RequestOptions = {}) =>
    api.get<OrderBookEvent>(
      `/v0/orders/securities/${securityId}/book`,
      options,
    ),
  listOrders: async (options: RequestOptions = {}): Promise<Order[]> => {
    const response = await api.get<ListOrdersResponse>("/v0/orders", options);
    return response.orders.map(toOrder);
  },
  getOrder: async (
    orderId: string,
    options: RequestOptions = {},
  ): Promise<Order | null> => {
    const response = await api.get<ListOrdersResponse>("/v0/orders", {
      ...options,
      params: { order_ids: orderId },
    });
    return response.orders.map(toOrder)[0] ?? null;
  },
  createOrder: (
    orderRequest: CreateOrderRequest,
    options: RequestOptions = {},
  ) => api.post<CreateOrderResponse>("/v0/orders", orderRequest, options),
  cancelOrder: (orderId: string, options: RequestOptions = {}) =>
    api.patch<CancelOrderResponse>(
      `/v0/orders/${encodeURIComponent(orderId)}/cancel`,
      undefined,
      options,
    ),
};
