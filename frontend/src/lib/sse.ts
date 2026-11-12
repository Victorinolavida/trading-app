import { useEffect, useRef } from "react";

type Options<T> = {
  url: string;
  enabled?: boolean;
  onMessage: (data: T) => void;
  eventName?: string; // named SSE event, defaults to "message"
};

export function useEventSource<T>({
  url,
  enabled = true,
  onMessage,
  eventName = "",
}: Options<T>) {
  // keep latest callback without reconnecting on every render
  const handlerRef = useRef(onMessage);
  handlerRef.current = onMessage;

  useEffect(() => {
    if (!enabled) return;

    const es = new EventSource(url, { withCredentials: true });

    const listener = (e: MessageEvent) => {
      try {
        handlerRef.current(JSON.parse(e.data) as T);
      } catch (err) {
        console.error("SSE parse error", err);
      }
    };

    es.addEventListener(eventName, listener);
    es.onerror = () => {
      // EventSource auto-reconnects; only handle fatal cases here
      if (es.readyState === EventSource.CLOSED) {
        console.warn("SSE closed");
      }
    };

    return () => {
      es.removeEventListener(eventName, listener);
      es.close();
    };
  }, [url, enabled, eventName]);
}
