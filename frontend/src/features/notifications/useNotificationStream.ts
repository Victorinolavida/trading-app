import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useLocation } from "react-router";
import { notificationKeys, notificationsApi } from "./api";

// SSE is only a signal. Fetching after every connection recovers missed events.
export function useNotificationStream() {
  const queryClient = useQueryClient();
  const { pathname } = useLocation();
  const inboxOpen = useRef(pathname === "/notifications");
  inboxOpen.current = pathname === "/notifications";

  useEffect(() => {
    let closed = false;
    let pending = Promise.resolve();

    const refresh = () => {
      void queryClient.invalidateQueries({ queryKey: notificationKeys.all });
    };
    const sync = () => {
      if (!inboxOpen.current) {
        refresh();
        return;
      }
      // Serialize read-all calls so a signal received during a request gets
      // another pass after it completes.
      pending = pending
        .then(async () => {
          if (closed || !inboxOpen.current) return;
          try {
            await notificationsApi.markAllRead();
          } finally {
            if (!closed) refresh();
          }
        })
        .catch(() => {
          // The next signal or reconnect retries; the inbox can also retry.
        });
    };

    const source = new EventSource("/api/v0/notifications/stream", {
      withCredentials: true,
    });
    source.addEventListener("open", sync);
    source.addEventListener("notifications.changed", sync);

    return () => {
      closed = true;
      source.removeEventListener("open", sync);
      source.removeEventListener("notifications.changed", sync);
      source.close();
    };
  }, [queryClient]);
}
