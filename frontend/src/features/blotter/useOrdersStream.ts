import { useEffect, useRef } from "react";
import type { Order } from "./types";

export interface OrdersStreamHandlers {
  onUpsert: (order: Order) => void;
  onSnapshot?: (orders: Order[]) => void;
}

export type OrdersStreamStatus = "idle" | "connecting" | "open" | "closed";

export interface OrdersStreamState {
  status: OrdersStreamStatus;
}

/**
 * Extension point for real-time order updates.
 *
 * TODO: open a WebSocket or EventSource here and call
 * `onUpsertRef.current(message.order)` for every create/update event (upsert
 * by `id`), or `onSnapshotRef.current(orders)` on a full snapshot. Keep the
 * handlers in refs so a future socket effect does not reconnect on every
 * parent render.
 */
export function useOrdersStream(
  handlers: OrdersStreamHandlers,
): OrdersStreamState {
  const onUpsertRef = useRef(handlers.onUpsert);
  const onSnapshotRef = useRef(handlers.onSnapshot);

  useEffect(() => {
    onUpsertRef.current = handlers.onUpsert;
    onSnapshotRef.current = handlers.onSnapshot;
  }, [handlers.onUpsert, handlers.onSnapshot]);

  return { status: "idle" };
}
