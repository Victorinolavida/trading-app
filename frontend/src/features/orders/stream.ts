import { skipToken, useQuery, useQueryClient } from "@tanstack/react-query";
import { BackendEvent, EventType, OrderBook } from "./types";
import { useEventSource } from "../../lib/sse";
import { orderKeys } from "./query";

export function useOrdersStream(securityId: string, enabled = true) {
  const qc = useQueryClient();
  const key = orderKeys.events(securityId);

  useEventSource<BackendEvent>({
    url: `/api/v0/orders/securities/${securityId}/book`,
    enabled,
    eventName: "message",
    onMessage: (evt) => {
      switch (evt.type) {
        case EventType.MarketDepth:
          qc.setQueryData<OrderBook>(key, evt.payload as OrderBook);
          break;
        default:
          // unknown shape? fall back to refetch
          qc.invalidateQueries({ queryKey: orderKeys.events(securityId) });
      }
    },
  });

  // read-only view of the cache; the stream is the only writer
  const query = useQuery<OrderBook>({
    queryKey: key,
    queryFn: skipToken, // never fetches
    staleTime: Infinity,
  });

  return { ...query, connection: status };
}
