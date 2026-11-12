
export interface BackendEvent<T = unknown> {
  type: EventType
  security_id: string
  sequence: number
  payload: T 
}

export interface OrderBook {
  bids: Side[]
  asks: Side[]
}

export interface Side {
  price: string
  nominal_value: string
  quantity: number
  order_counts: number
  direction: string
}

export enum EventType {
  MarketDepth = "market.depth",
  OrderExecuted = "orders.executed"
}

export type OrderBookEvent = BackendEvent<OrderBook>

