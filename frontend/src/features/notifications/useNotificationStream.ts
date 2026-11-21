import { useEffect, useRef } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { useLocation } from "react-router";
import { toast } from "sonner";
import { notificationKeys, notificationsApi } from "./api";
import { notificationTitles } from "./presentation";

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
    const onNotification = (event: MessageEvent<string>) => {
      try {
        const data: unknown = JSON.parse(event.data);
        if (typeof data === "object" && data !== null && "event_type" in data && typeof data.event_type === "string" && "id" in data && typeof data.id === "string" && data.id) {
          toast(notificationTitles[data.event_type] ?? "Order notification", {
            id: `notification-${data.id}`,
          });
        }
      } catch {
        // A malformed live signal still triggers a refresh of saved notifications.
      }
      sync();
    };
    source.addEventListener("open", sync);
    source.addEventListener("notifications.changed", onNotification);

    return () => {
      closed = true;
      source.removeEventListener("open", sync);
      source.removeEventListener("notifications.changed", onNotification);
      source.close();
    };
  }, [queryClient]);
}
