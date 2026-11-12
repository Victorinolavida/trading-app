import type {
  ALLOCATION_STATUSES,
  ORDER_DIRECTIONS,
  ORDER_STATUSES,
  ORDER_TYPES,
} from "./constants";

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
