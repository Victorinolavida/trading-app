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

export interface CreateOrderRequest {
  security_id: string;
  direction: "bid" | "ask";
  order_type: "limit" | "market";
  price?: string;
  yield?: string;
  quantity: number;
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
