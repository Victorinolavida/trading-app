import type {
  ALLOCATION_STATUSES,
  ORDER_DIRECTIONS,
  ORDER_STATUSES,
  ORDER_TYPES,
} from "./constants";

export interface BackendEvent<T = unknown> {
  type: EventType;
  security_id: string;
  sequence: number;
  payload: T;
}

export interface OrderBook {
  bids: Side[];
  asks: Side[];
}

export interface Side {
  price: string;
  nominal_value: string;
  quantity: number;
  order_counts: number;
  direction: string;
}

export enum EventType {
  MarketDepth = "market.depth",
  OrderExecuted = "orders.executed",
}

export type OrderBookEvent = BackendEvent<OrderBook>;

export type OrderDirection = (typeof ORDER_DIRECTIONS)[number];
export type OrderType = (typeof ORDER_TYPES)[number];
export type OrderStatus = (typeof ORDER_STATUSES)[number];
export type AllocationStatus = (typeof ALLOCATION_STATUSES)[number];

export interface Order {
  id: string;
  security_id: string;
  direction: OrderDirection;
  order_type: OrderType;
  price: string;
  yield: string;
  quantity: number;
  nominal_value: string;
  allocated_quantity: number;
  allocation_ptc: string;
  allocation_status: AllocationStatus;
  fx: string;
  status: OrderStatus;
  created_at: string;
  updated_at: string;
  expires_at: string;
}

export interface CreateOrderRequest {
  security_id: string;
  direction: "bid" | "ask";
  order_type: "limit" | "market";
  price?: string;
  yield?: string;
  quantity: number;
  expiration?: string;
}

export interface CreateOrderResponse {
  id: string;
  security_id: string;
  direction: string;
  order_type: string;
  price: string;
  yield: string;
  quantity: number;
  nominal_value: string;
  allocated_quantity: number;
  allocation_ptc: string;
  allocation_status: string;
  fx: string;
  status: string;
  created_at: string;
  updated_at: string;
}
