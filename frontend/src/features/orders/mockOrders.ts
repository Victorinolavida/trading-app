import type {
  AllocationStatus,
  Order,
  OrderDirection,
  OrderStatus,
  OrderType,
} from "./types";

const SECURITY_IDS = [
  "96c8f91e-3e49-4732-8aa0-01c0589ee62f",
  "0d4b8a2c-7f16-4a55-9e8d-2c7a4f1b6e30",
  "f2a9c6d1-8b34-4e72-a1f0-5d9e3c8b7a21",
  "7c1e5b9a-2d48-4f63-8a17-9e0b6c3d5f42",
  "b5d3f7a0-6c29-4e81-9b4a-1f8d2e7c6a53",
  "3e8a1c4d-9f27-4b65-8d03-6a5c2f9e1b74",
] as const;

const STATUS_PLAN: ReadonlyArray<readonly [OrderStatus, number]> = [
  ["active", 12],
  ["fully_allocated", 5],
  ["cancelled", 5],
  ["expired", 4],
  ["killed", 4],
];

const STATUS_SEQUENCE = STATUS_PLAN.flatMap(([status, count]) =>
  Array.from({ length: count }, () => status),
);

function iso(timestamp: number): string {
  return new Date(timestamp).toISOString();
}

function buildMockOrders(): Order[] {
  const now = Date.now();

  return STATUS_SEQUENCE.map((status, index) => {
    const direction: OrderDirection = index % 2 === 0 ? "bid" : "ask";
    const orderType: OrderType = index % 5 === 0 ? "market" : "limit";
    const quantity = 1_000 * ((index % 6) + 1);
    const price = (95 + (index % 17) * 0.35).toFixed(2);
    const yieldValue = (850 + (index % 9) * 25).toFixed(2);
    const nominalValue = (Number(price) * quantity).toFixed(2);

    // allocation_ptc is a percentage (0-100), not a 0-1 fraction.
    const allocationPtc =
      status === "fully_allocated"
        ? "100"
        : index % 7 === 0
          ? String(10 + (index % 5) * 10)
          : "0";
    const allocationStatus: AllocationStatus =
      allocationPtc === "100"
        ? "allocated"
        : allocationPtc === "0"
          ? "none"
          : "partial";
    const allocatedQuantity = Math.round(
      (quantity * Number(allocationPtc)) / 100,
    );

    const createdAt = now - (index * 7 + 3) * 60_000;
    const updatedAt = createdAt + (index % 4) * 45_000;
    const expiresAt =
      status === "active"
        ? now + (index < 3 ? (index + 1) * 20_000 : (index + 1) * 4 * 60_000)
        : createdAt + 15 * 60_000;

    return {
      id: `2b66245b-0131-4a0e-918b-${String(index).padStart(12, "0")}`,
      security_id: SECURITY_IDS[index % SECURITY_IDS.length],
      direction,
      order_type: orderType,
      price,
      yield: yieldValue,
      quantity,
      nominal_value: nominalValue,
      allocated_quantity: allocatedQuantity,
      allocation_ptc: allocationPtc,
      allocation_status: allocationStatus,
      fx: index % 4 === 0 ? "1.0850" : "1",
      status,
      created_at: iso(createdAt),
      updated_at: iso(updatedAt),
      expires_at: iso(expiresAt),
    };
  });
}

export const mockOrders: Order[] = buildMockOrders();

/**
 * Swappable data source. Replace the mock delay/return with the real API call
 * (or route it through TanStack Query) without touching the blotter UI.
 */
export async function fetchOrders(): Promise<Order[]> {
  await new Promise((resolve) => setTimeout(resolve, 250));
  return mockOrders.map((order) => ({ ...order }));
}
