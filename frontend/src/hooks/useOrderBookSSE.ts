import { useEffect } from "react";
import { OrderBookUpdate, ExecutedTrade, subscribeOrderBook } from "../lib/api";

export function useOrderBookSSE(
  securityId: string,
  onUpdate: (update: OrderBookUpdate) => void,
  onExecuted: (trade: ExecutedTrade) => void,
) {
  useEffect(() => {
    if (!securityId) return;

    const source = subscribeOrderBook(securityId, onUpdate, onExecuted);

    // Handle reconnection on error
    source.onerror = () => {
      setTimeout(() => {
        // Reconnect after a delay
        subscribeOrderBook(securityId, onUpdate, onExecuted);
      }, 3000);
    };

    // Cleanup on unmount
    return () => {
      source.close();
    };
  }, [securityId, onUpdate, onExecuted]);
}
